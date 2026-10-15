package viewport

import "strings"

type ViewportCameraMode int

const (
	ViewportCameraModeNone = ViewportCameraMode(iota)
	ViewportCameraMode3d
	ViewportCameraMode2d
	ViewportCameraModeTop
	ViewportCameraModeFront
	ViewportCameraModeSide
	ViewportCameraModeLeft
	ViewportCameraModeRight
)

var ViewportCameraModes = []ViewportCameraMode{
	ViewportCameraModeNone,
	ViewportCameraMode3d,
	ViewportCameraMode2d,
	ViewportCameraModeTop,
	ViewportCameraModeFront,
	ViewportCameraModeSide,
	ViewportCameraModeLeft,
	ViewportCameraModeRight,
}

func (v ViewportCameraMode) String() string {
	vi := int(v)
	if vi < 0 || vi >= len(cameraModeStrings) {
		return cameraModeStrings[ViewportCameraModeNone]
	}
	return cameraModeStrings[v]
}

func (v ViewportCameraMode) viewMode() StageViewMode {
	switch v {
	case ViewportCameraMode3d:
		return StageViewportPerspective
	case ViewportCameraMode2d:
		return StageViewportFront
	case ViewportCameraModeTop:
		return StageViewportTop
	case ViewportCameraModeFront:
		return StageViewportFront
	case ViewportCameraModeSide:
		return StageViewportSide
	case ViewportCameraModeLeft:
		return StageViewportLeft
	case ViewportCameraModeRight:
		return StageViewportRight
	default:
		return StageViewportPerspective
	}
}

type StageViewMode int

const (
	StageViewportPerspective StageViewMode = iota
	StageViewportTop
	StageViewportSide
	StageViewportFront
	StageViewportLeft  = StageViewportSide
	StageViewportRight = StageViewportFront
)

var stageViewModes = []StageViewMode{
	StageViewportPerspective,
	StageViewportTop,
	StageViewportSide,
	StageViewportFront,
}

var cameraModeStrings = []string{"None", "3D", "2D", "Top", "Front", "Side", "Left", "Right"}

func (k StageViewMode) String() string {
	if k <= 0 {
		return "Viewport"
	}
	return k.cameraMode().String()
}

// TODO this wont return stage_main but instead returns stage_perspective
func (k StageViewMode) RenderTargetName() string {
	fn := strings.ToLower(k.String())
	return "stage_" + fn
}

func (k StageViewMode) cameraMode() ViewportCameraMode {
	switch k {
	case StageViewportPerspective:
		return ViewportCameraMode3d
	case StageViewportTop:
		return ViewportCameraModeTop
	case StageViewportSide:
		return ViewportCameraModeSide
	case StageViewportFront:
		return ViewportCameraModeFront
	default:
		return ViewportCameraMode3d
	}
}
