/******************************************************************************/
/* editor_project_setup.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"fmt"
	"log/slog"

	"kaijuengine.com/build"
	"kaijuengine.com/editor/editor_embedded_content"
	"kaijuengine.com/editor/editor_embedded_content/versions"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/engine"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

func CreateNewProjectFromCLI(path string) {
	if build.Editor {
		proj := project.Project{}
		templatePath := engine.LaunchParams.ProjectTemplate
		if err := proj.Initialize(path, templatePath, versions.Editor); err != nil {
			slog.Error("failed to create the project", "error", err, "path", path)
			return
		}
		if name := engine.LaunchParams.ProjectName; name != "" {
			proj.SetName(name)
		}
		if err := proj.Close(); err != nil {
			slog.Error("failed to save the project configuration", "error", err)
			return
		}
		slog.Info("successfully created blank project", "path", path)
	} else {
		slog.Error("the -newproject flag is only available in editor builds")
	}
}

func (ed *Editor) createProject(name, path, templatePath string) {
	slog.Info("Creating new project...")
	ed.UIWorkspace().CloseOverlay("com.kaiju.area_splash_screen")
	ed.UIWorkspace().OpenOverlay("com.kaiju.area_throbber", -1, -1)
	// goroutine
	go func() {
		defer tracing.NewRegion("Editor.createProject").End()
		err := ed.project.Initialize(path, templatePath, versions.Editor)
		if err != nil && !klib.ErrorIs[project.ConfigLoadError](err) {
			slog.Error("failed to create the project", "error", err)
			return
		}
		ed.Host().RunOnMainThread(func() {
			ed.finalizeProjectLoad(ed.project.Name())
		})
	}()
}

func (ed *Editor) openProject(path string) {
	slog.Info("Opening project...")
	ed.UIWorkspace().CloseOverlay("com.kaiju.area_splash_screen")
	ed.UIWorkspace().OpenOverlay("com.kaiju.area_throbber", -1, -1)
	// goroutine
	go func() {
		defer tracing.NewRegion("Editor.openProject").End()
		if err := ed.project.Open(path); err != nil {
			slog.Error("failed to open the project", "error", err)
			lastCount := len(ed.settings.RecentProjects)
			ed.settings.RecentProjects = klib.SlicesRemoveElement(ed.settings.RecentProjects, path)
			if len(ed.settings.RecentProjects) != lastCount {
				ed.settings.Save()
			}
			return
		}
		projectVersion := ed.project.Settings.EditorVersion
		hasEngineSource := ed.project.FileSystem().HasEngineCode()
		// This is a special hidden feature for editor/engine developers to be able
		// to force updating engine code in projects. This makes it easier than
		// bumping the engine version to do the same thing (or deleting kaiju src)
		kb := &ed.host.Window.Keyboard
		forceReplace := kb.HasShift() || kb.HasCtrlOrMeta()
		if projectVersion != versions.Editor || !hasEngineSource || forceReplace {
			// TODO project upgrading
			slog.Error("Project Version Mispoat")
			return
		}
		// simple projects load fast so the UI is never seen
		// time.Sleep(time.Second)
		ed.Host().RunOnMainThread(func() {
			ed.finalizeProjectLoad(ed.project.Name())
		})
	}()
}

func (ed *Editor) finalizeProjectLoad(projectName string) {
	defer tracing.NewRegion("Editor.finalizeProjectLoad").End()
	ed.host.Window.SetTitle(fmt.Sprintf("%s - Kaiju Engine Editor", projectName))
	ed.project.SetName(projectName)
	ed.settings.AddRecentProject(ed.project.FileSystem().FullPath(""))
	slog.Info("Compiling the project to get things ready...")
	{
		// Read the project source synchronosly for now, if not, any stage loading
		// before this is complete will have issues.
		ed.project.ReadSourceCode()
	}
	editorContent := ed.host.AssetDatabase().(*editor_embedded_content.EditorContent)
	editorContent.Pfs = ed.project.FileSystem()
	editorContent.SetProjectContentIndex(ed.project.CacheDatabase().List())
	ed.events.OnContentAdded.Add(func(ids []string) {
		editorContent.IndexProjectContentIDs(ed.project.CacheDatabase(), ids)
	})
	ed.events.OnContentRemoved.Add(editorContent.RemoveProjectContentIDs)
	ed.host.TextureCache().SetUploadBudget(rendering.TextureUploadBudget{
		MaxCreatesPerFrame: launchMaxTextureCreatesPerFrame,
		MaxBytesPerFrame:   launchMaxTextureBytesPerFrame,
	})
	ed.setupWindowActivity()
	ed.connectFileDropRouter()
	// goroutine
	go ed.project.CompileDebug()

	for k, v := range editorPluginRegistry {
		if err := v.Launch(ed); err != nil {
			slog.Error("Failed to launch plugin", "key", k, "error", err)
			continue
		}
		ed.plugins = append(ed.plugins, v)
	}
	ed.UIWorkspace().CloseOverlay("com.kaiju.area_throbber")
	slog.Info("Project load complete!")
}
