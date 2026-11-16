package area_selector

import (
	"kaijuengine.com/editor/editor_areas"
	"kaijuengine.com/klib"
)

// This file contains the minimal definitions required to create and register a
// valid AreaType and define its UI handler. Reference this for creating Areas
// via plugins.

// Create the area's type definition with a unique ID and a friendly name.
var ExampleArea = editor_areas.AreaType{
	Key:            klib.NewKey("com.kaiju.area_selector", "Selector"),
	HandlerFactory: func() editor_areas.AreaHandlerInterface { return &AreaSelectorHandler{} },
	Flags:          editor_areas.AreaCapabilityFlagOverlayable,
}

func init() {
	editor_areas.Register(ExampleArea)
}

type AreaSelectorHandler struct{}

func (h *AreaSelectorHandler) Open(area *editor_areas.Area, editor editor_areas.EditorAreaInterface) {
}

func (h *AreaSelectorHandler) Close(area *editor_areas.Area, editor editor_areas.EditorAreaInterface) {
}

func (h *AreaSelectorHandler) FocusInterface(editor editor_areas.EditorAreaInterface) {
}

func (h *AreaSelectorHandler) BlurInterface(editor editor_areas.EditorAreaInterface) {
}

func (h *AreaSelectorHandler) Update(area *editor_areas.Area, editor editor_areas.EditorAreaInterface, deltaTime float64, posx, posy, width, height float32) {
}

func (h *AreaSelectorHandler) GetOverlayDimensions() (float32, float32) {
	return -1, -1
}
