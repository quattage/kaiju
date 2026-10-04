/******************************************************************************/
/* workspace.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_areas

import (
	"fmt"
	"log/slog"

	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
)

type WorkspaceManager struct {
	editor          EditorAreaInterface
	activeWorkspace string
	areaTypes       map[string]func() AreaType
	MainArea        *Area
	Overlays        []*Area
	StageView       editor_stage_view.StageView
}

func (wm *WorkspaceManager) Initialize(ed EditorAreaInterface) *Area {
	wm.editor = ed
	wm.activeWorkspace = ""
	klib.ScreenUnitsDPMM = func() float64 {
		return ed.Host().Window.DotsPerMillimeter()
	}
	klib.ScreenUnitsScaleFactor = func() float64 {
		return float64(ed.Settings().UIScale)
	}
	wm.Refresh(ed)
	area, err := wm.OpenOverlay("com.kaiju.area_splash_screen", -1, -1)
	if err != nil {
		panic("Failed to open splash screen! - " + err.Error())
	}
	wm.Update(0)
	wm.constructDefaultWorkspace()
	return area
}

// Refresh rebuilds only the necessary elements to bring the workspace
// up-to-date with the currently configured settings and theme
func (wm *WorkspaceManager) Refresh(editor EditorAreaInterface) {
	wm.Editor().SetMinimumWindowSize(-1, -1)
	editor.Host().Window.SetTitleBarColor(editor.Theme().ColorContextBar.AsColor8())
	editor.Theme().MarkGlobalCSSDirty()
	wm.finalizeATRegistry()
}

func (wm *WorkspaceManager) Editor() EditorAreaInterface {
	return wm.editor
}

// OpenOverlay creates a new Area of the specified type. If no type exists
// at the provided registry ID, an error will be returned.
// This Area is given overlay status and as such will float above other UI elements
// and capture focus. Position is relative to the top left corner of the overlay
// panel.
func (wm *WorkspaceManager) OpenOverlay(areaID string, posx, posy float32) (*Area, error) {
	if len(wm.areaTypes) <= 0 {
		panic("Attempted to modify workspace before initialization")
	}
	areaToOpen := wm.GetAreaType(areaID)
	if areaToOpen == nil {
		return nil, fmt.Errorf("No such area: '%s'", areaID)
	}
	if areaToOpen.Flags&AreaCapabilityFlagOverlayable == 0 {
		return nil, fmt.Errorf("%s isn't overlayable!", areaToOpen)
	}
	newArea := newBlankArea(areaToOpen)
	wx, wy := areaToOpen.HandlerFactory().GetOverlayDimensions()
	maxW := float32(wm.editor.Host().Window.Width())
	maxH := float32(wm.editor.Host().Window.Height())
	if wx < 800 && wy < 800 {
		wx = klib.MM2Pix(wx)
		wy = klib.MM2Pix(wy)
		wx = min(max(wx, 10), maxW-24)
		wy = min(max(wy, 10), maxH-24)
	}
	if posx < 0 {
		posx = (maxW - wx) / 2
	} else {
		posx = max(0, min(posx, maxW-wx))
	}
	if posy < 0 {
		posy = (maxH - wy) / 2
	} else {
		posy = max(0, min(posy, maxH-wy))
	}
	newArea.openWithSize(wm, wx, wy)
	newArea.Panel.Base().Layout().SetOffset(posx, posy)
	newArea.Panel.Base().Layout().SetZ(9)
	newArea.OverlayPX = (posx + wx*0.5) / maxW
	newArea.OverlayPY = (posy + wy*0.5) / maxH
	wm.Overlays = append(wm.Overlays, newArea)
	return newArea, nil
}

// Open creates, configures, and displays a new Area under the provided parent.
// If no AreaType exists at the provided registry ID, an error will be returned.
// The position of this window is calculated based on nesting.
func (wm *WorkspaceManager) Open(areaID string, parentArea *Area, direction SplitDirection, ratio float32) (*Area, error) {
	if len(wm.areaTypes) <= 0 {
		panic("Attempted to modify workspace before initialization")
	}
	areaToOpen := wm.GetAreaType(areaID)
	if areaToOpen == nil {
		return nil, fmt.Errorf("No such area: '%s'", areaID)
	}
	if parentArea == nil {
		if wm.MainArea == nil {
			wm.MainArea = newBlankArea(areaToOpen)
			slog.Info(fmt.Sprintf("Populated MainArea with %s", areaToOpen))
			wm.MainArea.openWithSize(wm, float32(wm.editor.Host().Window.Width()), float32(wm.editor.Host().Window.Height()))
			return wm.MainArea, nil
		}
		parentArea = wm.MainArea
	} else {
		if parentArea.Type.ID == AreaTypeComposite.ID {
			panic("invalid composite open request")
		}
		if wm.MainArea == nil {
			panic("no fallback parent available")
		}
	}
	if areaToOpen.Flags&AreaCapabilityFlagDockable == 0 {
		return nil, fmt.Errorf("%s isn't dockable!", areaToOpen)
	}
	areaCopy := *parentArea
	areaCopy.owner = parentArea
	areaCopy.SplitDirection = SplitHorizontal
	areaCopy.Ratio = -1
	newArea := &Area{
		Type:           *areaToOpen,
		owner:          parentArea,
		OverlayPX:      areaCopy.OverlayPX,
		OverlayPY:      areaCopy.OverlayPY,
		SplitDirection: SplitHorizontal,
		Ratio:          -1,
	}
	// the stacking order of the splits uses the sign of the ratio
	mar := wm.editor.Theme().SizeAreaMinimum
	if ratio <= 0 {
		parentArea.ChildA = &areaCopy
		parentArea.ChildB = newArea
		if ratio < -(1 - mar) {
			parentArea.Ratio = 1 - mar
		} else if ratio > -mar {
			parentArea.Ratio = mar
		} else {
			parentArea.Ratio = -ratio
		}
	} else {
		parentArea.ChildA = newArea
		parentArea.ChildB = &areaCopy
		if ratio > 1-mar {
			parentArea.Ratio = 1 - mar
		} else if ratio < mar {
			parentArea.Ratio = mar
		} else {
			parentArea.Ratio = ratio
		}
	}
	parentArea.manager = nil
	parentArea.Panel = nil
	parentArea.doc = nil
	parentArea.Type = AreaTypeComposite
	parentArea.SplitDirection = direction
	newArea.open(wm)
	return newArea, nil
}

func (wm *WorkspaceManager) Close(areaToClose *Area) {
	if len(wm.areaTypes) <= 0 {
		panic("Attempted to modify workspace before it finished loading")
	}
	if areaToClose == nil {
		slog.Error("Skipped attempt to close a nil area.")
	}
	if areaToClose.owner != nil {
		switch areaToClose {
		case areaToClose.owner.ChildA:
			areaToClose.owner.ChildA = nil
			return
		case areaToClose.owner.ChildB:
			areaToClose.owner.ChildB = nil
			return
		}
	}
	if areaToClose.Type.ID == AreaTypeComposite.ID {
		if areaToClose.ChildA != nil {
			wm.Close(areaToClose.ChildA)
		} else if areaToClose.ChildB != nil {
			wm.Close(areaToClose.ChildB)
		}
	}
	areaToClose.close(wm)
}

func (wm *WorkspaceManager) CloseOverlay(areaID string) {
	for _, area := range wm.Overlays {
		if area.Type.ID != areaID {
			continue
		}
		area.close(wm)
	}
}

func (wm *WorkspaceManager) FocusInterface() {
	if wm.MainArea == nil {
		return
	}
	wm.MainArea.PerformAsChildren(func(handler AreaHandler) {
		handler.FocusInterface(wm.editor)
	}, wm.editor)
}

func (wm *WorkspaceManager) BlurInterface() {
	if wm.MainArea == nil {
		return
	}
	wm.MainArea.PerformAsChildren(func(handler AreaHandler) {
		handler.BlurInterface(wm.editor)
	}, wm.editor)
}

func (wm *WorkspaceManager) Update(deltaTime float64) {
	defer tracing.NewRegion("WorkspaceManager.Update").End()
	maxW := float32(wm.editor.Host().Window.Width())
	maxH := float32(wm.editor.Host().Window.Height())
	if wm.MainArea != nil {
		wm.MainArea.Update(wm.editor, deltaTime, 0, 0, maxW, maxH)
	}
	for _, area := range wm.Overlays {
		areaW := area.Width()
		areaH := area.Height()
		posx := area.OverlayPX*maxW - areaW*0.5
		posy := area.OverlayPY*maxH - areaH*0.5
		area.Update(wm.editor, deltaTime, posx, posy, areaW, areaH)
	}
}

func (wm *WorkspaceManager) Serialize(ed EditorAreaInterface) {

}

func (wm *WorkspaceManager) constructDefaultWorkspace() {
	wm.Open("com.kaiju.area_stage_viewport", nil, SplitHorizontal, 0.5)
	p, _ := wm.Open("com.kaiju.area_content_browser", nil, SplitVertical, -0.5)
	wm.Open("com.kaiju.area_fallback", p, SplitHorizontal, 0.25)
}

// GetAreaType returns the AreaType mapped to the provide string ID, should one
// exist. Returns nil otherwise.
func (wm *WorkspaceManager) GetAreaType(id string) *AreaType {
	if len(id) <= 0 {
		return nil
	}
	areaType, ok := wm.areaTypes[id]
	if !ok {
		return nil
	}
	output := areaType()
	return &output
}

// finalizeATRegistry adds all deferred AreaType registry entries to the
// persistent map and closes the deferred registry.
func (wm *WorkspaceManager) finalizeATRegistry() {
	if len(deferredAreaTypeRegistry) <= 0 {
		return
	}
	if wm.areaTypes == nil {
		wm.areaTypes = make(map[string]func() AreaType, 32)
	}
	for _, factory := range deferredAreaTypeRegistry {
		areaType := factory()
		if !regex.MatchString(areaType.ID) {
			slog.Error(fmt.Sprintf("Skipped area registration at ID '%s' - This ID contains invalid characters! (expected %s)", areaType.ID, regex.String()))
			continue
		}
		if areaType.HandlerFactory == nil {
			slog.Error(fmt.Sprintf("Failed to register Area '%s' - Factory failed to supply a Handler!", areaType.ID))
			continue
		}
		_, exists := wm.areaTypes[areaType.ID]
		if exists {
			slog.Warn(fmt.Sprintf("Skipped redundant re-registration of AreaType at ID '%s'", areaType.ID))
			continue
		}
		slog.Debug(fmt.Sprintf("Registered %s", areaType.String()))
		wm.areaTypes[areaType.ID] = factory
	}
	deferredAreaTypeRegistry = nil
}
