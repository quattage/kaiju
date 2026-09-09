package splash_screen

import "kaijuengine.com/editor/editor_areas"

func init() {
	editor_areas.Register(func() editor_areas.AreaType {
		return editor_areas.AreaType{
			ID:      "com.kaiju.area_splash_screen",
			Name:    "Editor Splash Screen",
			Handler: &Handler{},
			Subtype: editor_areas.SubtypeOverlayOnly,
		}
	})
}

type Handler struct {
	// OnCreate will be called when the "Create" button is clicked, it will
	// return the name that the developer typed in and the path they selected.
	OnCreate func(name, path, templatePath string)
	// OnOpen will be called when the "Browse" button is clicked, it will return
	// the path that was selected.
	OnOpen func(string)
	// RecentProjects is a list of filepaths to recently closed projects.
	RecentProjects []string
}

func (as *Handler) Open(area *editor_areas.Area, editor editor_areas.EditorAreaInterface) {
	as.RecentProjects = editor.Settings().RecentProjects
}

func (as *Handler) Close(area *editor_areas.Area, editor editor_areas.EditorAreaInterface) {
}

func (as *Handler) FocusInterface(editor editor_areas.EditorAreaInterface) {

}

func (as *Handler) BlurInterface(editor editor_areas.EditorAreaInterface) {

}

func (as *Handler) Update(area *editor_areas.Area, editor editor_areas.EditorAreaInterface, deltaTime float64, posx, posy, width, height int) {
	area.Manager.EnableUpdate()
}

func (as *Handler) IsFocusedOnInput() bool {
	return true
}
