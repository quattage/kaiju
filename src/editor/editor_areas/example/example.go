/******************************************************************************/
/* example.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package example

import (
	"kaijuengine.com/editor/editor_areas"
	"kaijuengine.com/klib"
)

// This file contains the minimal definitions required to create and register a
// valid AreaType and define its UI handler. Reference this for creating Areas
// via plugins.

// Create the area's type definition with a unique ID and a friendly name.
var ExampleArea = editor_areas.AreaType{
	Key:            klib.NewKey("com.kaiju.area_example", "Example Area"),
	HandlerFactory: func() editor_areas.AreaHandlerInterface { return &Handler{} },
	Flags:          editor_areas.AreaCapabilityFlagDockable,
}

// Registering the AreaType keeps its handler factory in memory and allows the
// workspace manager to serialize it so that its data persists between editor
// launches.
func init() {
	editor_areas.Register(ExampleArea)
}

// Implement the Area using the contract described by AreaHandlerInterface
type Handler struct{}

// Called every time this Area is opened by the workspace manager, whether that
// be by the user or by the workspace manager itself.
func (h *Handler) Open(area *editor_areas.Area, editor editor_areas.EditorAreaInterface) {
}

// Called every time this Area is closed by the workspace manager, whether that
// be by the user or by the workspace manager itself.
func (h *Handler) Close(area *editor_areas.Area, editor editor_areas.EditorAreaInterface) {
}

// Called when this Area is focused. This area is focused when the user's cursor
// enters its rect.
func (h *Handler) FocusInterface(editor editor_areas.EditorAreaInterface) {
}

// Called when this Area is blurred. This area is focused when the user's cursor
// enters its rect.
func (h *Handler) BlurInterface(editor editor_areas.EditorAreaInterface) {
}

// Called continuously while this Area exists and is being drawn. Don't store
// the positional data passed here since it won't be up-to-date outside the
// scope of this method.
func (h *Handler) Update(area *editor_areas.Area, editor editor_areas.EditorAreaInterface, deltaTime float64, posx, posy, width, height float32) {
}

// Specify the dimensions of this Area when it is opened as an Overlay. If you
// don't ever plan on supporting overlay functionality for this Area, just
// return -1, -1 here.
func (h *Handler) GetOverlayDimensions() (float32, float32) {
	return -1, -1
}
