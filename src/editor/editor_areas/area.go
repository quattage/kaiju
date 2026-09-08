package editor_areas

import (
	"fmt"
	"log/slog"
	"math"
	"reflect"
	"runtime"

	"kaijuengine.com/engine/ui"
)

type AreaHandler interface {
	Open(area *Area, editor EditorAreaView)
	Close(area *Area, editor EditorAreaView)
	FocusInterface(editor EditorAreaView)
	BlurInterface(editor EditorAreaView)
	Update(area *Area, editor EditorAreaView, deltaTime float64, posx, posy, width, height int)
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
	SubtypeOverlay
)

// The AreaType is a registry construct that exists to convey implementation
// specifics to the Area itself. Each discrete Area type (such as the Stage viewport,
// shader editor, render graph, etc.) will exist as an AreaType.
type AreaType struct {
	ID      string
	Name    string
	Handler AreaHandler
	Subtype AreaSubtype
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

// The ContextBar is a dedicated, horizontally-aligned strip of UI elements
// situated at the top or bottom of an Area with its own dedicated handler.
type ContextBar struct {
	Parent    *Area
	Thickness uint32
	Handler   AreaHandler
}

// "Composite" Areas don't have any UI of their own and instead
// defer all their functionality to their children.
func (a *Area) IsComposite() bool {
	return a.Type.ID == AreaTypeComposite.ID
}

// IsDockable returns true if this Area is allowed to
// be parented to another Area.
func (a *Area) IsDockable() bool {
	return a.Type.Subtype != SubtypeOverlay
}

func (a *Area) TakedownLayout() {
	a.Manager.Shutdown()
	a.Root = nil
}

// PerformAsChildren selectively recurses into child Areas and runs the supplied function, provided
// the Area either has its own children or a valid handler.
func (a *Area) PerformAsChildren(operation func(AreaHandler), editor EditorAreaView) {
	if a.ChildA != nil && a.ChildB != nil {
		a.ChildA.PerformAsChildren(operation, editor)
		a.ChildB.PerformAsChildren(operation, editor)
	} else if a.Type.Handler != nil {
		operation(a.Type.Handler)
	} else {
		slog.Error(fmt.Sprintf("Failed to perform operation '%s' on '%s' - This Area has no children or handler!", runtime.FuncForPC(reflect.ValueOf(operation).Pointer()).Name(), a))
	}
}

func (a *Area) Update(editor EditorAreaView, deltaTime float64, posx, posy, width, height int) {
	if a.ChildA != nil && a.ChildB != nil {
		switch a.SplitDirection {
		case SplitHorizontal:
			split := int(math.Round(float64(width) * float64(a.Ratio)))
			a.ChildA.Update(editor, deltaTime, posx, posy, split, height)
			a.ChildB.Update(editor, deltaTime, split, posy, width-split, height)
		case SplitVertical:
			split := int(math.Round(float64(height) * float64(a.Ratio)))
			a.ChildA.Update(editor, deltaTime, posx, posy, width, height-split)
			a.ChildB.Update(editor, deltaTime, posx, split, width, height-split)
		}
	} else if a.Type.Handler != nil {
		a.Type.Handler.Update(a, editor, deltaTime, posx, posy, width, height)
	} else {
		slog.Error(fmt.Sprintf("Skipped updating Area due to missing handler - %s [%v, %v, %v, %v]", a, posx, posy, width, height))
	}
}
