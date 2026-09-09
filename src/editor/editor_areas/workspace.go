package editor_areas

import (
	"fmt"
	"log/slog"

	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/engine/ui"
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
	wm.finalizeATRegistry()
	wm.SwitchToWorkspace("Default")
}

// OpenOverlay creates a new Area of the specified type. If no type exists
// at the provided registry ID, an error will be returned.
// This Area is given overlay status and as such will float above other UI elements
// and capture focus. Position is relative to the top left corner of the overlay
// panel.
func (wm *WorkspaceManager) OpenOverlay(areaID string, posx, posy int) (*AreaHandler, error) {
	areaToOpen := wm.GetAreaType(areaID)
	if areaToOpen == nil {
		return nil, fmt.Errorf("No such area: '%s'", areaID)
	}
	return nil, nil
}

// Open creates, configures, and displays a new Area under the provided parent.
// If no AreaType exists at the provided registry ID, an error will be returned.
// The position of this window is calculated based on nesting.
func (wm *WorkspaceManager) Open(areaID string, parentArea *Area, direction SplitDirection, ratio float32) (*AreaHandler, error) {
	areaToOpen := wm.GetAreaType(areaID)
	if areaToOpen == nil {
		return nil, fmt.Errorf("No such area: '%s'", areaID)
	}
	if parentArea == nil {
		parentArea = wm.MainArea
	}
	if areaToOpen.Handler == nil {
		return nil, fmt.Errorf("%s has no handler!", areaToOpen)
	}
	if areaToOpen.Subtype == SubtypeOverlayOnly {
		return nil, fmt.Errorf("%s isn't dockable!", areaToOpen)
	}
	if parentArea.IsComposite() {
		// under normal circumstances, this will never be hit.
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
	// the stacking order of the splits uses the sign of the ratio
	mar := wm.editor.Settings().MinAreaRatio
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
	parentArea.Root = nil
	parentArea.Type = AreaTypeComposite
	parentArea.SplitDirection = direction
	newArea.open(wm)
	return &newArea.Type.Handler, nil
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
	areaToClose.close()
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

// add all deferred registry entries to the persistent registry
// and close the deferred registry.
func (wm *WorkspaceManager) finalizeATRegistry() {
	if len(deferredAreaTypeRegistry) <= 0 {
		panic("No area registries exist!")
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
