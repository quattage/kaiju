/******************************************************************************/
/* fallback.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package fallback

import "kaijuengine.com/editor/editor_areas"

// This AreaType definition is used when an AreaType ID can't be located
// due to a registry/plugin invalidation or workspace de-serialization issue.

func init() {
	editor_areas.Register(func() editor_areas.AreaType {
		return editor_areas.AreaType{
			ID:      "com.kaiju.area_fallback",
			Name:    "Fallback",
			Handler: &Handler{},
			Subtype: editor_areas.SubtypeDockable,
		}
	})
}

type Handler struct {
	// The ID of the AreaType whose failure prompted this Area to load instead
	FailedID string
}

func (as *Handler) Open(area *editor_areas.Area, editor editor_areas.EditorAreaInterface) {
}

func (as *Handler) Close(area *editor_areas.Area, editor editor_areas.EditorAreaInterface) {
	area.TakedownLayout()
}

func (as *Handler) FocusInterface(editor editor_areas.EditorAreaInterface) {

}

func (as *Handler) BlurInterface(editor editor_areas.EditorAreaInterface) {

}

func (as *Handler) Update(area *editor_areas.Area, editor editor_areas.EditorAreaInterface, deltaTime float64, posx, posy, width, height int) {

}

func (as *Handler) IsFocusedOnInput() bool {
	return false
}
