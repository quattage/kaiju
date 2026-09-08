package editor_areas

import "fmt"

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
