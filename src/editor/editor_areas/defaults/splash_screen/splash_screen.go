/******************************************************************************/
/* splash_screen.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package splash_screen

import (
	"fmt"

	"kaijuengine.com/editor/editor_areas"
	"kaijuengine.com/editor/editor_embedded_content/versions"
	"kaijuengine.com/engine/ui/markup/document"
)

func init() {
	editor_areas.Register(func() editor_areas.AreaType {
		return editor_areas.AreaType{
			ID:      "com.kaiju.area_splash_screen",
			Name:    "Editor Splash Screen",
			Handler: &Handler{},
			Flags:   editor_areas.AreaCapabilityFlagOverlayable,
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

func (as *Handler) Open(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
	as.RecentProjects = ed.Settings().RecentProjects
	area.SetMinimunWindowSize(ed)
	bkg := area.BackgroundImage(ed, "kaiju-splash.png", 1)
	bkg.TopShadow(ed, 0.8, 0.7)
	area.LogoImage(ed, "kaiju-wordmark.png", 0.4)
	version := bkg.AnnotateTopRight(ed, fmt.Sprintf("editor v%v", versions.Editor))
	version.SetOpacity(0.75)
	attrib := bkg.AnnotateBottomRight(ed, "Content Attribution")
	attrib.SetOpacity(0.75)
	area.HTMLContainer(ed, "splash.html", testFunc)
}

func testFunc(doc *document.Element) {

}

func (as *Handler) Close(area *editor_areas.Area, editor editor_areas.EditorAreaInterface) {
	editor.SetMinimumWindowSize(-1, -1)
}

func (as *Handler) FocusInterface(editor editor_areas.EditorAreaInterface) {
}

func (as *Handler) BlurInterface(editor editor_areas.EditorAreaInterface) {
}

func (as *Handler) Update(area *editor_areas.Area, editor editor_areas.EditorAreaInterface, deltaTime float64, posx, posy, width, height float32) {
}

func (h *Handler) GetOffsets(ed editor_areas.EditorAreaInterface) (float32, float32, float32, float32) {
	return 0, 0, 0, 0
}

func (h *Handler) GetOverlayDimensions() (float32, float32) {
	return 700, 800
}

func (as *Handler) IsFocusedOnInput() bool {
	return true
}
