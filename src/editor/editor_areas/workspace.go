package editor_areas

import (
	"fmt"
	"log/slog"

	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
)

type WorkspaceManager struct {
	editor          EditorAreaView
	workspaces      map[string]WorkspaceConfiguration
	activeWorkspace string
	areaTypes       map[string]AreaType
	MainArea        *Area
	Overlays        []*Area
	StageView       editor_stage_view.StageView
}

func (wm *WorkspaceManager) Initialize(editor EditorAreaView) {
	wm.editor = editor
	wm.activeWorkspace = ""
	if len(wm.areaTypes) <= 0 {
		panic("No area registries exist!")
	}
	wm.SwitchToWorkspace("Default")
}

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

func (wm *WorkspaceManager) GetAreaType(id string) *AreaType {
	if len(id) <= 0 {
		return nil
	}
	areaType, ok := wm.areaTypes[id]
	if !ok {
		return nil
	}
	return &areaType
}

func (wm *WorkspaceManager) GetDefaultAreaType() *AreaType {
	if len(wm.areaTypes) <= 0 {
		panic("No area registries exist!")
	}
	for _, at := range wm.areaTypes {
		return &at
	}
	return nil
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

// TODO move all vars marked with TEMP_SET over to some kind of settings struct
const TEMP_SET_ContextBarThick int = 25

// We don't allow Area split ratios to be smaller than this number, since doing
// so could result in an area whose content is so small that it can't be interacted
// with by the end user.
// TODO this ratio number needs to be addressed since it is arbitrary and not based on
// screen resolution or DPMM. For most displays this number as-is will likely be larger
// than necessary. This value should be derived from DPMM and UI scale settings.
const TEMP_SET_MinAreaRatio float32 = 0.1

// OpenOverlay immediately opens a floating window bound to the AreaType at the
// provided ID, if one exists. Position and size are relative to the top left corner of the window.
func (wm *WorkspaceManager) OpenOverlay(areaID string, posx, posy int) {
}

func (wm *WorkspaceManager) Open(areaID string, parentArea *Area, direction SplitDirection, ratio float32) {
	areaToOpen := wm.GetAreaType(areaID)
	wm.openDirectly(areaToOpen, parentArea, direction, ratio)
}

func (wm *WorkspaceManager) openDirectly(areaToOpen *AreaType, parentArea *Area, direction SplitDirection, ratio float32) {
	if parentArea == nil {
		parentArea = wm.MainArea
	}
	if areaToOpen.Handler == nil {
		slog.Error(fmt.Sprintf("Failed to open %s - this AreaType has no handler!", areaToOpen))
		return
	}
	if areaToOpen.Subtype == SubtypeOverlay {
		slog.Error(fmt.Sprintf("Failed to open %s - this area isn't dockable!", areaToOpen))
		return
	}
	if parentArea.IsComposite() {
		// under normal circumstances, neither child should ever be null
		if parentArea.ChildA != nil {
			wm.openDirectly(areaToOpen, parentArea.ChildA, direction, ratio)
		} else if parentArea.ChildB != nil {
			wm.openDirectly(areaToOpen, parentArea.ChildB, direction, ratio)
		} else {
			slog.Error(fmt.Sprintf("Failed to open %s with parent %s - the parent area is a composite, but has no children! This indicates that the workspace area tree has been invalidated!", areaToOpen, parentArea))
		}
		return
	}
	areaCopy := parentArea
	areaCopy.Parent = parentArea
	areaCopy.SplitDirection = SplitHorizontal
	areaCopy.Ratio = -1
	newArea := &Area{
		Type:           *areaToOpen,
		Manager:        &ui.Manager{},
		Parent:         parentArea,
		IsOverlay:      areaCopy.IsOverlay,
		SplitDirection: SplitHorizontal,
		Ratio:          -1,
	}
	newArea.Manager.Init(wm.editor.Host())
	newArea.Root = newArea.Manager.Add().ToPanel()
	newArea.Root.Init(nil, ui.ElementTypePanel)
	newArea.Root.SetColor(matrix.ColorAzure())
	// The sign of the ratio determines the stacking order of the resulting splits
	if ratio <= 0 {
		parentArea.ChildA = areaCopy
		parentArea.ChildB = newArea
		if ratio < -(1 - TEMP_SET_MinAreaRatio) {
			parentArea.Ratio = 1 - TEMP_SET_MinAreaRatio
		} else if ratio > -TEMP_SET_MinAreaRatio {
			parentArea.Ratio = TEMP_SET_MinAreaRatio
		} else {
			parentArea.Ratio = -ratio
		}
	} else {
		parentArea.ChildA = newArea
		parentArea.ChildB = areaCopy
		if ratio > 1-TEMP_SET_MinAreaRatio {
			parentArea.Ratio = 1 - TEMP_SET_MinAreaRatio
		} else if ratio < TEMP_SET_MinAreaRatio {
			parentArea.Ratio = TEMP_SET_MinAreaRatio
		} else {
			parentArea.Ratio = ratio
		}
	}
	parentArea.Manager = nil
	parentArea.Root = nil
	parentArea.Type = AreaTypeComposite
	parentArea.SplitDirection = direction
	areaToOpen.Handler.Open(newArea, wm.editor)
}

func (wm *WorkspaceManager) Close(areaToClose *Area) {
	if areaToClose == nil {
		slog.Error("Skipped attempt to close a nil area.")
	}
	if areaToClose.Parent != nil {
		switch areaToClose {
		case areaToClose.Parent.ChildA:
			areaToClose.Parent.ChildA = nil
			return
		case areaToClose.Parent.ChildB:
			areaToClose.Parent.ChildB = nil
			return
		}
	}
	if areaToClose.IsComposite() {
		if areaToClose.ChildA != nil {
			wm.Close(areaToClose.ChildA)
		} else if areaToClose.ChildB != nil {
			wm.Close(areaToClose.ChildB)
		}
	}
	areaToClose.Manager.Shutdown()
	areaToClose.Type = AreaType{
		ID:      "NIL_CLOSED",
		Name:    "NIL_CLOSED",
		Handler: nil,
		Subtype: SubtypeRestricted,
	}
	areaToClose.Manager = nil
	areaToClose.Root = nil
	areaToClose.Parent = nil
	areaToClose.ChildA = nil
	areaToClose.ChildB = nil
	areaToClose.SplitDirection = SplitHorizontal
	areaToClose.Ratio = -1
	areaToClose.IsOverlay = false
}

func (wm *WorkspaceManager) FocusInterface() {
	wm.MainArea.PerformAsChildren(func(handler AreaHandler) {
		handler.FocusInterface(wm.editor)
	}, wm.editor)
}

func (wm *WorkspaceManager) BlurInterface() {
	wm.MainArea.PerformAsChildren(func(handler AreaHandler) {
		handler.BlurInterface(wm.editor)
	}, wm.editor)
}

// func (wm *WorkspaceManager) UpdateAll(deltaTime float64) {
// 	wm.MainArea.Update(wm.editor, deltaTime, 0, 0, wm.editor->Host().Window.Width(), editor.Host().Window.Height())
// }
