package functions

import (
	"fmt"
	"strconv"
	"strings"

	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
)

func (f ColorMix) Process(panel *ui.Panel, elm *document.Element, value rules.PropertyValue) (string, error) {

	args := value.Args
	length := len(args)

	if length < 3 {
		return helpers.ColorMap["magenta"], fmt.Errorf("too few arguments")
	}

	mixFactor := args[length-1]
	mixFactorTrimmed := strings.TrimSuffix(mixFactor, "%")
	var mixPercent float32
	if len(mixFactorTrimmed) != len(mixFactor) {
		numeric, err := strconv.ParseFloat(mixFactorTrimmed, 32)
		if err != nil {
			return helpers.ColorMap["magenta"], err
		}
		mixPercent = float32(numeric) / 100
	} else {
		numeric, err := strconv.ParseFloat(mixFactorTrimmed, 32)
		if err != nil {
			return helpers.ColorMap["magenta"], err
		}
		mixPercent = float32(numeric)
	}

	colorA, errA := findHex(args[length-2])
	if errA != nil {
		return helpers.ColorMap["magenta"], errA
	}
	colorB, errB := findHex(args[length-3])
	if errB != nil {
		return helpers.ColorMap["magenta"], errB
	}
	colorOut := matrix.ColorMix(colorA, colorB, mixPercent)

	return colorOut.Hex(), nil
}

func findHex(input string) (matrix.Color, error) {
	hex, ok := helpers.ColorMap[input]
	if ok {
		return matrix.ColorFromHexString(hex)
	}
	return matrix.ColorFromHexString(input)
}
