package editor_areas

import (
	"log/slog"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
)

type AreaRect struct {
	Left   float32
	Top    float32
	Width  float32
	Height float32
}

// Returns the absolute pixel width of this Area's panel. If this Area
// doesn't have a panel (i.e. it's a composite) this method will calculate
// the width based on its children.
func (a *Area) GetWidth() float32 {
	if a.IsComposite() {
		if a.splitDirection == SplitHorizontal {
			return a.childA.GetWidth() + a.childB.GetWidth()
		}
		return a.childA.GetWidth()
	}
	if a.IsFunctional() {
		return a.Panel.Base().Entity().Transform.Scale().X()
	}
	return 0
}

// Returns the absolute pixel height of this Area's panel. If this Area
// doesn't have a panel (i.e. it's a composite) this method will calculate
// the height based on its children.
func (a *Area) GetHeight() float32 {
	if a.IsComposite() {
		if a.splitDirection == SplitVertical {
			return a.childA.GetHeight() + a.childB.GetHeight()
		}
		return a.childA.GetHeight()
	}
	if a.IsFunctional() {
		return a.Panel.Base().Entity().Transform.Scale().Y()
	}
	return 0
}

// Returns the absolute pixel position of this Area's panel. If this Area
// doesn't have a panel (i.e. it's a composite) it will infer the position
// from its children.
func (a *Area) GetPosition() matrix.Vec2 {
	if a.IsComposite() {
		return a.childA.GetPosition()
	}
	if a.IsFunctional() {
		// The panel backing the Area should never have any ancestor UI element.
		// Misuse of the API may allow this to occur, so this may need to be
		// addressed at some point in the future.
		return a.Panel.Base().Layout().Offset()
	}
	return matrix.Vec2Zero()
}

func (a *Area) GetRect() AreaRect {
	if a.IsComposite() {
		return AreaRect{-1, -1, -1, -1}
	} else if a.IsFunctional() {
		pos := a.GetPosition()
		a.rect = AreaRect{pos.X(), pos.Y(), a.GetWidth(), a.GetHeight()}
		return a.rect
	} else {
		slog.Warn("Rect update on dangling area")
		return AreaRect{-1, -1, -1, -1}
	}
}

func (a *Area) InvalidateRect() {
	a.rect = AreaRect{-1, -1, -1, -1}
	if a.IsComposite() {
		a.childA.InvalidateRect()
		a.childB.InvalidateRect()
	}
}

func (a *Area) CursorPositionWithin(cursor *hid.Cursor) matrix.Vec2 {
	rect := a.GetRect()
	if !rect.Valid() {
		return cursor.ScreenPosition()
	}
	return rect.LocalTopFromScreen(cursor.ScreenPosition())
}

func (b AreaRect) Valid() bool {
	return b.Width > 0 && b.Height > 0
}

func (b AreaRect) Size() matrix.Vec2 {
	return matrix.NewVec2(b.Width, b.Height)
}

func (b AreaRect) ContainsScreenPosition(pos matrix.Vec2) bool {
	if !b.Valid() {
		return false
	}
	return pos.X() >= b.Left &&
		pos.X() <= b.Left+b.Width &&
		pos.Y() >= b.Top &&
		pos.Y() <= b.Top+b.Height
}

func (b AreaRect) LocalTopFromScreen(pos matrix.Vec2) matrix.Vec2 {
	return matrix.NewVec2(pos.X()-b.Left, pos.Y()-b.Top)
}

func (b AreaRect) LocalBottomFromScreen(pos matrix.Vec2) matrix.Vec2 {
	return matrix.NewVec2(pos.X()-b.Left, b.Height-(pos.Y()-b.Top))
}

func (b AreaRect) LocalBottomRectFromScreenRect(area matrix.Vec4) matrix.Vec4 {
	return matrix.Vec4Area(
		area.Left()-matrix.Float(b.Left),
		matrix.Float(b.Height)-(area.Top()-matrix.Float(b.Top)),
		area.Right()-matrix.Float(b.Left),
		matrix.Float(b.Height)-(area.Bottom()-matrix.Float(b.Top)),
	)
}
