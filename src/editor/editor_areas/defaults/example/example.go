package example

import "kaijuengine.com/editor/editor_areas"

// This file contains the minimal definitions required to create and register a
// valid AreaType and define its UI handler. Plugin authors may find it convenient
// to copy and modify the contents of this file to skip the boilerplate.

var Type = editor_areas.AreaType{
	ID:      "com.kaiju.area_example",
	Name:    "Example Area ",
	Handler: &Handler{},
	Subtype: editor_areas.SubtypeDockable,
}

type Handler struct{}

func (as *Handler) Open(area *editor_areas.Area, editor editor_areas.EditorAreaView) {
}

func (as *Handler) Close(area *editor_areas.Area, editor editor_areas.EditorAreaView) {
	area.TakedownLayout()
}

func (as *Handler) FocusInterface(editor editor_areas.EditorAreaView) {

}

func (as *Handler) BlurInterface(editor editor_areas.EditorAreaView) {

}

func (as *Handler) Update(area *editor_areas.Area, editor editor_areas.EditorAreaView, deltaTime float64, posx, posy, width, height int) {

}

func (as *Handler) IsFocusedOnInput() bool {
	return true
}
