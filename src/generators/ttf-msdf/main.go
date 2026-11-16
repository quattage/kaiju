/******************************************************************************/
/* main.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package main

import (
	"os"
	"path/filepath"
	"strings"

	"kaijuengine.com/generators/common"
	"kaijuengine.com/klib"
)

func main() {
	if len(os.Args) == 1 {
		panic("Expected the first argument to be the TTF file to convert")
	}
	ttfName := os.Args[1]
	dirName := filepath.Join(common.BinDir, ttfName)
	if s, err := os.Stat(dirName); err != nil {
		panic(err)
	} else if s.IsDir() {
		os.Mkdir(filepath.Join(dirName, "out"), os.ModePerm)
		klib.Must(filepath.Walk(dirName, func(path string, _ os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Ext(path) == ".ttf" {
				common.GenerateSDFAtlasFromTTF(strings.TrimSuffix(path, ".ttf"))
			}
			return nil
		}))
	} else {
		common.GenerateSDFAtlasFromTTF(ttfName)
	}
}
