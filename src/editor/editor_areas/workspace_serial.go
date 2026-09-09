/******************************************************************************/
/* workspace_serial.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_areas

import (
	"fmt"
	"log/slog"

	"kaijuengine.com/engine/ui"
)

type WorkspaceConfiguration struct {
	ID   string
	Name string
	tree []branch
}

type branch struct {
	a        int16
	b        int16
	ratio    float32
	dir      uint8
	areaType string
}

func (wm *WorkspaceManager) SerializeCurrentWorkspace(id string, name string) *WorkspaceConfiguration {
	output := &WorkspaceConfiguration{
		ID:   id,
		Name: name,
	}
	if wm.MainArea == nil {
		return output
	}
	serializeArea(wm.MainArea, &output.tree)
	return output
}

func serializeArea(area *Area, tree *[]branch) int16 {
	index := int16(len(*tree))
	*tree = append(*tree, branch{})
	member := branch{
		a:        -1,
		b:        -1,
		areaType: area.Type.String(),
	}
	if area.ChildA != nil || area.ChildB != nil {
		member.ratio = area.Ratio
		member.dir = uint8(area.SplitDirection)
		if area.ChildA != nil {
			member.a = int16(serializeArea(area.ChildA, tree))
		}
		if area.ChildB != nil {
			member.b = int16(serializeArea(area.ChildB, tree))
		}
	}
	(*tree)[index] = member
	return index
}

func (wm *WorkspaceManager) LoadWorkspace(config *WorkspaceConfiguration) {
	newAreas := []*Area{}
	wm.MainArea = wm.deserializeArea(config.tree, 0, newAreas)
	for _, area := range newAreas {
		if !area.IsComposite() && area.Type.Handler != nil {
			area.open(wm)
		}
	}
}

func (wm *WorkspaceManager) deserializeArea(tree []branch, index int16, newAreas []*Area) *Area {
	member := tree[index]
	if member.a >= 0 && member.b >= 0 {
		area := &Area{
			Type:           AreaTypeComposite,
			Manager:        &ui.Manager{},
			ChildA:         wm.deserializeArea(tree, member.a, newAreas),
			ChildB:         wm.deserializeArea(tree, member.b, newAreas),
			SplitDirection: SplitDirection(member.dir),
			Ratio:          member.ratio,
		}
		newAreas = append(newAreas, area)
		return area
	}
	at := wm.GetAreaType(member.areaType)
	if at == nil {
		at = wm.GetAreaType("com.kaiji.area_fallback")
		slog.Warn(fmt.Sprintf("Couldn't load area '%s', loaded fallback instead", member.areaType))
	}
	area := &Area{
		Type:           *at,
		Manager:        &ui.Manager{},
		ChildA:         nil,
		ChildB:         nil,
		SplitDirection: SplitHorizontal,
		Ratio:          -1,
	}
	newAreas = append(newAreas, area)
	return area
}

func (a *Area) String() string {
	ovl := ""
	if a.IsOverlay {
		ovl = "(overlay)"
	}
	if a.IsComposite() {
		return fmt.Sprintf("Area[%s, %s, Children: (%s, %s), Split: %f %s]", a.Type.String(), ovl, a.ChildA.Type.String(), a.ChildB.Type.String(), a.Ratio, a.SplitDirection)
	} else {
		return fmt.Sprintf("Area[%s, %s]", a.Type.String(), ovl)
	}
}

func (at *AreaType) String() string {
	return fmt.Sprintf("AreaType[%s, '%s']", at.ID, at.Name)
}

func (wc *WorkspaceConfiguration) String() string {
	return fmt.Sprintf("WorkspaceConfiguration[%s, '%s']", wc.ID, wc.Name)
}

func (sd SplitDirection) String() string {
	switch sd {
	case SplitHorizontal:
		return "Horizontal"
	case SplitVertical:
		return "Vertical"
	default:
		return fmt.Sprintf("SplitDirection[%d]", sd)
	}
}

func (av AreaSubtype) String() string {
	switch av {
	case SubtypeDockable:
		return "Available"
	case SubtypeRestricted:
		return "OverlayOnly"
	default:
		return fmt.Sprintf("AreaAccessibility[%d]", av)
	}
}
