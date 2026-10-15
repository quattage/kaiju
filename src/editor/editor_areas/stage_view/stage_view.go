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
	"kaijuengine.com/engine"
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
	texture      *rendering.Texture

	host *engine.Host
	area *editor_areas.Area

	// transformMan  editor_stage_view.TransformationManager
	// transformTool editor_tooling.TransformTool
	// vertexSnap    editor_stage_view.VertexSnapTool
	// selectTool    editor_tooling.SelectTool
	// stagePicking  editor_stage_view.StagePicking

	grid struct {
		transform matrix.Transform
		shader    *shader_data_registry.ShaderDataGrid
		enabled   bool
	}
}

func (h *StageViewHandler) GetViewportCamera() *viewport.ViewportCamera {
	if h != nil && h.host != nil && h.area != nil {
		return &h.camera
	}
	return nil
}

func (h *StageViewHandler) Open(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
	h.host = ed.Host()
	h.area = area
	area.AnnotateCenter(ed, "stage")
}

func (h *StageViewHandler) Close(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
	h.host = nil
	h.area = nil

}

func (h *StageViewHandler) FocusInterface(ed editor_areas.EditorAreaInterface) {
}

func (h *StageViewHandler) BlurInterface(ed editor_areas.EditorAreaInterface) {
}

func (h *StageViewHandler) Update(area *editor_areas.Area, ed editor_areas.EditorAreaInterface, deltaTime float64, posx, posy, width, height float32) {
}

func (h *StageViewHandler) GetOverlayDimensions() (float32, float32) {
	return -1, -1
}

// THIS PARTY'S GOIN' CRAZY SEND THE INVOICE TO GOD
func (h *StageViewHandler) setupRenderTarget() {
	defer tracing.NewRegion("StageViewHandler.setupRenderTarget").End()
	if h.host == nil {
		return
	}
	bounds := h.area.GetRect()
	width, height := stageViewportTargetSize(bounds.Size())
	name := h.Mode.RenderTargetName()
	if h.renderTarget == nil {
		if target, ok := h.host.RenderTargets.Target(name); ok {
			h.renderTarget = target
		} else {
			target, err := h.host.RenderTargets.Create(rendering.RenderTargetOptions{
				Name:   name,
				Width:  width,
				Height: height,
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
		if view, ok := h.host.RenderViews.View(name); ok {
			h.renderView = view
		} else if h.renderTarget != nil {
			view, err := h.host.RenderViews.Create(rendering.RenderViewOptions{
				Name:      name,
				Target:    h.renderTarget,
				Camera:    h.camera.Camera(),
				LayerMask: rendering.RenderLayerWorld | rendering.RenderLayerEditor,
				Clear:     true,
				Sort:      -100,
			})
			if err != nil {
				slog.Error("failed to create stage render view", "error", err)
			} else {
				h.renderView = view
			}
		}
	}
}

func (h *StageViewHandler) resizeStageTargetToViewport(size matrix.Vec2) bool {
	if h.renderTarget == nil {
		return false
	}
	width, height := stageViewportTargetSize(size)
	return h.renderTarget.Resize(width, height)
}

func (h *StageViewHandler) takedownRenderTarget() {
	defer tracing.NewRegion("StageViewHandler.takedownRenderTarget").End()
	if h.host == nil {
		return
	}
	name := h.Mode.RenderTargetName()
	if h.renderView == nil {
		if view, ok := h.host.RenderViews.View(name); ok {
			h.renderView = view
		} else {
			return
		}
	}
	if err := h.host.RenderViews.Destroy(name); err != nil {
		slog.Error("Failed to destroy hidden stage render view", "view", name, "error", err)
	}
	h.renderView = nil
}

func stageViewportTargetSize(size matrix.Vec2) (int, int) {
	width := int(math.Ceil(float64(size.X())))
	height := int(math.Ceil(float64(size.Y())))
	return max(1, width), max(1, height)
}
