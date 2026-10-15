/******************************************************************************/
/* editor_area_interface.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_areas

import (
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_active_stage"
	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/editor_settings"
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

	// UpdateSettings persists the current Settings struct and re-applies
	// frame rate / scroll speed / etc. to the live host.
	UpdateSettings()

	ShowReferences(id string)

	// SetMinimumWindoWSize ensures that the window cannot be smaller
	// than the provided width and height. This method is called by the editor
	// on startup to ensure that the editor window can't be resized in a way
	// that could cause issues with the Area system.
	SetMinimumWindowSize(width, height int)

	// GetActiveStage returns the current Stage selected by the editor.
	GetActiveStage() *editor_active_stage.ActiveStage
}
