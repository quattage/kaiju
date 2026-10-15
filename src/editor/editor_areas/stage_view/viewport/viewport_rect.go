package viewport

import "kaijuengine.com/matrix"

type ViewportRect struct {
	Left    float32
	Top     float32
	Width   float32
	Height  float32
	Enabled bool
}

func (b ViewportRect) Valid() bool {
	return b.Width > 0 && b.Height > 0
}

func (b ViewportRect) Size() matrix.Vec2 {
	return matrix.NewVec2(b.Width, b.Height)
}

func (b ViewportRect) ContainsScreenPosition(pos matrix.Vec2) bool {
	if !b.Valid() {
		return false
	}
	return pos.X() >= b.Left &&
		pos.X() <= b.Left+b.Width &&
		pos.Y() >= b.Top &&
		pos.Y() <= b.Top+b.Height
}

func (b ViewportRect) LocalTopFromScreen(pos matrix.Vec2) matrix.Vec2 {
	return matrix.NewVec2(pos.X()-b.Left, pos.Y()-b.Top)
}

func (b ViewportRect) LocalBottomFromScreen(pos matrix.Vec2) matrix.Vec2 {
	return matrix.NewVec2(pos.X()-b.Left, b.Height-(pos.Y()-b.Top))
}

func (b ViewportRect) LocalBottomAreaFromScreenArea(area matrix.Vec4) matrix.Vec4 {
	return matrix.Vec4Area(
		area.Left()-matrix.Float(b.Left),
		matrix.Float(b.Height)-(area.Top()-matrix.Float(b.Top)),
		area.Right()-matrix.Float(b.Left),
		matrix.Float(b.Height)-(area.Bottom()-matrix.Float(b.Top)),
	)
}
