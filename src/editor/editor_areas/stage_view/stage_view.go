/******************************************************************************/
/* stage_viewport.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_view

import (
	"log/slog"
	"math"

	"kaijuengine.com/editor/editor_areas"
	"kaijuengine.com/editor/editor_areas/stage_view/viewport"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"

	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

var StageViewArea = editor_areas.AreaType{
	Key:            klib.NewKey("com.kaiju.area_stage_view", "Stage Viewport"),
	HandlerFactory: func() editor_areas.AreaHandlerInterface { return &StageViewHandler{} },
	Flags:          editor_areas.AreaCapabilityFlagDockable | editor_areas.AreaCapabilityFlagDiscoverable,
}

func init() {
	editor_areas.Register(StageViewArea)
}

type StageViewHandler struct {
	Mode        viewport.StageViewMode
	Label       string
	defaultView struct {
		options rendering.RenderViewOptions
		active  bool
	}

	camera       viewport.ViewportCamera
	bounds       viewport.ViewportRect
	renderTarget *rendering.RenderTarget
	renderView   *rendering.RenderView
	overlays     rendering.Drawings

	// transformMan  editor_stage_view.TransformationManager
	// transformTool editor_tooling.TransformTool
	// vertexSnap    editor_stage_view.VertexSnapTool
	// selectTool    editor_tooling.SelectTool
	// stagePicking  editor_stage_view.StagePicking

	gridShader *shader_data_registry.ShaderDataGrid
}

func (h *StageViewHandler) GetViewportCamera() *viewport.ViewportCamera {
	return &h.camera
}

func (h *StageViewHandler) Open(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
	h.overlays = rendering.NewDrawings()
	h.setupGrid(ed)
	h.setupCamera(ed)
}

func (h *StageViewHandler) Close(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
	h.overlays.Clear()
	h.overlays.Destroy(ed.Host().Window.GpuInstance.PrimaryDevice())
}

func (h *StageViewHandler) FocusInterface(ed editor_areas.EditorAreaInterface) {
}

func (h *StageViewHandler) BlurInterface(ed editor_areas.EditorAreaInterface) {
}

func (h *StageViewHandler) Update(area *editor_areas.Area, ed editor_areas.EditorAreaInterface, deltaTime float64, posx, posy, width, height float32) {
	h.camera.Update(ed.Host(), deltaTime)
	h.setupRenderTarget(ed, width, height)
	h.syncViewportRect(ed, area, posx, posy, width, height)
}

func (h *StageViewHandler) GetOverlayDimensions() (float32, float32) {
	return -1, -1
}

func (h *StageViewHandler) syncViewportRect(ed editor_areas.EditorAreaInterface, area *editor_areas.Area, posx, posy, width, height float32) {
	defer tracing.NewRegion("StageViewHandler.syncViewportRect").End()
	if cam := h.camera.Camera(); cam != nil {
		if !h.camera.IsBoundsApproximately(width, height) {
			h.camera.SetViewportBounds(posx, posy, width, height)
			cam.ViewportChanged(width, height)
			if h.renderView != nil {
				h.renderView.SetCamera(cam)
			}
		}
		h.syncGridShaderData(cam)
	}
	if h.renderTarget == nil {
		return
	}
	if h.renderTarget.Resize(max(1, int(math.Ceil(float64(width)))), max(1, int(math.Ceil(float64(height))))) {
		area.SetBackdropTexture(ed, nil)
		return
	} else if h.renderTarget.ResizeDirty() {
		return
	}
	tex, err := h.renderTarget.Texture(rendering.RenderTargetOutputColor)
	if err != nil {
		slog.Error("Failed to load stage view render target", "error", err)
		return
	}
	if tex != nil && tex != area.Panel.Background() {
		area.SetBackdropTexture(ed, tex)
	}
}

func (h *StageViewHandler) syncGridShaderData(cam cameras.Camera) {
	if h.gridShader == nil {
		return
	}
	var dist float32
	if cam.IsOrthographic() {
		proj := cam.Projection()
		h.gridShader.ClipRect = matrix.Vec4{
			4.0 / proj.At(0, 0),
			4.0 / proj.At(1, 1),
			0, 0,
		}
		dist = 3
	} else {
		h.gridShader.FarClip = cam.FarPlane()
		h.gridShader.ClipRect = matrix.Vec4{h.gridShader.FarClip, h.gridShader.FarClip, 0, 0}
		height := klib.FAbs(cam.Position().Y())
		forward := klib.FAbs(cam.Forward().Y())
		intersect := height / max(forward, 1e-6)
		dist = matrix.Lerp(intersect, height, 1-forward)
	}
	h.gridShader.CalculateLevel(dist)
}

func (h *StageViewHandler) setupCamera(ed editor_areas.EditorAreaInterface) {
	defer tracing.NewRegion("StageViewHandler.setupCamera").End()
	h.camera.SetMode(viewport.ViewportCameraMode3d, ed.Host())
	h.camera.Settings = &ed.Settings().EditorCamera
	h.camera.UseAsPrimary(ed.Host())
}

func (h *StageViewHandler) setupGrid(ed editor_areas.EditorAreaInterface) {
	defer tracing.NewRegion("StageViewHandler.setupGrid").End()
	material, err := ed.Host().MaterialCache().Material(assets.MaterialDefinitionGrid)
	if err != nil {
		slog.Error("failed to load the grid material", "error", err)
		return
	}
	sd := shader_data_registry.Create("grid")
	h.gridShader = sd.(*shader_data_registry.ShaderDataGrid)
	h.gridShader.SetDefaults()
	grid := rendering.NewEmptyMesh(ed.Host().MeshCache(), "viewport_grid", h.gridShader.VertexCount(), true)
	ed.Host().Drawings.AddDrawing(rendering.Drawing{
		Material:   material,
		Mesh:       grid,
		ShaderData: h.gridShader,
		Transform:  nil,
		Layer:      rendering.RenderLayerEditor,
		ViewCuller: nil,
	})
}

func (h *StageViewHandler) setupRenderTarget(ed editor_areas.EditorAreaInterface, width, height float32) {
	defer tracing.NewRegion("StageViewHandler.setupRenderTarget").End()
	name := h.Mode.RenderTargetName()
	if h.renderTarget == nil {
		if target, ok := ed.Host().RenderTargets.Target(name); ok {
			h.renderTarget = target
		} else {
			target, err := ed.Host().RenderTargets.Create(rendering.RenderTargetOptions{
				Name:   name,
				Width:  max(1, int(math.Ceil(float64(width)))),
				Height: max(1, int(math.Ceil(float64(height)))),
				Depth:  true,
			})
			if err != nil {
				slog.Error("failed to create stage render target", "error", err)
			} else {
				h.renderTarget = target
			}
		}
	}
	if h.renderView == nil {
		if view, ok := ed.Host().RenderViews.View(name); ok {
			h.renderView = view
		} else if h.renderTarget != nil {
			view, err := ed.Host().RenderViews.Create(rendering.RenderViewOptions{
				Name:      name,
				Target:    h.renderTarget,
				Camera:    h.camera.Camera(),
				LayerMask: rendering.RenderLayerWorld | rendering.RenderLayerEditor,
				Clear:     true,
				Sort:      -100,
				ViewMode:  rendering.RenderViewModeNormal,
			})
			if err != nil {
				slog.Error("failed to create stage render view", "error", err)
			} else {
				h.renderView = view
			}
		}
	}
}
