/******************************************************************************/
/* editor_common_workspace_interface.go                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_active_stage"
	"kaijuengine.com/editor/editor_areas"
	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
)

// This file contains Editor's implementation of
// EditorPluginInterface, EditorAreaInterface, and EditorToolsInterface.

func (ed *Editor) Events() *editor_events.EditorEvents {
	return &ed.events
}

func (ed *Editor) Actions() *editor_action.Service {
	if ed.actions == nil {
		ed.initializeActions()
	}
	return ed.actions
}

func (ed *Editor) History() *memento.History {
	return &ed.history
}

func (ed *Editor) Project() *project.Project {
	return &ed.project
}

func (ed *Editor) ProjectFileSystem() *project_file_system.FileSystem {
	return ed.project.FileSystem()
}

func (ed *Editor) Cache() *content_database.Cache {
	return ed.project.CacheDatabase()
}

func (ed *Editor) Settings() *editor_settings.Settings {
	return &ed.settings
}

func (ed *Editor) Theme() *editor_settings.WorkspaceTheme {
	return &ed.theme
}

func (ed *Editor) UIWorkspace() *editor_areas.WorkspaceManager {
	return &ed.workspace
}

func (ed *Editor) ShowReferences(id string) {
	panic("unimpl")
}

func (ed *Editor) GetActiveStage() *editor_active_stage.ActiveStage {
	return ed.workspace.GetActiveStage()
}
