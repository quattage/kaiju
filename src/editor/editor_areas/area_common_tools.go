/******************************************************************************/
/* transform_gizmo.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_areas

import (
	"kaijuengine.com/matrix"
)

type ToolMode = uint8

const (
	ToolModeNone ToolMode = iota
	ToolModeMove
	ToolModeRotate
	ToolModeScale
)

type GizmoAxis = uint8

const (
	GizmoAxisX GizmoAxis = 1 << iota
	GizmoAxisY
	GizmoAxisZ
)

type Gizmo struct {
	axes GizmoAxis
	drag uint8
}

func (g *Gizmo) Enabled() bool {
	return g.axes > 0
}

func (g *Gizmo) Dragging() bool {
	return g.Enabled() && g.drag > 0
}

func (tg *Gizmo) DisableAll() {
	tg.axes = 0
}

func (tg *Gizmo) EnableAll() {
	tg.axes = 7
}

func (tg *Gizmo) SetX(enabled bool) {
	if enabled {
		tg.axes |= GizmoAxisX
	} else {
		tg.axes &^= GizmoAxisX
	}
}

func (tg *Gizmo) SetY(enabled bool) {
	if enabled {
		tg.axes |= GizmoAxisY
	} else {
		tg.axes &^= GizmoAxisY
	}
}

func (tg *Gizmo) SetZ(enabled bool) {
	if enabled {
		tg.axes |= GizmoAxisZ
	} else {
		tg.axes &^= GizmoAxisZ
	}
}

type CombinedGizmo struct {
	target   matrix.Transform
	Location Gizmo
	Rotation Gizmo
	Scale    Gizmo
}

func (cg *CombinedGizmo) Target() *matrix.Transform {
	return &cg.target
}
