/******************************************************************************/
/* throbber.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package throbber

import (
	"log/slog"

	"kaijuengine.com/editor/editor_areas"
)

func init() {
	editor_areas.Register(func() editor_areas.AreaType {
		return editor_areas.AreaType{
			ID:             "com.kaiju.area_throbber",
			Name:           "Editor Loading Throbber",
			Flags:          editor_areas.AreaCapabilityFlagOverlayable,
			HandlerFactory: func() editor_areas.AreaHandler { return &Handler{} },
		}
	})
}

type Handler struct {
}

func (h *Handler) Open(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
	slog.Info("" + area.String())
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

func (h *Handler) GetOffsets(ed editor_areas.EditorAreaInterface) (float32, float32, float32, float32) {
	return 0, 0, 0, 0
}

func (h *Handler) GetOverlayDimensions() (float32, float32) {
	return 700, 300
}

func (h *Handler) IsFocusedOnInput() bool {
	return true
}
