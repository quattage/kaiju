/******************************************************************************/
/* example.go                                                                 */
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
			Subtype: editor_areas.SubtypeInternal,
		}
	})
}

type Handler struct {
	tb layouting.ContextBarElement
}

func (h *Handler) Open(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
	area.SetBackdropColor(ed.Theme().BackgroundColor.AsColor())
	h.tb = layouting.ContextBar(area, ed)
	h.tb.LeftField().Label("Top Left", ed)
	h.tb.CenterField().Label("Top Center", ed)
	h.tb.RightField().Label("Top Right", ed)
	sb := layouting.ContextBar(area, ed)
	sb.LeftField().Label("Bottom Left", ed)
	sb.CenterField().Label("Bottom Center", ed)
	sb.RightField().Label("Bottom Right", ed)
}

func (h *Handler) Close(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
}

func (h *Handler) FocusInterface(ed editor_areas.EditorAreaInterface) {}

func (h *Handler) BlurInterface(ed editor_areas.EditorAreaInterface) {}

func (h *Handler) Update(area *editor_areas.Area, ed editor_areas.EditorAreaInterface, deltaTime float64, posx, posy, width, height float32) {
}

func (h *Handler) GetDisplacement(ed editor_areas.EditorAreaInterface, posx, posy, width, height float32) (float32, float32, float32, float32) {
	return posx, posy + ed.Theme().ContextBarThickness, width, height - ed.Theme().ContextBarThickness*2
}

func (h *Handler) IsFocusedOnInput() bool {
	return false
}
