/******************************************************************************/
/* transform_gizmo.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_areas

import (
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/matrix"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
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

type axisWidget struct {
	hitbox     graviton.AABB
	shaderData rendering.DrawInstance
	pickData   rendering.DrawInstance
}

type Gizmo struct {
	axes        GizmoAxis
	widgets     [3]axisWidget
	drag        uint8
	size        float32
	OnDragStart events.EventWithArg[matrix.Vec3]
	OnDragEnd   events.EventWithArg[matrix.Vec3]
}

func (g *Gizmo) SetSize(size float32) {
	g.size = size
}

func (g *Gizmo) GetSize() float32 {
	return g.size
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
		if tg.widgets[0].shaderData != nil {
			tg.widgets[0].shaderData.Activate()
		}
	} else {
		tg.axes &^= GizmoAxisX
		if tg.widgets[0].shaderData != nil {
			tg.widgets[0].shaderData.Deactivate()
		}
	}
}

func (tg *Gizmo) SetY(enabled bool) {
	if enabled {
		tg.axes |= GizmoAxisY
		if tg.widgets[1].shaderData != nil {
			tg.widgets[1].shaderData.Activate()
		}
	} else {
		tg.axes &^= GizmoAxisY
		if tg.widgets[1].shaderData != nil {
			tg.widgets[1].shaderData.Activate()
		}
	}
}

func (tg *Gizmo) SetZ(enabled bool) {
	if enabled {
		tg.axes |= GizmoAxisZ
		if tg.widgets[2].shaderData != nil {
			tg.widgets[2].shaderData.Activate()
		}
	} else {
		tg.axes &^= GizmoAxisZ
		if tg.widgets[2].shaderData != nil {
			tg.widgets[2].shaderData.Activate()
		}
	}
}

func (tg *Gizmo) DrawTranslationArrows(transform *matrix.Transform, ed EditorAreaInterface) {
	// TODO move axis colors to the theme
	mat, _ := ed.Host().MaterialCache().Material("gizmo_overlay.material")
	if tg.axes <= 0 {
		return
	}
	mesh := rendering.NewMeshArrow(ed.Host().MeshCache(), tg.size, tg.size*0.1, tg.size*0.4, tg.size*0.3, 8)
	if tg.axes&GizmoAxisX != 0 {
		// TODO rotate the mesh to match the axis
		tg.widgets[0].shaderData = shader_data_registry.Create("unlit")
		ssd := tg.widgets[0].shaderData.(*shader_data_registry.ShaderDataUnlit)
		ssd.Color = matrix.ColorRed()
		draw := rendering.Drawing{
			Material:   mat,
			Mesh:       mesh,
			ShaderData: tg.widgets[0].shaderData,
			Transform:  transform,
			Layer:      rendering.RenderLayerEditor,
			ViewCuller: &ed.Host().Cameras.Primary,
		}
		ed.Host().Drawings.AddDrawing(draw)
	}
	if tg.axes&GizmoAxisY != 0 {
		tg.widgets[1].shaderData = shader_data_registry.Create("unlit")
		ssd := tg.widgets[1].shaderData.(*shader_data_registry.ShaderDataUnlit)
		ssd.Color = matrix.ColorGreen()
		draw := rendering.Drawing{
			Material:   mat,
			Mesh:       mesh,
			ShaderData: tg.widgets[1].shaderData,
			Transform:  transform,
			Layer:      rendering.RenderLayerEditor,
			ViewCuller: &ed.Host().Cameras.Primary,
		}
		ed.Host().Drawings.AddDrawing(draw)
	}
	if tg.axes&GizmoAxisZ != 0 {
		tg.widgets[2].shaderData = shader_data_registry.Create("unlit")
		ssd := tg.widgets[2].shaderData.(*shader_data_registry.ShaderDataUnlit)
		ssd.Color = matrix.ColorBlue()
		draw := rendering.Drawing{
			Material:   mat,
			Mesh:       mesh,
			ShaderData: tg.widgets[2].shaderData,
			Transform:  transform,
			Layer:      rendering.RenderLayerEditor,
			ViewCuller: &ed.Host().Cameras.Primary,
		}
		ed.Host().Drawings.AddDrawing(draw)
	}
}

func (tg *Gizmo) TranslationMesh(ed EditorAreaInterface) *rendering.Mesh {
	const key = "_editor_translation_gizmo_planes"
	mc := ed.Host().MeshCache()
	if mesh, ok := mc.FindMesh(key); ok {
		return mesh
	}
	// TODO seperate by axis
	// TODO settings for gizmo scale
	verts := make([]rendering.Vertex, 4)
	verts[0].Position = matrix.Vec3{-1.0, -1.0, 0.0}
	verts[0].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[0].UV0 = matrix.Vec2{0.0, 1.0}
	verts[0].Color = matrix.ColorWhite()
	verts[1].Position = matrix.Vec3{-1.0, 1.0, 0.0}
	verts[1].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[1].UV0 = matrix.Vec2{0.0, 0.0}
	verts[1].Color = matrix.ColorWhite()
	verts[2].Position = matrix.Vec3{1.0, 1.0, 0.0}
	verts[2].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[2].UV0 = matrix.Vec2{1.0, 0.0}
	verts[2].Color = matrix.ColorWhite()
	verts[3].Position = matrix.Vec3{1.0, -1.0, 0.0}
	verts[3].Normal = matrix.Vec3{0.0, 0.0, 1.0}
	verts[3].UV0 = matrix.Vec2{1.0, 1.0}
	verts[3].Color = matrix.ColorWhite()
	indexes := []uint32{
		0, 2, 1, 0, 3, 2,
		0, 1, 2, 0, 2, 3,
	}
	return mc.Mesh(key, verts, indexes)
}

func (tg *Gizmo) RotationMesh(ed EditorAreaInterface) *rendering.Mesh {
	const key = "_editor_rotation_gizmo_circles"
	mc := ed.Host().MeshCache()
	if mesh, ok := mc.FindMesh(key); ok {
		return mesh
	}
	// TODO settings for gizmo scale
	return rendering.NewMeshCircleWire(mc, tg.size, 32)
}

func (tg *Gizmo) ScaleMesh(ed EditorAreaInterface) *rendering.Mesh {
	const key = "_editor_scale_gizmo_handles"
	mc := ed.Host().MeshCache()
	if mesh, ok := mc.FindMesh(key); ok {
		return mesh
	}
	return rendering.NewMeshHandle(mc, 2, tg.size*0.1, tg.size*0.5, 5)
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
