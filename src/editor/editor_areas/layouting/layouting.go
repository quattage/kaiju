package layouting

import (
	"fmt"
	"log/slog"

	"kaijuengine.com/editor/editor_areas"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering/textures"
)

type DiscreteFieldRowAlignment uint8

const (
	LeftField DiscreteFieldRowAlignment = iota
	CenterField
	RightField
)

func PanelAsRow(manager *ui.Manager, editor editor_areas.EditorAreaInterface) *ui.Panel {
	newPanel := manager.Add().ToPanel()
	newPanel.Init(nil, ui.ElementTypePanel)
	newPanel.Base().Layout().SetPositioning(ui.PositioningRelative)
	newPanel.Base().Layout().SetOffset(0, 0)
	newPanel.DontFitContent()
	newPanel.SetFlex()
	newPanel.SetFlexAlignContent(ui.FlexAlignStretch)
	newPanel.SetFlexDirection(ui.FlexDirectionRow)
	newPanel.Base().Layout().SetPadding(widePadding, thinPadding, widePadding, thinPadding)
	return newPanel
}

func PanelAsColumn(manager *ui.Manager, editor editor_areas.EditorAreaInterface) *ui.Panel {
	newPanel := manager.Add().ToPanel()
	newPanel.Init(nil, ui.ElementTypePanel)
	newPanel.Base().Layout().SetPositioning(ui.PositioningRelative)
	newPanel.Base().Layout().SetOffset(0, 0)
	newPanel.DontFitContent()
	newPanel.SetFlex()
	newPanel.SetFlexAlignContent(ui.FlexAlignStretch)
	newPanel.SetFlexDirection(ui.FlexDirectionColumn)
	newPanel.SetFlexJustify(ui.FlexJustifySpaceBetween)
	newPanel.Base().Layout().SetPadding(widePadding, thinPadding, widePadding, thinPadding)
	return newPanel
}

func ImagePanel(owner *ui.Panel, manager *ui.Manager, editor editor_areas.EditorAreaInterface, texture string, ratio float32, blend bool) *ui.Panel {
	panel := manager.Add().ToPanel()
	tex, err := editor.Host().TextureCache().Texture(texture, textures.TextureSamplerModeClip, textures.TextureFilterLinear)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to load texture '%s'", texture), "error", err)
		return panel
	}
	panel.Init(tex, ui.ElementTypePanel)
	if blend {
		panel.SetUseBlending(true)
	}
	panel.SetColor(matrix.ColorWhite())
	panel.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	panel.Base().Layout().SetOffset(0, 0)
	aspect := panel.Background().Size().Y() / panel.Background().Size().X()
	width := owner.Base().Layout().PixelSize().X() * ratio
	panel.Base().Layout().Scale(width, width*aspect)
	panel.DontFitContent()
	panel.Base().Layout().SetZ(owner.Base().Layout().Z() + float32(owner.Base().Entity().ChildCount()))
	return panel
}

type Container struct {
	owner    *editor_areas.Area
	panel    *ui.Panel
	isColumn bool
}

func (c *Container) Label(editor editor_areas.EditorAreaInterface, text string) *ui.Label {
	newLabel := c.owner.Manager.Add().ToLabel()
	newLabel.Init(text)
	newLabel.SetFontSize(editor.Theme().FontSize.FPix())
	newLabel.SetColor(editor.Theme().ActiveTextColor.AsColor())
	c.panel.AddChild(newLabel.Base())
	return newLabel
}

func (c *Container) TopFade(editor editor_areas.EditorAreaInterface, darkness, ratio float32) *ui.Panel {
	ps := c.panel.Base().Layout().PixelSize()
	overlay := ImagePanel(c.panel, c.owner.Manager, editor, "gradient.png", 1, true)
	overlay.SetColor(matrix.ColorBlack().WithAlpha(darkness))
	c.panel.AddChild(overlay.Base())
	height := ps.Y() * ratio
	overlay.Base().Layout().Scale(ps.X(), height)
	overlay.Base().Layout().SetOffset(0, ps.Y()-height)
	overlay.Base().ShaderData().Size2D = matrix.Vec4{0, 0, ps.X(), ps.Y()}
	return overlay
}

func (c *Container) BottomFade(editor editor_areas.EditorAreaInterface, darkness, ratio float32) *ui.Panel {
	ps := c.panel.Base().Layout().PixelSize()
	overlay := ImagePanel(c.panel, c.owner.Manager, editor, "gradient.png", 1, true)
	overlay.SetColor(matrix.ColorBlack().WithAlpha(darkness))
	c.panel.AddChild(overlay.Base())
	height := ps.Y() * ratio
	overlay.Base().Layout().Scale(ps.X(), height)
	overlay.Base().Layout().SetOffset(0, ps.Y()-height)
	overlay.Base().ShaderData().Size2D = matrix.Vec4{0, 0, ps.X(), ps.Y()}
	return overlay
}

