/******************************************************************************/
/* editor.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"log/slog"
	"time"

	"kaijuengine.com/build"
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_areas"
	"kaijuengine.com/editor/editor_embedded_content"
	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/editor_logging"
	"kaijuengine.com/editor/editor_plugin"
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_database/content_previews"
	"kaijuengine.com/editor/webapi"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	platformPower "kaijuengine.com/platform/power"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/textures"
)

// Editor is the entry point structure for the entire editor. It acts as the
// delegate to the various systems and holds the primary members that make up
// the bulk of the editor identity.
//
// The design goal of the editor is different than that of the [engine.Host], as
// it is not intended to be passed around for access to the system. Instead it
// will supply interface functions that are needed to the systems that it holds
// internally.
type Editor struct {
	host           *engine.Host
	settings       editor_settings.Settings
	project        project.Project
	logging        editor_logging.Logging
	history        memento.History
	events         editor_events.EditorEvents
	wsm            editor_areas.WorkspaceManager
	plugins        []editor_plugin.EditorPlugin
	fileDropRouter FileDropRouter
	window         struct {
		activateId     events.Id
		deactivateId   events.Id
		lastActiveTime time.Time
	}
	contentPreviewer content_previews.ContentPreviewer
	updateId         engine.UpdateId
	webAPIServer     *webapi.Server[*Editor]
	actions          *editor_action.Service
	actionPaletteKey struct {
		pending bool
		moved   bool
	}
	power powerState
	// sessionDisabledPlugins holds module paths of plugins the user chose
	// to skip via the startup-validation modal's "Continue" button (only
	// MISSING plugins are recorded here — stale plugins are not tracked,
	// see editor_plugin_validation.go for the rationale). Process-local
	// only; never persisted to plugin.json. Cleared at next process start.
	// Read by MissingCompiledPlugins to suppress repeat modals within the
	// same process. Touched only from the main UI goroutine — no lock
	// needed.
	sessionDisabledPlugins map[string]struct{}
}

const (
	editorPowerPollInterval         = 5.0
	launchMaxTextureCreatesPerFrame = 8
	launchMaxTextureBytesPerFrame   = 32 * 1024 * 1024
)

type powerStatusQuery func() (platformPower.Status, error)

type powerState struct {
	query       powerStatusQuery
	lastStatus  platformPower.Status
	initialized bool
	pollElapsed float64
}

func (ed *Editor) Host() *engine.Host { return ed.host }

func (ed *Editor) ContentPreviewer() *content_previews.ContentPreviewer {
	return &ed.contentPreviewer
}

func (ed *Editor) IsInputFocused() bool {
	return false
}

func (ed *Editor) earlyLoadUI() {
	defer tracing.NewRegion("Editor.earlyLoadUI").End()
	ed.wsm.Initialize(ed)
}

func (ed *Editor) UpdateSettings() {
	ed.setFrameRateLimitForPowerStatus(ed.queryAndCachePowerStatus())
	if matrix.Approx(ed.settings.UIScrollSpeed, 0) {
		ed.settings.UIScrollSpeed = 1
	}
	ed.settings.NormalizeWebAPI()
	ui.UIScrollSpeed = ed.settings.UIScrollSpeed
	if err := ed.settings.Save(); err != nil {
		slog.Error("failed to save the editor settings", "error", err)
		return
	}
	ed.updateWebAPI()
}

func (ed *Editor) queryAndCachePowerStatus() platformPower.Status {
	status := ed.queryPowerStatus()
	ed.power.lastStatus = status
	ed.power.initialized = true
	ed.power.pollElapsed = 0
	return status
}

func (ed *Editor) queryPowerStatus() platformPower.Status {
	query := ed.power.query
	if query == nil {
		query = platformPower.Query
	}
	status, err := query()
	if err != nil {
		slog.Debug("failed to query power status", "error", err)
		return platformPower.Status{Source: platformPower.SourceUnknown, BatteryPercent: -1}
	}
	return status
}

func (ed *Editor) setFrameRateLimitForPowerStatus(status platformPower.Status) {
	if ed.host == nil {
		return
	}
	ed.host.SetFrameRateLimit(int64(ed.effectiveRefreshRate(status)))
}

func (ed *Editor) effectiveRefreshRate(status platformPower.Status) int32 {
	refreshRate := ed.settings.RefreshRate
	if ed.settings.UseBatteryRefreshRate && status.OnBattery() {
		refreshRate = ed.settings.BatteryRefreshRate
	}
	return klib.Clamp(refreshRate, 0, 320)
}

// initialLoad is called just after plugins have finished validating.
// This is used to start up the engine splash and bootstrap the currently
// active UI workspace
func (ed *Editor) initialLoad() {
	ed.UIWorkspace().Open()
}

func (ed *Editor) postProjectLoad() {
	defer tracing.NewRegion("Editor.lateLoadUI").End()
	ed.settings.AddRecentProject(ed.project.FileSystem().FullPath(""))
	slog.Info("compiling the project to get things ready")
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
	if build.Debug && ed.initAutoTest() {
		ed.updateId = ed.host.Updater.AddUpdate(ed.runAutoTest)
	} else {
		ed.updateId = ed.host.Updater.AddUpdate(ed.update)
	}
	for k, v := range editorPluginRegistry {
		if err := v.Launch(ed); err != nil {
			slog.Error("failed to launch plugin", "key", k, "error", err)
			continue
		}
		ed.plugins = append(ed.plugins, v)
	}
	// Pre-warm the, quite large, material icons PNG file
	ed.host.TextureCache().Texture("MaterialIcons-Regular.png", textures.TextureFilterLinear)
}

func (ed *Editor) update(deltaTime float64) {

}

func (ed *Editor) updatePowerState(deltaTime float64) {
	if !ed.settings.UseBatteryRefreshRate {
		return
	}
	if !ed.power.initialized {
		ed.setFrameRateLimitForPowerStatus(ed.queryAndCachePowerStatus())
		return
	}
	ed.power.pollElapsed += deltaTime
	if ed.power.pollElapsed < editorPowerPollInterval {
		return
	}
	ed.power.pollElapsed = 0
	status := ed.queryPowerStatus()
	if status.Source == ed.power.lastStatus.Source && status.HasBattery == ed.power.lastStatus.HasBattery {
		ed.power.lastStatus = status
		return
	}
	ed.power.lastStatus = status
	ed.setFrameRateLimitForPowerStatus(status)
}
