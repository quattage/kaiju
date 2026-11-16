/******************************************************************************/
/* shader_data_grid.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shader_data_registry

import (
	"unsafe"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

type GridFeatureFlags uint32

const (
	GridAligned GridFeatureFlags = 1 << iota
	GridPlaneXY
	GridPlaneXZ
	GridPlaneYZ
	GridAxisX
	GridAxisY
	GridAxisZ
)

func init() {
	register(func() rendering.DrawInstance {
		sdg := ShaderDataGrid{
			ShaderDataBase: rendering.NewShaderDataBase(),
		}
		sdg.SetDefaults()
		return &sdg
	}, "grid")
}

type ShaderDataGrid struct {
	rendering.ShaderDataBase `visible:"false"`
	GridFeatureFlags         uint32
	GridLevel                float32
	FarClip                  float32
	GridLineCount            uint32
	ClipRect                 matrix.Vec4
	Steps                    [8]matrix.Vec4
	Colors                   [4]matrix.Color
}

func (sd *ShaderDataGrid) SetDefaults() {
	sd.GridFeatureFlags = uint32(GridPlaneXZ | GridAxisX | GridAxisZ)
	sd.GridLevel = 1.1
	sd.GridLineCount = 151
	sd.ClipRect = matrix.NewVec4(0.1, 0.1, 0, 0)
	value := float32(1)
	for i := range sd.Steps {
		sd.Steps[i] = matrix.Vec4{value, value, value, 0}
		value *= 10
	}

}

func (sd ShaderDataGrid) VertexCount() uint32 {
	// 4 vertex terminals * 3 grid levels * x total lines
	// the 3 here always must match the OVERLAY_GRID_STEPS_DRAW constant
	// defined in grid_shared.glsl
	return 4 * 3 * sd.GridLineCount
}

func (sd *ShaderDataGrid) CalculateLevel(distance float32) {
	for x := range len(sd.Steps) {
		current := sd.Steps[x][0]
		next := current * 10
		if x < len(sd.Steps)-1 {
			next = sd.Steps[x+1][0]
		}
		if next >= distance || x == len(sd.Steps)-1 {
			sd.GridLevel = float32(x) + (distance-current)/(next-current)
			return
		}
	}
	sd.GridLevel = 0
}

func (ShaderDataGrid) Size() int {
	return 0
}

func (sd *ShaderDataGrid) InstanceBoundDataSize() int {
	return int(unsafe.Sizeof(sd.GridFeatureFlags)) +
		int(unsafe.Sizeof(sd.GridLineCount)) +
		int(unsafe.Sizeof(sd.Steps))
}

func (sd *ShaderDataGrid) BoundDataPointer() unsafe.Pointer {
	return unsafe.Pointer(&sd.GridFeatureFlags)
}

func (s *ShaderDataGrid) UpdateBoundData() bool {
	return true
}
