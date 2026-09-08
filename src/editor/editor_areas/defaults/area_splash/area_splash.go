package area_splash

import "kaijuengine.com/editor/editor_areas"

var AreaTypeSplash = editor_areas.AreaType{
	ID:      "com.kaiju.area_splash",
	Name:    "Editor Startup Splash",
	Handler: &AreaSplash{},
	Subtype: editor_areas.SubtypeRestricted,
}

type AreaSplash struct {
	// OnCreate will be called when the "Create" button is clicked, it will
	// return the name that the developer typed in and the path they selected.
	OnCreate func(name, path, templatePath string)
	// OnOpen will be called when the "Browse" button is clicked, it will return
	// the path that was selected.
	OnOpen func(string)
	// RecentProjects is a list of paths to recent projects.
	RecentProjects []string
}

func (as *AreaSplash) Open(area *editor_areas.Area, editor editor_areas.EditorAreaView) {
	as.RecentProjects = editor.Settings().RecentProjects
}

func (as *AreaSplash) Close(area *editor_areas.Area, editor editor_areas.EditorAreaView) {
}

func (as *AreaSplash) FocusInterface(editor editor_areas.EditorAreaView) {

}

func (as *AreaSplash) BlurInterface(editor editor_areas.EditorAreaView) {

}

func (as *AreaSplash) Update(area *editor_areas.Area, editor editor_areas.EditorAreaView, deltaTime float64, posx, posy, width, height int) {
	area.Manager.EnableUpdate()
}

func (as *AreaSplash) IsFocusedOnInput() bool {
	return true
}
