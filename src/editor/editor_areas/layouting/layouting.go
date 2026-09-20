package layouting

import (
	"fmt"
	"log/slog"
	"weak"

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

type BottomRightStylizer struct{ ui.BasicStylizer }
type BottomLeftStylizer struct{ ui.BasicStylizer }

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
	owner *editor_areas.Area
	panel *ui.Panel
}

func (c *Container) Label(editor editor_areas.EditorAreaInterface, text string) *ui.Label {
	newLabel := c.owner.Manager.Add().ToLabel()
	newLabel.Init(text)
	newLabel.SetFontSize(editor.Theme().SizeTextGlobal)
	newLabel.SetColor(editor.Theme().ColorTextActive)
	newLabel.SetWrap(false)
	c.panel.AddChild(newLabel.Base())
	return newLabel
}

func (c *Container) LabelLeft(editor editor_areas.EditorAreaInterface, text string) *ui.Panel {
	bg := c.owner.Manager.Add().ToPanel()
	bg.Init(nil, ui.ElementTypePanel)
	fs := editor.Theme().SizeTextGlobal
	ff := fs * 0.3
	bg.Base().Layout().SetPadding(fs*0.5, ff, fs*0.5, ff)
	bg.SetColor(editor.Theme().ColorBackground)
	bg.Base().Layout().SetZ(c.panel.Base().Layout().Z() + 0.1)
	bg.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	content := c.owner.Manager.Add().ToLabel()
	content.Init(text)
	content.SetFontSize(fs)
	content.SetColor(editor.Theme().ColorTextPassive)
	content.EnforceBGColor(editor.Theme().ColorPanel)
	bg.SetBorderRadius(ff, ff, ff, ff)
	bg.AddChild(content.Base())
	c.panel.AddChild(bg.Base())
	return bg
}

func (c *Container) LabelRight(editor editor_areas.EditorAreaInterface, text string) *ui.Panel {
	bg := c.owner.Manager.Add().ToPanel()
	bg.Init(nil, ui.ElementTypePanel)
	fs := editor.Theme().SizeTextGlobal
	ff := fs * 0.3
	bg.Base().Layout().SetPadding(fs*0.5, ff, fs*0.5, ff)
	bg.SetColor(editor.Theme().ColorBackground)
	bg.Base().Layout().SetZ(c.panel.Base().Layout().Z() + 0.1)
	bg.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	content := c.owner.Manager.Add().ToLabel()
	content.Init(text)
	content.SetFontSize(editor.Theme().SizeTextGlobal)
	content.SetColor(editor.Theme().ColorTextPassive)
	content.EnforceBGColor(editor.Theme().ColorBackground)
	bg.SetBorderRadius(ff, ff, ff, ff)
	bg.Base().Layout().Stylizer = ui.RightStylizer{BasicStylizer: ui.BasicStylizer{Parent: weak.Make(c.panel.Base())}}
	size := bg.Base().Layout().PixelSize()
	bg.Base().Layout().SetOffset(-size.X(), size.Y())
	bg.AddChild(content.Base())
	c.panel.AddChild(bg.Base())
	return bg
}

func (c *Container) LabelBottomLeft(editor editor_areas.EditorAreaInterface, text string) *ui.Panel {
	bg := c.owner.Manager.Add().ToPanel()
	bg.Init(nil, ui.ElementTypePanel)
	fs := editor.Theme().SizeTextGlobal
	ff := fs * 0.3
	bg.Base().Layout().SetPadding(fs*0.5, ff, fs*0.5, ff)
	bg.SetColor(editor.Theme().ColorBackground)
	bg.Base().Layout().SetZ(c.panel.Base().Layout().Z() + 0.1)
	bg.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	content := c.owner.Manager.Add().ToLabel()
	content.Init(text)
	content.SetFontSize(editor.Theme().SizeTextGlobal)
	content.SetColor(editor.Theme().ColorTextPassive)
	content.EnforceBGColor(editor.Theme().ColorBackground)
	bg.SetBorderRadius(ff, ff, ff, ff)
	bg.Base().Layout().Stylizer = BottomLeftStylizer{BasicStylizer: ui.BasicStylizer{Parent: weak.Make(c.panel.Base())}}
	size := bg.Base().Layout().PixelSize()
	bg.Base().Layout().SetOffset(size.X(), -size.Y())
	bg.AddChild(content.Base())
	c.panel.AddChild(bg.Base())
	return bg
}

