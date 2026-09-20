/******************************************************************************/
/* primary.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package primary

import (
	"kaijuengine.com/editor/editor_areas"
	"kaijuengine.com/editor/editor_areas/layouting"
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
	menu_bar   layouting.ContextBarElement
	status_bar layouting.ContextBarElement
}

func (h *Handler) Open(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
	area.SetBackdropColor(ed.Theme().ColorBackground)
	h.menu_bar = layouting.ContextBar(area, ed)
	h.menu_bar.LeftField().Label(ed, "Top Left")
	h.menu_bar.CenterField().Label(ed, "Top Center")
	h.menu_bar.RightField().Label(ed, "Top Right")
	h.status_bar = layouting.ContextBar(area, ed)
	h.status_bar.LeftField().Label(ed, "Bottom Left")
	h.status_bar.CenterField().Label(ed, "Bottom Center")
	h.status_bar.RightField().Label(ed, "Bottom Right")
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
