/******************************************************************************/
/* workspace_session.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package project

import (
	"encoding/json"
	"time"

	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/profiler/tracing"
)

type WorkspaceConfiguration struct {
	ID   string
	Name string
	tree []branch
}

type branch struct {
	a        int16
	b        int16
	ratio    float32
	dir      uint8
	areaType string
}

type Session struct {
	LastModified      time.Time `visible:"false"`
	selectedWorkspace uint16
	workspaces        []WorkspaceConfiguration
}

func (ws *Session) Save(fs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("Settings.Save").End()
	f, err := fs.Create(project_file_system.SessionRestoreFile)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(*ws)
}

func (ws *Session) load(fs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("Settings.load").End()
	f, err := fs.Open(project_file_system.SessionRestoreFile)
	if err != nil {
		return err
	}
	err = json.NewDecoder(f).Decode(ws)
	if err != nil {
		return err
	}
	ws.LastModified = time.Now()

	return ws.Save(fs)
}