func (c *Container) LabelBottomRight(editor editor_areas.EditorAreaInterface, text string) *ui.Panel {
	bg := c.owner.Manager.Add().ToPanel()
	bg.Init(nil, ui.ElementTypePanel)
	fs := editor.Theme().SizeTextGlobal
	ff := fs * 0.3
	bg.Base().Layout().SetPadding(fs*0.5, ff, fs*0.5, ff)
	bg.SetColor(editor.Theme().ColorBackground)
	bg.Base().Layout().SetZ(c.panel.Base().Layout().Z() + 0.1)
	bg.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	content := c.owner.Manager.Add().ToLabel()
	content.Init(text)
	content.SetFontSize(editor.Theme().SizeTextGlobal)
	content.SetColor(editor.Theme().ColorTextPassive)
	content.EnforceBGColor(editor.Theme().ColorBackground)
	bg.SetBorderRadius(ff, ff, ff, ff)
	bg.Base().Layout().Stylizer = BottomRightStylizer{BasicStylizer: ui.BasicStylizer{Parent: weak.Make(c.panel.Base())}}
	size := bg.Base().Layout().PixelSize()
	bg.Base().Layout().SetOffset(-size.X(), -size.Y())
	bg.AddChild(content.Base())
	c.panel.AddChild(bg.Base())
	return bg
}

func (c *Container) TopShadow(editor editor_areas.EditorAreaInterface, darkness, ratio float32) *ui.Panel {
	ps := c.panel.Base().Layout().PixelSize()
	overlay := ImagePanel(c.panel, c.owner.Manager, editor, "gradient.png", 1, true)
	overlay.SetColor(matrix.ColorBlack().WithAlpha(darkness))
	c.panel.AddChild(overlay.Base())
	height := ps.Y() * ratio
	overlay.Base().Layout().Scale(ps.X(), height)
	overlay.Base().ShaderData().Size2D = matrix.Vec4{0, 0, ps.X(), ps.Y()}
	overlay.Base().ShaderData().UVs = matrix.Vec4{0, 1, 1, -1}
	inheritRoundness(overlay, c.panel)
	return overlay
}

func (c *Container) BottomShadow(editor editor_areas.EditorAreaInterface, darkness, ratio float32) *ui.Panel {
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
	cbContent.Base().Layout().Scale(a.Root.Base().Layout().PixelSize().X(), editor.Theme().SizeContextBar)
	cbContent.SetColor(editor.Theme().ColorContextBar)
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
	panel := PanelAsColumn(a.Manager, editor)
	a.Root.AddChild(panel.Base())
	panel.Base().Layout().SetZ(a.Root.Base().Layout().Z() + 0.1)
	return &Container{owner: a, panel: panel}
}

func BackgroundImage(a *editor_areas.Area, editor editor_areas.EditorAreaInterface, texture string, widthRatio float32) *Container {
	panel := ImagePanel(a.Root, a.Manager, editor, texture, widthRatio, false)
	a.Root.AddChild(panel.Base())
	inheritRoundness(panel, a.Root)
	return &Container{owner: a, panel: panel}
}

func LogoImage(a *editor_areas.Area, editor editor_areas.EditorAreaInterface, texture string, widthRatio float32) *Container {
	panel := ImagePanel(a.Root, a.Manager, editor, texture, widthRatio, true)
	a.Root.AddChild(panel.Base())
	return &Container{owner: a, panel: panel}
}

func HTMLContainer(a *editor_areas.Area, editor editor_areas.EditorAreaInterface, location string) *Container {
	css := editor.Theme().GetGlobalCSS()
	fmt.Printf(css)
	return &Container{}
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
	return &Container{owner: cb.owner, panel: panel}
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

func (s BottomRightStylizer) ProcessStyle(layout *ui.Layout) []error {
	width := float32(layout.Ui().Host().Window.Width())
	height := float32(layout.Ui().Host().Window.Height())
	parent := s.Parent.Value()
	if parent != nil {
		width = parent.Layout().PixelSize().X()
		height = parent.Layout().PixelSize().Y()
	}
	selfWidth := layout.PixelSize().X()
	layout.SetInnerOffsetTop(-height + layout.PixelSize().Y())
	layout.SetInnerOffsetLeft(width - selfWidth)
	return nil
}

func (s BottomLeftStylizer) ProcessStyle(layout *ui.Layout) []error {
	height := float32(layout.Ui().Host().Window.Height())
	parent := s.Parent.Value()
	if parent != nil {
		height = parent.Layout().PixelSize().Y()
	}
	selfHeight := layout.PixelSize().Y()
	layout.SetInnerOffsetTop(-height + selfHeight)
	return nil
}