const widePadding = 10
const thinPadding = 3

type ContextBarElement struct {
	owner  *editor_areas.Area
	fields [3]*Container
	panel  *ui.Panel
}

func ContextBar(a *editor_areas.Area, editor editor_areas.EditorAreaInterface) ContextBarElement {
	cbContent := PanelAsRow(a.Manager, editor)
	cbContent.Base().Layout().SetZ(a.Root.Base().Layout().Z() + 1)
	cbContent.Base().Layout().Scale(a.Root.Base().Layout().PixelSize().X(), editor.Theme().ContextBarThickness.FPix())
	cbContent.SetColor(editor.Theme().ContextBarColor.AsColor())
	a.Root.AddChild(cbContent.Base())
	output := ContextBarElement{owner: a, panel: cbContent}
	output.fields[LeftField] = output.newField(ui.FlexJustifyStart, ui.FlexAlignStart, ui.FlexJustifySelfStart)
	output.fields[CenterField] = output.newField(ui.FlexJustifyCenter, ui.FlexAlignCenter, ui.FlexJustifySelfCenter)
	output.fields[RightField] = output.newField(ui.FlexJustifyEnd, ui.FlexAlignEnd, ui.FlexJustifySelfEnd)
	for _, field := range output.fields {
		cbContent.AddChild(field.panel.Base())
	}
	inheritRoundness(output.panel, output.owner.Root)
	return output
}

func BlankContainer(a *editor_areas.Area, editor editor_areas.EditorAreaInterface) *Container {
	panel := PanelAsRow(a.Manager, editor)
	a.Root.AddChild(panel.Base())
	panel.Base().Layout().SetZ(a.Root.Base().Layout().Z() + 1)
	inheritRoundness(panel, a.Root)
	return &Container{owner: a, panel: panel, isColumn: false}
}

func BackgroundImage(a *editor_areas.Area, editor editor_areas.EditorAreaInterface, texture string, widthRatio float32) *Container {
	panel := ImagePanel(a.Root, a.Manager, editor, texture, widthRatio, false)
	a.Root.AddChild(panel.Base())
	inheritRoundness(panel, a.Root)
	return &Container{owner: a, panel: panel, isColumn: false}
}

func LogoImage(a *editor_areas.Area, editor editor_areas.EditorAreaInterface, texture string, widthRatio float32) *Container {
	panel := ImagePanel(a.Root, a.Manager, editor, texture, widthRatio, true)
	a.Root.AddChild(panel.Base())
	return &Container{owner: a, panel: panel, isColumn: false}
}

func inheritRoundness(panel *ui.Panel, parent *ui.Panel) {
	parent.Base().Clean()
	barY := panel.Base().Layout().PixelPosition().Y()
	parentY := parent.Base().Layout().PixelPosition().Y()
	if klib.FAbs(barY-parentY) < 0.01 {
		radii := parent.Base().ShaderData().BorderRadius
		panel.SetBorderRadius(radii[3], radii[2], 0, 0)
	} else if klib.FAbs(barY-parentY) > parent.Base().Layout().PixelSize().Y()-panel.Base().Layout().PixelSize().Y()-0.01 {
		radii := parent.Base().ShaderData().BorderRadius
		panel.SetBorderRadius(0, 0, radii[1], radii[0])
	}
}

func (cb *ContextBarElement) newField(justify ui.FlexJustify, align ui.FlexAlign, justifySelf ui.FlexJustifySelf) *Container {
	panel := cb.owner.Manager.Add().ToPanel()
	panel.Init(nil, ui.ElementTypePanel)
	panel.SetColor(matrix.ColorZero())
	panel.SetFlex()
	panel.SetFlexDirection(ui.FlexDirectionRow)
	panel.SetFlexJustify(justify)
	panel.Base().Layout().SetJustifySelf(justifySelf)
	panel.SetFlexAlignItems(align)
	panel.Base().Layout().Scale(0, 0)
	return &Container{owner: cb.owner, panel: panel, isColumn: false}
}

func (cb *ContextBarElement) Field(field DiscreteFieldRowAlignment) *Container {
	return cb.fields[field]
}

func (cb *ContextBarElement) LeftField() *Container {
	return cb.Field(LeftField)
}

func (cb *ContextBarElement) CenterField() *Container {
	return cb.Field(CenterField)
}

func (cb *ContextBarElement) RightField() *Container {
	return cb.Field(RightField)
}
