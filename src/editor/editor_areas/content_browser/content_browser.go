/******************************************************************************/
/* content_browser.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_browser

import (
	"kaijuengine.com/editor/editor_areas"
	"kaijuengine.com/klib"
)

var ContentBrowserArea = editor_areas.AreaType{
	Key:            klib.NewKey("com.kaiju.area_content_browser", "Content Browser"),
	HandlerFactory: func() editor_areas.AreaHandlerInterface { return &Handler{} },
	Flags:          editor_areas.AreaCapabilityFlagDockable | editor_areas.AreaCapabilityFlagDiscoverable,
}

func init() {
	editor_areas.Register(ContentBrowserArea)
}

type Handler struct {
}

func (h *Handler) Open(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
	area.AnnotateCenter(ed, "content")
}

func (h *Handler) Close(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
}

func (h *Handler) FocusInterface(ed editor_areas.EditorAreaInterface) {
}

func (h *Handler) BlurInterface(ed editor_areas.EditorAreaInterface) {
}

func (h *Handler) Update(area *editor_areas.Area, ed editor_areas.EditorAreaInterface, deltaTime float64, posx, posy, width, height float32) {
}

func (h *Handler) GetOverlayDimensions() (float32, float32) {
	return -1, -1
}
