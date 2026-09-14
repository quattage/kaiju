/******************************************************************************/
/* area.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_areas

import (
	"fmt"
	"log/slog"
	"regexp"

	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
)

type AreaHandler interface {
	Open(area *Area, editor EditorAreaInterface)
	Close(area *Area, editor EditorAreaInterface)
	FocusInterface(editor EditorAreaInterface)
	BlurInterface(editor EditorAreaInterface)
	Update(area *Area, editor EditorAreaInterface, deltaTime float64, posx, posy, width, height float32)
	GetDisplacement(ed EditorAreaInterface, posx, posy, width, height float32) (float32, float32, float32, float32)
	IsFocusedOnInput() bool
}

type SplitDirection uint8

const (
	SplitHorizontal SplitDirection = iota
	SplitVertical
)

type AreaSubtype uint8

const (
	// Describes Areas that can be manipulated directly by the user
	// with no restrictions. Most Areas will use this, such
	// as the Stage Editor, RenderGraph, and Engine Perferences.
	SubtypeDockable AreaSubtype = iota
	// Restricted accessibility indicates that the AreaType isn't surfaced
	// to the user, but can still be freely moved and docked should one be
	// created elsewhere by editor or plugin code.
	// This is primarily intended to be used by the AreaTypeComposite type.
	SubtypeRestricted
	// OverlayOnly describes Areas that can't be docked or created manually.
	// Areas that use OverlayOnly are typically context-dependent
	// sub-windows with ephemeral state, such as individual context
	// menus, the color wheel, confirmation prompts, and the action pallette
	SubtypeOverlayOnly
	// Internal indicates that this Area cannot be moved, accessed, docked, or
	// created. It is only managable by editor or plugin code.
	SubtypeInternal
)

// The AreaType is a registry construct that exists to convey implementation
// specifics to the Area itself. Each discrete Area type
// (such as the Stage viewport, shader editor, render graph, etc.)
// will exist as an AreaType.
type AreaType struct {
	ID      string
	Name    string
	Handler AreaHandler
	// this should be a bitflag enum to describe capabilities
	Subtype AreaSubtype
}

var deferredAreaTypeRegistry = []func() AreaType{}
var regex = regexp.MustCompile(`^[a-z_.]+$`)

// Register allows packages to append their own AreaType factories during
// static initialization.
func Register(areaType func() AreaType) {
	if deferredAreaTypeRegistry == nil {
		slog.Error("Skipped attempt to register an AreaType factory after the registry has closed.")
		return
	}
	deferredAreaTypeRegistry = append(deferredAreaTypeRegistry, areaType)
}

// AreaTypeComposite is a dedicated, default, registry-specific
// AreaType that exists solely to represent Areas that defer their
// functionality to their children. The AreaTypeComposite's handler
// is, by default, nil. The hierarchy of Areas is walked recursively
// The handler here could *potentially* be assigned by a plugin to
// essentially override the recursive Area walking operations performed
// by the functions at the bottom of this file.
// Should such a thing be desirable for whatever reason, this would
// theoretically allow plugins to fundamentally alter the layout behaviour
// of Areas.
var AreaTypeComposite = AreaType{
	ID:      "com.kaiju.area_composite",
	Name:    "Composite",
	Handler: nil,
	Subtype: SubtypeRestricted,
}

type Area struct {
	Type           AreaType
	Manager        *ui.Manager
	Root           *ui.Panel
	Parent         *Area
	ChildA         *Area
	ChildB         *Area
	SplitDirection SplitDirection
	Ratio          float32
	IsOverlay      bool
}

// Called by the workspace manager and serializer to set up this Area's
// manager and panel
func (a *Area) open(wm *WorkspaceManager) {
	a.openWithPreSize(wm, 1, 1)
}

func (a *Area) openWithPreSize(wm *WorkspaceManager, width, height int) {
	a.Manager = &ui.Manager{}
	a.Manager.Init(wm.editor.Host())
	a.Root = a.Manager.Add().ToPanel()
	a.Root.Init(nil, ui.ElementTypePanel)
	a.SetBackdropColor(wm.editor.Theme().PanelColor.AsColor())
	a.Root.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	a.Root.Base().Layout().SetOffset(0, 0)
	a.Root.Base().Layout().Scale(float32(width), float32(height))
	a.Root.DontFitContent()
	a.Root.SetFlex()
	a.Root.SetFlexDirection(ui.FlexDirectionColumn)
	a.Root.SetFlexAlignItems(ui.FlexAlignStretch)
	a.Root.SetFlexJustify(ui.FlexJustifySpaceBetween)
	a.Type.Handler.Open(a, wm.editor)
	slog.Debug(fmt.Sprintf("Opened '%s'", a.Type.ID))
}

func (a *Area) close(wm *WorkspaceManager) {
	a.Type.Handler.Close(a, wm.editor)
	a.Manager.Shutdown()
	a.Type = AreaType{
		ID:      "NIL_CLOSED",
		Name:    "NIL_CLOSED",
		Handler: nil,
		Subtype: SubtypeRestricted,
	}
	a.Manager = nil
	a.Root = nil
	a.Parent = nil
	a.ChildA = nil
	a.ChildB = nil
	a.SplitDirection = SplitHorizontal
	a.Ratio = -1
	a.IsOverlay = false
}

func (a *Area) SetHighlighted(highlighted bool) {
	a.assertInitialized()
	a.Root.SetOutline(2, 0, matrix.ColorRed())
}

func (a *Area) SetBackdropColor(color matrix.Color) {
	a.assertInitialized()
	a.Root.SetColor(color)
}

func (a *Area) assertInitialized() {
	if a.Root == nil || a.Manager == nil {
		panic("UI audit on uninitialized area")
	}
}

// "Composite" Areas don't have any UI of their own and instead
// defer all their functionality to their children.
func (a *Area) IsComposite() bool {
	return a.Type.ID == AreaTypeComposite.ID
}

// IsDockable returns true if this Area is allowed to
// be parented to another Area.
func (a *Area) IsDockable() bool {
	return a.Type.Subtype != SubtypeOverlayOnly
}

func (a *Area) Width() float32 {
	if a.Root == nil {
		return 0
	}
	return a.Root.Base().Entity().Transform.Scale().X()
}

func (a *Area) Height() float32 {
	if a.Root == nil {
		return 0
	}
	return a.Root.Base().Entity().Transform.Scale().Y()
}

func (a *Area) TakedownLayout() {
	a.Manager.Shutdown()
	a.Root = nil
}

// PerformAsChildren selectively recurses into child Areas and runs the
// supplied function, provided the Area either has its own children or a
// valid handler.
func (a *Area) PerformAsChildren(operation func(AreaHandler), editor EditorAreaInterface) {
	if a.Type.Handler != nil {
		operation(a.Type.Handler)
	}
	if a.ChildA != nil && a.ChildB != nil {
		if a.ChildA == a || a.ChildB == a {
			panic("Cycle in Area hierarchy")
		}
		a.ChildA.PerformAsChildren(operation, editor)
		a.ChildB.PerformAsChildren(operation, editor)
	}
}

func (a *Area) Update(editor EditorAreaInterface, deltaTime float64, posx, posy, width, height float32) {
	if a.Root != nil {
		a.Root.Base().Layout().SetOffset(posx, posy)
		a.Root.Base().Layout().Scale(width, height)
	}
	if a.Type.Handler != nil {
		a.Type.Handler.Update(a, editor, deltaTime, posx, posy, width, height)
		posx, posy, width, height = a.Type.Handler.GetDisplacement(editor, posx, posy, width, height)
	}
	if a.ChildA != nil && a.ChildB != nil {
		if a.ChildA == a || a.ChildB == a {
			panic("Cycle in Area hierarchy")
		}
		switch a.SplitDirection {
		case SplitHorizontal:
			split := width * a.Ratio
			a.ChildA.Update(editor, deltaTime, posx, posy, split, height)
			a.ChildB.Update(editor, deltaTime, split, posy, width-split, height)
		case SplitVertical:
			split := height * a.Ratio
			a.ChildA.Update(editor, deltaTime, posx, posy, width, height-split)
			a.ChildB.Update(editor, deltaTime, posx, split, width, height-split)
		}
	}
}
