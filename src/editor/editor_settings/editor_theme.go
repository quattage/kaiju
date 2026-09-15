/******************************************************************************/
/* editor_theme.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_settings

import (
	"encoding/json"
	"os"
	"path/filepath"

	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/filesystem"
	"kaijuengine.com/platform/profiler/tracing"
)

const (
	themeFileName = "theme.json"
)

type WorkspaceTheme struct {
	// MinAreaRatio is the smallest ratio value permissable by composite Areas.
	// This is to ensure that Area splits don't become so small that their content
	// becomes inacccessible by the end user.
	MinAreaRatio        float32 `default:"0.1" visible:"false"`
	BackgroundColor     matrix.Color8
	ContextBarColor     matrix.Color8
	PanelColor          matrix.Color8
	PassiveTextColor    matrix.Color8
	ActiveTextColor     matrix.Color8
	ContextBarThickness klib.Unit
	FontSize            klib.Unit
	OverlayRoundness    klib.Unit
}

func (wt *WorkspaceTheme) setDefaults() {
	wt.MinAreaRatio = 0.1
	wt.BackgroundColor = matrix.NewColor8(5, 6, 7, 255)
	wt.ContextBarColor = matrix.NewColor8(9, 12, 15, 255)
	wt.PanelColor = matrix.NewColor8(14, 16, 18, 255)
	wt.PassiveTextColor = matrix.NewColor8(154, 168, 159, 255)
	wt.ActiveTextColor = matrix.NewColor8(224, 226, 226, 255)
	wt.ContextBarThickness = klib.Pix2MMU(24)
	wt.FontSize = klib.Pix2MMU(12)
	wt.OverlayRoundness = klib.Pix2MMU(8)
}

// only one theme is supported for now

func (wt *WorkspaceTheme) Save() error {
	defer tracing.NewRegion("WorkspaceTheme.Save").End()
	appData, err := filesystem.GameDirectory()
	if err != nil {
		return AppDataMissingError{err}
	}
	f, err := os.Create(filepath.Join(appData, themeFileName))
	if err != nil {
		return WriteError{err, false}
	}
	if err := json.NewEncoder(f).Encode(*wt); err != nil {
		return WriteError{err, true}
	}
	return nil
}

func (wt *WorkspaceTheme) Load() error {
	defer tracing.NewRegion("Settings.Load").End()
	appData, err := filesystem.GameDirectory()
	if err != nil {
		return AppDataMissingError{err}
	}
	wt.setDefaults()
	path := filepath.Join(appData, themeFileName)
	if _, err := os.Stat(path); err != nil {
		return wt.Save()
	}
	f, err := os.Open(path)
	if err != nil {
		return ReadError{err, false}
	}
	if err := json.NewDecoder(f).Decode(wt); err != nil {
		return ReadError{err, true}
	}
	return nil
}
