/******************************************************************************/
/* throbber.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package throbber

import (
	"kaijuengine.com/editor/editor_areas"
	"kaijuengine.com/klib"
)

var LoadingScreenArea = editor_areas.AreaType{
	Key:            klib.NewKey("com.kaiju.area_loading_screen", "Loading Screen"),
	HandlerFactory: func() editor_areas.AreaHandlerInterface { return &Handler{} },
	Flags:          editor_areas.AreaCapabilityFlagOverlayable,
}

func init() {
	editor_areas.Register(LoadingScreenArea)
}

type Handler struct {
}

func (h *Handler) Open(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
	bkg := area.EmptyContainer(ed)
	bkg.AnnotateBottomLeft(ed, "Loading...")
}

func (h *Handler) Close(area *editor_areas.Area, editor editor_areas.EditorAreaInterface) {
}

func (h *Handler) FocusInterface(editor editor_areas.EditorAreaInterface) {
}

func (h *Handler) BlurInterface(editor editor_areas.EditorAreaInterface) {
}

func (h *Handler) Update(area *editor_areas.Area, editor editor_areas.EditorAreaInterface, deltaTime float64, posx, posy, width, height float32) {
}

func (h *Handler) GetOverlayDimensions() (float32, float32) {
	return 700, 300
}
