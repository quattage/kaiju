package area_example

import "kaijuengine.com/editor/editor_areas"

// This file contains the minimal definitions required to create and register a
// valid AreaType and define its UI handler. Plugin authors may find it convenient
// to copy and modify the contents of this file to skip the boilerplate.

var AreaTypeExample = editor_areas.AreaType{
	ID:      "com.kaiju.area_example",
	Name:    "Example Area ",
	Handler: &AreaExample{},
	Subtype: editor_areas.SubtypeDockable,
}

type AreaExample struct{}

func (as *AreaExample) Open(area *editor_areas.Area, editor editor_areas.EditorAreaView) {
}

func (as *AreaExample) Close(area *editor_areas.Area, editor editor_areas.EditorAreaView) {
	area.TakedownLayout()
}

func (as *AreaExample) FocusInterface(editor editor_areas.EditorAreaView) {

}

func (as *AreaExample) BlurInterface(editor editor_areas.EditorAreaView) {

}

func (as *AreaExample) Update(area *editor_areas.Area, editor editor_areas.EditorAreaView, deltaTime float64, posx, posy, width, height int) {

}

func (as *AreaExample) IsFocusedOnInput() bool {
	return true
}
