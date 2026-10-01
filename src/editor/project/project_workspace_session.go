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

type WorkspaceSession struct {
	LastModified time.Time `visible:"false"`
}

func (ws *WorkspaceSession) Save(fs *project_file_system.FileSystem) error {
	// walk the worksapce stack and serialize tne binary tree and all Area settings
	defer tracing.NewRegion("Settings.Save").End()
	f, err := fs.Create(project_file_system.SessionRestoreFile)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(*ws)
}

func (ws *WorkspaceSession) load(fs *project_file_system.FileSystem) error {
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

	// TODO address any mismatches and repair broken Area states individually
	return ws.Save(fs)
}
