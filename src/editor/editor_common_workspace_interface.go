/******************************************************************************/
/* editor_common_workspace_interface.go                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_areas"
	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
)

// Editor implements editor_workspace.WorkspaceEditorInterface and (with
// the additions in editor_plugin.go) editor_plugin.EditorInterface. The
// methods below are the shared subset surfaced to every workspace.

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

func (ed *Editor) StageView() *editor_stage_view.StageView {
	return &ed.wsm.StageView
}

func (ed *Editor) UIWorkspace() *editor_areas.WorkspaceManager {
	return &ed.wsm
}

func (ed *Editor) BlurInterface() {
	ed.UIWorkspace().BlurInterface()
}

func (ed *Editor) FocusInterface() {
	ed.UIWorkspace().FocusInterface()
}

func (ed *Editor) ShowReferences(id string) {
	panic("unimpl")
}

func (ed *Editor) SwitchToWorkspace(id string) {
	ed.UIWorkspace().SwitchToWorkspace(id)
}
