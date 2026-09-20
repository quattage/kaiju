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
	// GetOffsets returns relative pixel measurements [posx, posy, width, height]
	// to augment the recursive Area tiling layout. Use this to define a boundary
	// for an Area (such as its top/bottom bars) that the Area tiler won't draw over.
	// Note that these are raw pixel measurements, so you'll need to do the conversions
	// yourself to ensure consistency across displays of different pixel densities.
	GetOffsets(ed EditorAreaInterface) (float32, float32, float32, float32)
	// GetOverlayDimensions returns the size [width, height] in virtualized milimeters
	// that the Area content set to by default when it is open as an overlay.
	// Units are normalized against DPMM and convertex to pixels internally for
	// consistency across displays
	GetOverlayDimensions() (float32, float32)
	IsFocusedOnInput() bool
}

type SplitDirection uint8

const (
	SplitHorizontal SplitDirection = iota
	SplitVertical
)

type AreaCapabilityFlags = int

const (
	// Dockable AreaTypes can be dropped into the primary Area hierarchy
	// and tiled by the workspcae manager. Most Areas should be dockable, but
	// some may not desire this. (For example, the splash screen isn't dockable)
	AreaCapabilityFlagDockable = AreaCapabilityFlags(1 << iota)
	// Overlayable AreaTypes are permitted to float above the layout.
	// Various context menus including the color picker are overlayable.
	AreaCapabilityFlagOverlayable = AreaCapabilityFlags(1 << iota)
	// Discoverble AreaTypes are surfaced to the game developer in the editor.
	// The AreaType Selector only includes discoverable areas. If an AreaType
	// isn't marked as discoverable, the only way it may be opened is via
	// editor or plugin code.
	AreaCapabilityFlagDiscoverable = AreaCapabilityFlags(1 << iota)
	// For standard areas, Draggable indicates that areas using this
	// AreaType can have their bounds adjusted by dragging with the mouse.
	// For Overlays, this flag indicates that the entire overlay Area and its
	// contents can be moved by dragging its header with the mouse.
	AreaCapabilityFlagDraggable = AreaCapabilityFlags(1 << iota)
)

// The AreaType is a registry construct that exists to convey implementation
// specifics to the Area itself. Each discrete Area type
// (such as the Stage viewport, shader editor, render graph, etc.)
// will exist as an AreaType.
type AreaType struct {
	ID      string
	Name    string
	Handler AreaHandler
	Flags   AreaCapabilityFlags
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
	Flags:   0,
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
	OverlayPX      float32
	OverlayPY      float32
}

type DebugSquare struct {
	posGetter func() matrix.Vec2
	panel     *ui.Panel
	size      float32
}

func (ds *DebugSquare) Update() {
	pos := ds.posGetter()
	ds.panel.Base().Layout().SetOffset(pos.X()-ds.size/2, pos.Y()-ds.size/2)
	ds.panel.Base().Layout().Scale(ds.size, ds.size)
	ds.panel.Base().Layout().SetZ(99)
	ds.panel.SetColor(matrix.ColorRed())
	half := ds.size / 2
	ds.panel.SetBorderRadius(half, half, half, half)
}

// Called by the workspace manager and serializer to set up this Area's
// manager and panel
func (a *Area) open(wm *WorkspaceManager) {
	a.openWithSize(wm, 10, 10, 0)
}

func (a *Area) openWithSize(wm *WorkspaceManager, width, height, radius float32) {
	a.Manager = &ui.Manager{}
	a.Manager.Init(wm.editor.Host())
	a.Root = a.Manager.Add().ToPanel()
	a.Root.Init(nil, ui.ElementTypePanel)
	a.SetBackdropColor(wm.editor.Theme().ColorPanel)
	a.Root.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	a.Root.Base().Layout().SetOffset(0, 0)
	a.Root.Base().Layout().Scale(float32(width), float32(height))
	a.Root.DontFitContent()
	a.Root.SetFlex()
	a.Root.SetFlexDirection(ui.FlexDirectionColumn)
	a.Root.SetFlexAlignItems(ui.FlexAlignStretch)
	a.Root.SetFlexJustify(ui.FlexJustifySpaceBetween)
	a.Root.SetBorderRadius(radius, radius, radius, radius)
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
		Flags:   0,
	}
	a.Manager = nil
	a.Root = nil
	a.Parent = nil
	a.ChildA = nil
	a.ChildB = nil
	a.SplitDirection = SplitHorizontal
	a.Ratio = -1
	a.OverlayPX = -1
	a.OverlayPY = -1
}

func (a *Area) SetBackdropColor(color matrix.Color) {
	a.assertInitialized()
	a.Root.SetColor(color)
}

func (a *Area) BackdropColor() matrix.Color {
	a.assertInitialized()
	return a.Root.Base().ShaderData().BgColor
}

func (a *Area) SetOutlineColor(color matrix.Color) {
	a.assertInitialized()
	a.Root.SetOutline(2, 0, color)
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

func (a *Area) TopLeftCorner() matrix.Vec2 {
	return a.Root.Base().Layout().PixelPosition()
}

func (a *Area) TopRightCorner() matrix.Vec2 {
	return a.TopLeftCorner().Add(matrix.Vec2{a.Root.Base().Layout().PixelSize().X(), 0})
}

func (a *Area) BottomLeftCorner() matrix.Vec2 {
	return a.TopLeftCorner().Add(matrix.Vec2{0, a.Root.Base().Layout().PixelSize().Y()})
}

func (a *Area) BottomRightCorner() matrix.Vec2 {
	return a.TopLeftCorner().Add(matrix.Vec2{a.Root.Base().Layout().PixelSize().X(), a.Root.Base().Layout().PixelSize().Y()})
}

// ConstrainAndScale sets the offset and dimensions of this Area's
// panel while ensuring that the panel cannot leave the window.
// The dimensions provided here are normalized against the DPMM of the window.
func (a *Area) ConstrainAndScale(editor EditorAreaInterface, posx, posy, width, height float32) {
	if a.Type.Handler == nil {
		return
	}
	a.assertInitialized()
	maxW := float32(editor.Host().Window.Width())
	maxH := float32(editor.Host().Window.Height())
	wx, wy := a.Type.Handler.GetOverlayDimensions()
	wx = min(max(wx, 256), maxW-24)
	wy = min(max(wy, 256), maxH-24)
}

func (a *Area) assertInitialized() {
	if a.Root == nil || a.Manager == nil {
		panic("UI audit on uninitialized area")
	}
}

func (a *Area) IsOverlay() bool {
	return a.OverlayPX > -1 && a.OverlayPY > -1
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
		ox, oy, ds, dy := a.Type.Handler.GetOffsets(editor)
		posx += ox
		posy += oy
		width += ds
		height += dy
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
