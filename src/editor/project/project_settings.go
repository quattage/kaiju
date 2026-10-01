/******************************************************************************/
/* project_settings.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package project

import (
	"encoding/json"
	"fmt"

	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/platform/profiler/tracing"
)

type Settings struct {
	Name                 string
	Organization         string
	Repository           string
	EntryPointStage      string
	ArchiveEncryptionKey string
	Android              AndroidSettings
	EditorVersion        float64 `visible:"false"`
}

type AndroidSettings struct {
	RootProjectName string
	ApplicationId   string
}

func (c *Settings) Save(fs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("Settings.Save").End()
	f, err := fs.Create(project_file_system.ProjectConfigFile)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(*c)
}

func (c *Settings) load(fs *project_file_system.FileSystem) error {
	defer tracing.NewRegion("Settings.load").End()
	f, err := fs.Open(project_file_system.ProjectConfigFile)
	if err != nil {
		return err
	}
	err = json.NewDecoder(f).Decode(c)
	if err != nil {
		return err
	}
	return c.Save(fs)
}

func (c *Settings) Validate() error {
	var err error = nil
	if len(c.Name) < 3 {
		err = fmt.Errorf("Name must be at least 3 characters!")
	}
	return err
}
