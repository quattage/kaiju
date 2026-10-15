/******************************************************************************/
/* workspace.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_areas

import (
	"fmt"
	"log/slog"

	"kaijuengine.com/editor/editor_active_stage"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

var areaTypes klib.Registry[AreaType] = klib.NewRegistry[AreaType](
	"com.kaiju.reg_area_type", "AreaTypes registry")

type WorkspaceManager struct {
	editor          EditorAreaInterface
	activeWorkspace string
	stage           editor_active_stage.ActiveStage
	main            *Area
	overlays        []*Area
	focused         *Area
	previousFocused *Area
	nearbyGutter    Gutter
	lazy            bool
}

func Register(at AreaType) {
	slog.Info("Registered " + at.Key.String())
	areaTypes.AddUnique(at)
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
	ed.Host().SetSwapChainClearColor(ed.Theme().ColorBackground)
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
	areaToOpen, ok := areaTypes.GetByString(areaID)
	if !ok {
		return nil, fmt.Errorf("No such area: '%s'", areaID)
	}
	if areaToOpen.Flags&AreaCapabilityFlagOverlayable == 0 {
		return nil, fmt.Errorf("%s isn't overlayable!", areaToOpen.String())
	}
	newArea := newBlankArea(&areaToOpen)
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
	newArea.open(wm, wx, wy)
	newArea.SetBackdropColor(matrix.ColorMix(wm.Editor().Theme().ColorPanel, matrix.ColorWhite(), 0.008))
	newArea.Panel.Base().Layout().SetOffset(posx, posy)
	newArea.Panel.Base().Layout().SetZ(9)
	newArea.overlayPX = (posx + wx*0.5) / maxW
	newArea.overlayPY = (posy + wy*0.5) / maxH
	wm.overlays = append(wm.overlays, newArea)
	return newArea, nil
}

// Open creates, configures, and displays a new Area under the provided parent.
// If no AreaType exists at the provided registry ID, an error will be returned.
// The position of this window is calculated based on nesting.
func (wm *WorkspaceManager) Open(areaID string, parentArea *Area, direction SplitDirection, ratio float32) (*Area, error) {
	areaToOpen, ok := areaTypes.GetByString(areaID)
	if !ok {
		return nil, fmt.Errorf("No such area: '%s'", areaID)
	}
	if parentArea == nil {
		if wm.main == nil {
			wm.main = newBlankArea(&areaToOpen)
			slog.Info(fmt.Sprintf("Populated MainArea with %s", areaToOpen.String()))
			wm.main.open(wm, float32(wm.editor.Host().Window.Width()), float32(wm.editor.Host().Window.Height()))
			return wm.main, nil
		}
		parentArea = wm.main
	} else {
		if parentArea.childA != nil {
			parentArea = parentArea.childA
		} else if parentArea.childB != nil {
			parentArea = parentArea.childB
		}
	}
	if areaToOpen.Flags&AreaCapabilityFlagDockable == 0 {
		return nil, fmt.Errorf("%s isn't dockable!", areaToOpen.String())
	}
	areaCopy := *parentArea
	areaCopy.owner = parentArea
	areaCopy.splitDirection = SplitNone
	areaCopy.ratio = -1
	newArea := &Area{
		areaType:       &areaToOpen,
		owner:          parentArea,
		overlayPX:      areaCopy.overlayPX,
		overlayPY:      areaCopy.overlayPY,
		splitDirection: SplitNone,
		ratio:          -1,
		focused:        false,
	}
	parentArea.resetFields()
	parentArea.splitDirection = direction
	// the stacking order of the splits uses the sign of the ratio
	if ratio <= 0 {
		parentArea.childA = &areaCopy
		parentArea.childB = newArea
		parentArea.ratio = -ratio
	} else {
		parentArea.childA = newArea
		parentArea.childB = &areaCopy
		parentArea.ratio = ratio
	}
	newArea.open(wm, 10, 10)
	return newArea, nil
}

func (wm *WorkspaceManager) Close(areaToClose *Area) {
	// TODO impl
}

func (wm *WorkspaceManager) CloseOverlayByID(id string) {
	for _, area := range wm.overlays {
		if area.areaType != nil && area.areaType.Key.ID() != id {
			continue
		}
		area.closeAndInvalidate(wm.editor)
	}
}

func (wm *WorkspaceManager) CloseOverlayByKey(key klib.RegistryKey) {
	wm.CloseOverlayByID(key.ID())
}

func (wm *WorkspaceManager) CloseOverlay(at *AreaType) {
	for _, area := range wm.overlays {
		if area.areaType != at {
			continue
		}
		area.closeAndInvalidate(wm.editor)
	}
}

func (wm *WorkspaceManager) Update(deltaTime float64) {
	defer tracing.NewRegion("WorkspaceManager.Update").End()
	maxW := float32(wm.editor.Host().Window.Width())
	maxH := float32(wm.editor.Host().Window.Height())
	nearbyGutters := []Gutter{}
	gutterThick := (wm.Editor().Theme().SizeTrenchGlobal *
		float32(wm.Editor().Host().Window.DotsPerMillimeter()) * 0.4) / 2
	if wm.main != nil {
		wm.main.updateArea(wm, deltaTime, 0, 0, maxW, maxH, gutterThick, &nearbyGutters)
	}
	for _, area := range wm.overlays {
		areaW := area.GetWidth()
		areaH := area.GetHeight()
		posx := area.overlayPX*maxW - areaW*0.5
		posy := area.overlayPY*maxH - areaH*0.5
		area.updateArea(wm, deltaTime, posx, posy, areaW, areaH, gutterThick, nil)
	}
	wm.updateMouseManipulation(nearbyGutters, gutterThick)
}

func (a *Area) updateArea(wm *WorkspaceManager, deltaTime float64, posx, posy, width, height, gutterThick float32, nearby *[]Gutter) {
	if a.IsFunctional() {
		if a.IsOverlay() || a.areaType.Flags == 0 || a.owner == wm.main {
			a.Panel.Base().Layout().SetOffset(posx, posy)
			a.Panel.Base().Layout().Scale(width, height)
		} else {
			a.Panel.Base().Layout().SetOffset(posx+gutterThick, posy+gutterThick)
			a.Panel.Base().Layout().Scale(width-gutterThick, height-gutterThick)
		}
		a.GetHandler().Update(a, wm.Editor(), deltaTime, posx, posy, width, height)
	} else if a.IsComposite() {
		if a.childA == a || a.childB == a {
			// A composite area parented to itself is not allowed and should
			// never occur.
			panic("cycle in area hierarchy")
		}
		pp := wm.Editor().Host().Window.Cursor.ScreenPosition()
		switch a.splitDirection {
		case SplitHorizontal:
			split := width * a.ratio
			wm.constrainArea(a)
			if pp.Y() > posy+gutterThick && pp.Y() < posy+height-gutterThick {
				dist := klib.FAbs(pp.X() - (posx + split + gutterThick/2))
				if dist < gutterThick*2 {
					*nearby = append(*nearby, Gutter{a, posx + split - gutterThick/2, posy, gutterThick * 2, height, dist, false})
				}
			}
			// We recurse into children. The cumulative position and size here
			// are computed based on the area's Ratio and never stored.
			a.childA.updateArea(wm, deltaTime, posx, posy, split, height, gutterThick, nearby)
			a.childB.updateArea(wm, deltaTime, split, posy, width-split, height, gutterThick, nearby)
		case SplitVertical:
			split := height * a.ratio
			wm.constrainArea(a)
			if pp.X() > posx+gutterThick && pp.X() < posx+width-gutterThick {
				dist := klib.FAbs(pp.Y() - (posy + split + gutterThick/2))
				if dist < gutterThick*2 {
					*nearby = append(*nearby, Gutter{a, posx, posy + split - gutterThick/2, width, gutterThick * 2, dist, false})
				}
			}
			// We recurse into children. The cumulative position and size here
			// are computed based on the area's Ratio and never stored.
			a.childA.updateArea(wm, deltaTime, posx, posy, width, split, gutterThick, nearby)
			a.childB.updateArea(wm, deltaTime, posx, split, width, height-split, gutterThick, nearby)
		}
	} else if !wm.IsGrabbing() {
		// An area that is owned and has no children is guaranteed to be the
		// deepest for this branch, so we manage focus on it by detecting
		// whether or not the mouse cursor is within it
		mp := wm.Editor().Host().Window.Cursor.ScreenPosition()
		if mp.X() >= posx && mp.Y() >= posy && mp.X() <= posx+width && mp.Y() <= posy+height {
			// we maintain the old focus and current focus as instance varaibles
			// for use with other editor/ui stuff later
			wm.previousFocused = wm.focused
			wm.focused = a
			if wm.previousFocused != nil {
				wm.previousFocused.focused = false
				wm.previousFocused.blurInterface(wm)
			}
			if wm.focused != nil {
				wm.focused.focused = true
				wm.focused.focusInterface(wm)
			}
		}
	}
}

func (wm *WorkspaceManager) updateMouseManipulation(gutters []Gutter, gutterThick float32) {
	if wm.IsFocusedOnOverlay() {
		if wm.nearbyGutter.grabbed {
			wm.setGutter(Gutter{})
		}
		return
	}
	if !wm.selectClosestGutter(gutters, gutterThick) {
		return
	}
	cursor := wm.editor.Host().Window.Cursor
	if !cursor.Held() {
		wm.nearbyGutter.grabbed = false
		return
	}
	wm.nearbyGutter.grabbed = true
	if wm.focused != nil {
		wm.focused.blurInterface(wm)
		wm.focused = nil
	}
	var pd float32
	switch wm.nearbyGutter.composite.splitDirection {
	case SplitHorizontal:
		pd = -cursor.Delta().X() / wm.nearbyGutter.composite.GetWidth()
	case SplitVertical:
		pd = cursor.Delta().Y() / wm.nearbyGutter.composite.GetHeight()
	}
	wm.nearbyGutter.composite.ratio += pd
	wm.constrainArea(wm.nearbyGutter.composite)
	wm.nearbyGutter.composite.InvalidateRect()
	if !wm.lazy {
		// due to the way that the UI layout system propagates updates, calling
		// this immediately significantly improves the responsiveness of Area
		// resizing, but this is mostly cosmetic.
		wm.editor.Host().UIUpdater.Update(0)
	}
}

func (wm *WorkspaceManager) selectClosestGutter(gutters []Gutter, gutterThick float32) bool {
	if wm.nearbyGutter.grabbed {
		return true
	}
	if len(gutters) <= 0 {
		wm.setGutter(Gutter{})
		return false
	}
	closest := gutters[0]
	for _, gutter := range gutters[1:] {
		if gutter.dist < wm.nearbyGutter.dist {
			closest = gutter
		}
	}
	if closest.dist > gutterThick {
		wm.setGutter(Gutter{})
		return false
	} else {
		wm.setGutter(closest)
	}
	return true
}

// Moves and resizes areas depending on context to ensure that UI elements can
// never be made so small as to be uninteractible.
func (wm *WorkspaceManager) constrainArea(area *Area) {
	if area.IsOverlay() {
		// TODO deal with overlays by clamping them inside the window bounds
		return
	}
	if !area.IsComposite() {
		// recurse upwards
		wm.constrainArea(area.owner)
		return
	}
	minPixels := minFunctionalWidth * float32(wm.editor.Host().Window.DotsPerMillimeter())
	var totalPixels float32
	switch area.splitDirection {
	case SplitHorizontal:
		totalPixels = max(2, area.GetWidth())
	case SplitVertical:
		totalPixels = max(2, area.GetHeight())
	}
	minRatio := minPixels / totalPixels
	newRatio := max(minRatio, min(1.0-minRatio, area.ratio))
	if matrix.ApproxTo(area.ratio, newRatio, 0.0001) {
		return
	}
	area.ratio = newRatio
}

func (a *Area) focusInterface(wm *WorkspaceManager) {
	if a.Panel != nil && a.manager != nil && !a.IsOverlay() {
		a.Highlight(wm.editor)
	}
	if a.internalHandler != nil {
		a.internalHandler.FocusInterface(wm.Editor())
	}
}

func (a *Area) blurInterface(wm *WorkspaceManager) {
	if a.Panel != nil && a.manager != nil {
		a.Unhighlight(wm.editor)
	}
	if a.internalHandler != nil {
		a.internalHandler.BlurInterface(wm.Editor())
	}
}

func (wm *WorkspaceManager) IsGrabbing() bool {
	return wm.nearbyGutter.composite != nil && wm.nearbyGutter.grabbed
}

func (wm *WorkspaceManager) GetFocused() *Area {
	return wm.focused
}

func (wm *WorkspaceManager) IsFocusedOnOverlay() bool {
	return wm.focused != nil && wm.focused.IsOverlay()
}

func (wm *WorkspaceManager) setGutter(gutter Gutter) {
	if gutter.composite == nil && wm.nearbyGutter.composite != nil {
		wm.nearbyGutter = Gutter{}
		wm.editor.Host().Window.CursorStandard()
		// wm.drawDebugPanel(matrix.Vec4One())
	} else if gutter.composite != wm.nearbyGutter.composite {
		wm.nearbyGutter = gutter
		// wm.drawDebugPanel(matrix.NewVec4(wm.closeGutter.posx, wm.closeGutter.posy, wm.closeGutter.width, wm.closeGutter.height))
		switch gutter.composite.splitDirection {
		case SplitHorizontal:
			wm.editor.Host().Window.CursorSizeWE()
		case SplitVertical:
			wm.editor.Host().Window.CursorSizeNS()
		default:
			wm.editor.Host().Window.CursorStandard()
		}
	}
}

func (wm *WorkspaceManager) Serialize(ed EditorAreaInterface) {

}

func (wm *WorkspaceManager) constructDefaultWorkspace() {
	_, err := wm.Open("com.kaiju.area_stage_view", nil, SplitHorizontal, 0.5)
	if err != nil {
		panic(err.Error())
	}
	p, _ := wm.Open("com.kaiju.area_content_browser", nil, SplitVertical, -0.5)
	p1, _ := wm.Open("com.kaiju.area_fallback", p, SplitHorizontal, 0.75)
	wm.Open("com.kaiju.area_example", p1, SplitHorizontal, 0.75)
}

func (wm *WorkspaceManager) GetActiveStage() *editor_active_stage.ActiveStage {
	return &wm.stage
}

func (wm *WorkspaceManager) GetAreaByType(*AreaType) *Area {
	return nil
}

func (wm *WorkspaceManager) GetAreaByID(id string) *Area {
	return nil
}
