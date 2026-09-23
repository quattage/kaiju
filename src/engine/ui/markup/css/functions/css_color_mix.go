package functions

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
)

type colorMixStop struct {
	color      matrix.Color
	percentage float32
	explicit   bool
}

func (f ColorMix) Process(panel *ui.Panel, elm *document.Element, value rules.PropertyValue) (string, error) {
	args := value.Args

	slog.Warn("PROCESS COLOR MIX " + value.Str)

	if len(args) < 4 {
		return "", fmt.Errorf(
			"color-mix requires a color space and two colors",
		)
	}

	// we don't support anything other than srgb for now
	if args[0] != "in" {
		args[0] = "in"
	}
	if !strings.EqualFold(args[1], "srgb") {
		args[1] = "srgb"
	}

	stops, err := parseColorMixStops(args[2:])
	if err != nil {
		return "", err
	}

	left := stops[0]
	right := stops[1]
	switch {
	case !left.explicit && !right.explicit:
		left.percentage = 50
		right.percentage = 50
	case left.explicit && !right.explicit:
		right.percentage = 100 - left.percentage
	case !left.explicit && right.explicit:
		left.percentage = 100 - right.percentage
	default:
		break
	}

	sum := left.percentage + right.percentage
	if sum <= 0 {
		return "", fmt.Errorf("color-mix percentages must have a positive sum")
	}

	amount := right.percentage / sum
	amount = klib.Clamp(amount, 0, 1)
	result := matrix.ColorMix(left.color, right.color, amount)
	return result.Hex(), nil
}

func parseColorMixStops(args []string) ([2]colorMixStop, error) {
	var stops [2]colorMixStop

	colorIndex := 0
	for x := 0; x < len(args); {
		if colorIndex >= len(stops) {
			break
		}
		color, err := parseColorMixColor(args[x])
		if err != nil {
			return stops, err
		}
		x++
		stop := colorMixStop{color: color}
		if x < len(args) && strings.HasSuffix(args[x], "%") {
			percentage, err := parseColorMixPercentage(args[x])
			if err != nil {
				return stops, err
			}
			stop.percentage = percentage
			stop.explicit = true
			x++
		}
		stops[colorIndex] = stop
		colorIndex++
	}

	if colorIndex != 2 {
		return stops, fmt.Errorf(
			"color-mix requires exactly two colors",
		)
	}

	return stops, nil
}

func parseColorMixColor(value string) (matrix.Color, error) {
	if mapped, ok := helpers.ColorMap[value]; ok {
		value = mapped
	}
	color, err := matrix.ColorFromHexString(value)
	if err != nil {
		return matrix.ColorTransparent(), fmt.Errorf(
			"invalid color %q in color-mix: %w",
			value,
			err,
		)
	}
	return color, nil
}

func parseColorMixPercentage(value string) (float32, error) {
	raw := strings.TrimSuffix(value, "%")
	if len(raw) != len(value) {
		percentage, err := strconv.ParseFloat(raw, 32)
		if err != nil {
			return 0, fmt.Errorf(
				"invalid color-mix percentage %q",
				value,
			)
		}
		return klib.Clamp(float32(percentage), 0, 1) * 100, nil
	}
	percentage, err := strconv.ParseFloat(raw, 32)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid color-mix percentage %q",
			value,
		)
	}
	return klib.Clamp(float32(percentage), 0, 100), nil
}
