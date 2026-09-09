/******************************************************************************/
/* editor_area_interface.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_areas

import (
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_database/content_previews"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine"
)

// EditorAreaInterface is an interface that provides a view of all editor
// functionality to every Area. This is what is provided to each Area
// during their lifetime functions.
//
// It intentionally exposes editor-level services
// (host, settings, events, history, project, stage view, content previewer)
// plus the workspace/area registry and switching API, but does not contain any
// per-workspace or per-area methods. Cross-area operations go through events
// (Events()) or through Workspace(id) lookups against well-known string IDs
// or typed service interfaces.
//
// Methods on this interface map 1:1 to methods on the Editor struct so the
// editor implements the interface implicitly.
type EditorAreaInterface interface {
	// Engine / runtime
	Host() *engine.Host
	Cache() *content_database.Cache
	ContentPreviewer() *content_previews.ContentPreviewer

	// Editor services
	Actions() *editor_action.Service
	Settings() *editor_settings.Settings
	Theme() *editor_settings.WorkspaceTheme
	Events() *editor_events.EditorEvents
	History() *memento.History
	Project() *project.Project
	ProjectFileSystem() *project_file_system.FileSystem
	StageView() *editor_stage_view.StageView

	// Focus management — Areas blur the rest of the editor while a
	// modal/overlay is in front of them and re-focus on close.
	BlurInterface()
	FocusInterface()
	IsInputFocused() bool

	// UpdateSettings persists the current Settings struct and re-applies
	// frame rate / scroll speed / etc. to the live host.
	UpdateSettings()

	// ShowReferences opens the references viewer overlay for the given
	// content id. Lives here because the overlay is editor-owned, not
	// Area-owned.
	ShowReferences(id string)

	// SwitchToWorkspace reconfigures the Area stack to match
	// what's descrubed by the WorkspaceConfiguration mapped to the
	// provided string ID, if one exists.
	SwitchToWorkspace(id string)
}
