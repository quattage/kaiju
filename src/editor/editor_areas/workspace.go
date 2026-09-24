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
	workspaces      map[string]WorkspaceConfiguration
	activeWorkspace string
	areaTypes       map[string]func() AreaType
	MainArea        *Area
	Overlays        []*Area
	StageView       editor_stage_view.StageView
}

func (wm *WorkspaceManager) Initialize(editor EditorAreaInterface) {
	wm.editor = editor
	wm.activeWorkspace = ""
	klib.ScreenUnitsDPMM = func() float64 {
		return editor.Host().Window.DotsPerMillimeter()
	}
	klib.ScreenUnitsScaleFactor = func() float64 {
		return float64(editor.Settings().UIScale)
	}
	wm.Refresh(editor)
	wm.Open("com.kaiju.area_primary", nil, SplitHorizontal, -1)
	wm.OpenOverlay("com.kaiju.area_splash_screen", -1, -1)
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
		panic("Attempted to modify workspace before it finished loading")
	}
	areaToOpen := wm.GetAreaType(areaID)
	if areaToOpen == nil {
		return nil, fmt.Errorf("No such area: '%s'", areaID)
	}
	if areaToOpen.Flags&AreaCapabilityFlagOverlayable == 0 {
		return nil, fmt.Errorf("%s isn't overlayable!", areaToOpen)
	}
	newArea := newBlankArea(areaToOpen)
	wx, wy := areaToOpen.Handler.GetOverlayDimensions()
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
	newArea.openWithSize(wm, wx, wy, wm.editor.Theme().SizeRadiusGlobal)
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
		panic("Attempted to modify workspace before it finished loading")
	}
	areaToOpen := wm.GetAreaType(areaID)
	if areaToOpen == nil {
		return nil, fmt.Errorf("No such area: '%s'", areaID)
	}
	if parentArea == nil {
		if wm.MainArea == nil {
			wm.MainArea = newBlankArea(areaToOpen)
			slog.Info(fmt.Sprintf("Populated MainArea with %s", areaToOpen))
			wm.MainArea.openWithSize(wm, float32(wm.editor.Host().Window.Width()), float32(wm.editor.Host().Window.Height()), 0)
			return wm.MainArea, nil
		}
		parentArea = wm.MainArea
	}
	if areaToOpen.Flags&AreaCapabilityFlagDockable == 0 {
		return nil, fmt.Errorf("%s isn't dockable!", areaToOpen)
	}
	if parentArea.Type.ID == AreaTypeComposite.ID {
		// under normal circumstances, this shouldn't be hit.
		// the recursive call here does mean that the area is looked up redundantly
		// but since this should be pretty rare it's probably fine. This will only
		// ever be an issue for areas that are nested very deeply.
		if parentArea.ChildA != nil {
			wm.Open(areaID, parentArea.ChildA, direction, ratio)
		} else if parentArea.ChildB != nil {
			wm.Open(areaID, parentArea.ChildB, direction, ratio)
		} else {
			return nil, fmt.Errorf("%s couldn't be parented - The target parent %s is a composite, but has no children! This indicates that the workspace area tree has been invalidated!", areaToOpen, parentArea)
		}
		return nil, nil
	}
	areaCopy := parentArea
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
		parentArea.ChildA = areaCopy
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
		parentArea.ChildB = areaCopy
		if ratio > 1-mar {
			parentArea.Ratio = 1 - mar
		} else if ratio < mar {
			parentArea.Ratio = mar
		} else {
			parentArea.Ratio = ratio
		}
	}
	parentArea.Manager = nil
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

// add all deferred registry entries to the persistent registry
// and close the deferred registry.
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
		if areaType.Handler == nil {
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

// // A quick debug helper that places a small floating square above all UI elements
// // at the provided position.
// func (wm *WorkspaceManager) RevealPosition(at func() matrix.Vec2) {
// 	if wm.debugManager == nil {
// 		wm.debugManager = &ui.Manager{}
// 		wm.debugManager.Init(wm.editor.Host())
// 	}
// 	ds := DebugSquare{
// 		posGetter: at,
// 		panel:     wm.debugManager.Add().ToPanel(),
// 		size:      8,
// 	}
// 	ds.panel.Init(nil, ui.ElementTypePanel)
// 	ds.Update()
// 	wm.debugs = append(wm.debugs, &ds)
// }

// GetWorkspace returns the WorkspaceConfiguration mapped to the provided string
// ID, should one exist. Returns nil otherwise.
func (wm *WorkspaceManager) GetWorkspace(id string) *WorkspaceConfiguration {
	if len(id) <= 0 {
		return nil
	}
	config, ok := wm.workspaces[id]
	if !ok {
		return nil
	}
	return &config
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

// SwitchToWorkspace performs lifecycle operations necessary to transfer the current
// configuration of Areas to the one described by the WorkspaceConfiguration at the provided
// index. It is primarily intended to be called by EditorServices.SwitchToWorkspace
func (wm *WorkspaceManager) SwitchToWorkspace(id string) {
	if len(id) <= 0 {
		return
	}
	if wm.MainArea != nil && len(wm.activeWorkspace) <= 0 {
		wm.Close(wm.MainArea)
	}
	config := wm.GetWorkspace(id)
	if config == nil {
		slog.Error(fmt.Sprintf("Couldn't switch to workspace '%s' - No configuration at this index exists!", id))
		return
	}
	wm.activeWorkspace = id
	wm.LoadWorkspace(config)
}
