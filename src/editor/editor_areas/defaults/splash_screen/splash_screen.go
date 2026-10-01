/******************************************************************************/
/* splash_screen.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package splash_screen

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"kaijuengine.com/editor/editor_areas"
	"kaijuengine.com/editor/editor_embedded_content/versions"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/profiler/tracing"
)

func init() {
	editor_areas.Register(func() editor_areas.AreaType {
		return editor_areas.AreaType{
			ID:             "com.kaiju.area_splash_screen",
			Name:           "Editor Splash Screen",
			Flags:          editor_areas.AreaCapabilityFlagOverlayable,
			HandlerFactory: func() editor_areas.AreaHandler { return &Handler{} },
		}
	})
}

type Handler struct {
	RecentProjects []RecentProject
	// HandleCreateProject is called when the a new project is created via the
	// splash screen. This event is configured automatically by the editor.
	HandleCreateProject func(name, path, templatePath string)
	// HandleOpenProject is called when a project is opened via the splash
	// screen either by navigating to it directly or by clicking on its recent
	// project entry. This event is configured automatically by the editor.
	HandleOpenProject func(path string)
}

type RecentProject struct {
	project.Settings
	project.WorkspaceSession
	Path     string
	Modified string
	Time     string
}

func (h *Handler) Open(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
	h.RefreshRecents(ed.Settings().RecentProjects)
	area.SetMinimunWindowSize(ed)
	bkg := area.BackgroundImage(ed, "kaiju-splash.png", 1)
	bkg.TopShadow(ed, 0.8, 0.7)
	area.LogoImage(ed, "kaiju-wordmark.png", 0.4)
	version := bkg.AnnotateTopRight(ed, fmt.Sprintf("editor v%v", versions.Editor))
	version.SetOpacity(0.75)
	attrib := bkg.AnnotateBottomRight(ed, "Content Attribution")
	attrib.SetOpacity(0.75)
	area.HTMLContainer(ed, "splash.html", h, h.openRecentProject)
}

// RefreshRecents updates the splash screen's recent projects pane by ingesting
// the provided list of file paths. This method is called once automatically
// during the splash screen's Open process.
// Note that this method doesn't update or rebuild the Area's UI. If this is
// called after the splash screen has opened, in order to actually see new
// recent project entries in the splash screen's UI panel, you'll need to invoke
// a rebuild yourself.
func (h *Handler) RefreshRecents(recents []string) {
	if len(recents) <= 0 {
		slog.Warn("Skipped loading recent projects from an empty source")
	}
	defer tracing.NewRegion("splash_screen.RefreshRecents").End()
	slog.Debug(fmt.Sprintf("Loading %v recent projects", len(recents)))
	for x := range recents {
		settings, err := openProjectSettings(recents[x])
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to open recent project %v - couldn't read settings file", recents[x]), "error", err)
			continue
		}
		err = settings.Validate()
		if err != nil {
			slog.Warn(fmt.Sprintf("Skipped inclusion of recent project %v - Invalid settings file!", recents[x]), "warn", err)
			continue
		}
		h.RecentProjects = append(h.RecentProjects, RecentProject{
			Settings: *settings,
			Path:     recents[x],
			Modified: "2 months ago",
			Time:     "7/2/26 10:18pm",
		})
	}
}

func openProjectSettings(path string) (*project.Settings, error) {
	if len(path) <= 0 {
		return nil, fmt.Errorf("path is empty!")
	}
	file, err := os.Open(filepath.Join(path, project_file_system.ProjectConfigFile))
	if err != nil {
		return nil, err
	}
	settings := &project.Settings{}
	err = json.NewDecoder(file).Decode(settings)
	if err != nil {
		return nil, err
	}
	file, err = os.Open(filepath.Join(path, project_file_system.SessionRestoreFile))
	err = json.NewDecoder(file).Decode(settings)
	if err != nil {
		return settings, nil
	}
	return settings, nil
}

func (h *Handler) openRecentProject(elm *document.Element) {
	defer tracing.NewRegion("splash_screen.openRecentProject").End()
	h.openProjectAtPath(elm.Attribute("target"))
}

func (h *Handler) openProjectAtPath(path string) {
	if h.HandleOpenProject == nil {
		slog.Error("OnOpen hasn't been assigned yet! This indicates a load order issue.")
		return
	}
	h.HandleOpenProject(path)
}

func (h *Handler) Close(area *editor_areas.Area, editor editor_areas.EditorAreaInterface) {
	editor.SetMinimumWindowSize(-1, -1)
}

func (h *Handler) FocusInterface(editor editor_areas.EditorAreaInterface) {
}

func (h *Handler) BlurInterface(editor editor_areas.EditorAreaInterface) {
}

func (h *Handler) Update(area *editor_areas.Area, editor editor_areas.EditorAreaInterface, deltaTime float64, posx, posy, width, height float32) {
}

func (h *Handler) GetOffsets(ed editor_areas.EditorAreaInterface) (float32, float32, float32, float32) {
	return 0, 0, 0, 0
}

func (h *Handler) GetOverlayDimensions() (float32, float32) {
	return 800, 900
}

func (h *Handler) IsFocusedOnInput() bool {
	return true
}
