/******************************************************************************/
/* example.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package example

import "kaijuengine.com/editor/editor_areas"

// This file contains the minimal definitions required to create and register a
// valid AreaType and define its UI handler. Plugin authors may find it convenient
// to copy and modify the contents of this file instead of manually writing the
// boilerplate.

func init() {
	editor_areas.Register(func() editor_areas.AreaType {
		return editor_areas.AreaType{
			ID:      "com.kaiju.area_example",
			Name:    "ExampleArea",
			Handler: &Handler{},
			Subtype: editor_areas.SubtypeRestricted,
		}
	})
}

type Handler struct{}

func (as *Handler) Open(area *editor_areas.Area, editor editor_areas.EditorAreaInterface) {
}

func (as *Handler) Close(area *editor_areas.Area, editor editor_areas.EditorAreaInterface) {
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
