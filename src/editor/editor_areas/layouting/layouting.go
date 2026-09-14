package layouting

import (
	"kaijuengine.com/editor/editor_areas"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
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

type Container struct {
	owner    *editor_areas.Area
	panel    *ui.Panel
	isColumn bool
}

func (c *Container) Label(text string, editor editor_areas.EditorAreaInterface) *ui.Label {
	newLabel := c.owner.Manager.Add().ToLabel()
	newLabel.Init(text)
	newLabel.SetFontSize(editor.Theme().FontSize)
	newLabel.SetColor(editor.Theme().ActiveTextColor.AsColor())
	c.panel.AddChild(newLabel.Base())
	return newLabel
}

const widePadding = 10
const thinPadding = 3

type ContextBarElement struct {
	owner  *editor_areas.Area
	panel  *ui.Panel
	fields [3]*Container
}

func ContextBar(a *editor_areas.Area, editor editor_areas.EditorAreaInterface) ContextBarElement {
	cbContent := PanelAsRow(a.Manager, editor)
	cbContent.Base().Layout().Scale(a.Root.Base().Layout().PixelSize().X(), 24*editor.Theme().ContextBarThickness)
	cbContent.SetColor(editor.Theme().ContextBarColor.AsColor())
	a.Root.AddChild(cbContent.Base())
	output := ContextBarElement{owner: a, panel: cbContent}
	output.fields[LeftField] = output.newField(ui.FlexJustifyStart, ui.FlexAlignStart, ui.FlexJustifySelfStart)
	output.fields[CenterField] = output.newField(ui.FlexJustifyCenter, ui.FlexAlignCenter, ui.FlexJustifySelfCenter)
	output.fields[RightField] = output.newField(ui.FlexJustifyEnd, ui.FlexAlignEnd, ui.FlexJustifySelfEnd)
	for _, field := range output.fields {
		cbContent.AddChild(field.panel.Base())
	}
	return output
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

func (cb *ContextBarElement) Width() float32 {
	return cb.panel.Base().Entity().Transform.WorldScale().X()
}
