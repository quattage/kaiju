/******************************************************************************/
/* primary.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package primary

// The primary area is responsible for defining the main menu and status bars

import (
	"kaijuengine.com/editor/editor_areas"
)

func init() {
	editor_areas.Register(func() editor_areas.AreaType {
		return editor_areas.AreaType{
			ID:      "com.kaiju.area_primary",
			Name:    "Primary Area",
			Handler: &Handler{},
			Flags:   0,
		}
	})
}

type Handler struct {
	menu_bar   editor_areas.ContextBarElement
	status_bar editor_areas.ContextBarElement
}

func (h *Handler) Open(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
	area.StretchFlexColumn()
	area.SetBackdropColor(ed.Theme().ColorBackground)
	h.menu_bar = area.ContextBar(ed)
	h.status_bar = area.ContextBar(ed)
}

func (h *Handler) Close(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
}

func (h *Handler) FocusInterface(ed editor_areas.EditorAreaInterface) {}

func (h *Handler) BlurInterface(ed editor_areas.EditorAreaInterface) {}

func (h *Handler) Update(area *editor_areas.Area, ed editor_areas.EditorAreaInterface, deltaTime float64, posx, posy, width, height float32) {
}

func (h *Handler) GetOffsets(ed editor_areas.EditorAreaInterface) (float32, float32, float32, float32) {
	return 0, ed.Theme().SizeContextBar, 0, -ed.Theme().SizeContextBar * 2
}

func (h *Handler) GetOverlayDimensions() (float32, float32) {
	return -1, -1
}

func (h *Handler) IsFocusedOnInput() bool {
	return false
}
