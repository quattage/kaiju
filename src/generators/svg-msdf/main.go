package main

import (
	"os"
	"path/filepath"

	"kaijuengine.com/generators/common"
)

func main() {
	if err := common.GenerateSDFAtlasFromSVGs("../icons"); err != nil {
		panic(err)
	}
	outputDir := "./editor/editor_embedded_content/editor_content/fonts"
	if err := os.Rename("../icons/out/TempIcons.png", filepath.Join(outputDir, "KaijuIcons-Regular.png")); err != nil {
		panic(err)
	}
	if err := os.Rename("../icons/out/TempIcons.bin", filepath.Join(outputDir, "KaijuIcons-Regular.bin")); err != nil {
		panic(err)
	}
}
