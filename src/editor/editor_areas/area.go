/******************************************************************************/
/* area.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_areas

import (
	"fmt"
	"log/slog"
	"regexp"
	"weak"

	"golang.org/x/net/html"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering/textures"
)

type SplitDirection uint8
type AreaCapabilityFlags = int
type DiscreteFieldRowAlignment uint8

const (
	SplitHorizontal SplitDirection = iota
	SplitVertical
)

const (
	LeftField DiscreteFieldRowAlignment = iota
	CenterField
	RightField
)

const (
	// Dockable AreaTypes can be dropped into the primary Area hierarchy
	// and tiled by the workspcae manager. Most Areas should be dockable, but
	// some may not desire this. (For example, the splash screen isn't dockable)
	AreaCapabilityFlagDockable = (1 << iota)
	// Overlayable AreaTypes are permitted to float above the layout.
	// Various context menus including the color picker are overlayable.
	AreaCapabilityFlagOverlayable
	// Discoverble AreaTypes are surfaced to the game developer in the editor.
	// The AreaType Selector only includes discoverable areas. If an AreaType
	// isn't marked as discoverable, the only way it may be opened is via
	// editor or plugin code.
	AreaCapabilityFlagDiscoverable
	// For standard areas, Draggable indicates that areas using this
	// AreaType can have their bounds adjusted by dragging with the mouse.
	// For Overlays, this flag indicates that the entire overlay Area and its
	// contents can be moved by dragging its header with the mouse.
	AreaCapabilityFlagDraggable
)

type BottomRightStylizer struct{ ui.BasicStylizer }
type BottomLeftStylizer struct{ ui.BasicStylizer }

// A Container is a hehlper struct that makes laying out elements in code
// easier
type Container struct {
	owner *Area
	Panel *ui.Panel
	doc   *document.Document
}

type ContextBarElement struct {
	owner  *Area
	fields [3]*Container
	Panel  *ui.Panel
}

// The AreaType is a registry construct that exists to convey implementation
// specifics to the Area itself. Each discrete Area type
// (such as the Stage viewport, shader editor, render graph, etc.)
// will exist as an AreaType.
type AreaType struct {
	ID      string
	Name    string
	Handler AreaHandler
	Flags   AreaCapabilityFlags
}

// The Area is the fundamental core of what Kaiju presents to the end user.
type Area struct {
	Container
	Type           AreaType
	Manager        *ui.Manager
	ChildA         *Area
	ChildB         *Area
	SplitDirection SplitDirection
	Ratio          float32
	OverlayPX      float32
	OverlayPY      float32
}

type AreaHandler interface {
	Open(area *Area, editor EditorAreaInterface)
	Close(area *Area, editor EditorAreaInterface)
	FocusInterface(editor EditorAreaInterface)
	BlurInterface(editor EditorAreaInterface)
	Update(area *Area, editor EditorAreaInterface, deltaTime float64, posx, posy, width, height float32)
	// GetOffsets returns relative pixel measurements [posx, posy, width,
	// height] to augment the recursive Area tiling layout. Use this to define
	// a boundary for an Area (such as its top/bottom bars) that the Area tiler
	// won't draw over.
	GetOffsets(ed EditorAreaInterface) (float32, float32, float32, float32)
	// GetOverlayDimensions returns the size [width, height] that this Area
	// assumes by default when opened as an overlay.
	GetOverlayDimensions() (float32, float32)
	IsFocusedOnInput() bool
}

// AreaTypeComposite is a dedicated, default, registry-specific
// AreaType that exists solely to represent Areas that defer their
// functionality to their children. The AreaTypeComposite's handler
// is, by default, nil. The hierarchy of Areas is walked recursively
// The handler here could *potentially* be assigned by a plugin to
// essentially override the recursive Area walking operations performed
// by the functions at the bottom of this file.
// Should such a thing be desirable for whatever reason, this would
// theoretically allow plugins to fundamentally alter the layout behaviour
// of Areas.
var AreaTypeComposite = AreaType{
	ID:      "com.kaiju.area_composite",
	Name:    "Composite",
	Handler: nil,
	Flags:   0,
}

var deferredAreaTypeRegistry = []func() AreaType{}
var regex = regexp.MustCompile(`^[a-z_.]+$`)

// Register allows packages to append their own AreaType factories during
// static initialization.
func Register(areaType func() AreaType) {
	if deferredAreaTypeRegistry == nil {
		slog.Error("Skipped attempt to register an AreaType factory after the registry has closed.")
		return
	}
	deferredAreaTypeRegistry = append(deferredAreaTypeRegistry, areaType)
}

func newBlankArea(areaType *AreaType) *Area {
	return &Area{
		Type:           *areaType,
		owner:          nil,
		ChildA:         nil,
		ChildB:         nil,
		SplitDirection: SplitHorizontal,
		Ratio:          -1,
		OverlayPX:      -1,
		OverlayPY:      -1,
	}
}

// Called by the workspace manager and serializer to set up this Area's
// manager and panel
func (a *Area) open(wm *WorkspaceManager) {
	a.openWithSize(wm, 10, 10, 0)
}

func (a *Area) openWithSize(wm *WorkspaceManager, width, height, radius float32) {
	a.Manager = &ui.Manager{}
	a.Manager.Init(wm.editor.Host())
	a.Panel = a.Manager.Add().ToPanel()
	a.Panel.Init(nil, ui.ElementTypePanel)
	a.SetBackdropColor(wm.editor.Theme().ColorPanel)
	a.Panel.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	a.Panel.Base().Layout().SetOffset(0, 0)
	a.Panel.Base().Layout().Scale(float32(width), float32(height))
	a.Panel.DontFitContent()
	a.Panel.SetBorderRadius(radius, radius, radius, radius)
	a.Type.Handler.Open(a, wm.editor)
	slog.Debug(fmt.Sprintf("Opened '%s'", a.Type.ID))
}

func (a *Area) close(wm *WorkspaceManager) {
	a.Type.Handler.Close(a, wm.editor)
	a.Manager.Shutdown()
	a.Type = AreaType{
		ID:      "NIL_CLOSED",
		Name:    "NIL_CLOSED",
		Handler: nil,
		Flags:   0,
	}
	a.Manager = nil
	a.Panel = nil
	a.owner = nil
	a.ChildA = nil
	a.ChildB = nil
	a.SplitDirection = SplitHorizontal
	a.Ratio = -1
	a.OverlayPX = -1
	a.OverlayPY = -1
}

// StretchFlexColumn configures this Area's root panel to be a flex column with
// stretch alignment. This makes menu & status bar alignment an emergent
// property of the panel as a flex container, rather than requiring explicit
// placement.
func (a *Area) StretchFlexColumn() {
	a.assertInitialized()
	a.Panel.SetFlex()
	a.Panel.SetFlexDirection(ui.FlexDirectionColumn)
	a.Panel.SetFlexAlignItems(ui.FlexAlignStretch)
	a.Panel.SetFlexJustify(ui.FlexJustifySpaceBetween)
}

// HTMLContainer creates a document from the HTML file at the provided relative
// path. The resulting document is given the global CSS for the current editor
// theme, configured for this Area, and added to it. This is the only call you
// should need to get UI to appear in an area.
func (a *Area) HTMLContainer(ed EditorAreaInterface, relativePath string, functions ...func(*document.Element)) *Container {
	css := ed.Theme().GetGlobalCSS(ed.Host())
	doc, err := ed.Host().AssetDatabase().ReadText(relativePath)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to read file '%s'", relativePath), "error", err)
		return &Container{owner: a, doc: nil, Panel: nil}
	}
	slog.Debug(fmt.Sprintf("Opening document '%s'", relativePath))
	docRoot := a.MakeDocumentRoot()
	if len(functions) <= 0 {
		return a.applyDocument(docRoot.UIPanel, markup.DocumentFromHTMLString(a.Manager, doc, css, nil, nil, docRoot))
	}
	var funcMap map[string]func(*document.Element) = make(map[string]func(*document.Element))
	for x, member := range functions {
		if member == nil {
			continue
		}
		fnName := klib.NameOfFunction(member)
		if fnName == "" {
			slog.Error(fmt.Sprintf("Failed to encode function at index %v", x))
		}
		slog.Debug(fmt.Sprintf("Mapped HTML function '%s'", fnName))
		funcMap[fnName] = member
	}
	return a.applyDocument(docRoot.UIPanel, markup.DocumentFromHTMLString(a.Manager, doc, css, nil, funcMap, docRoot))
}

func (a *Area) applyDocument(root *ui.Panel, doc *document.Document) *Container {
	if doc == nil || len(doc.Elements) == 0 {
		slog.Error(fmt.Sprintf(
			"Couldn't apply HTML document for '%s': no elements",
			a.Type.String(),
		))
		return &Container{
			owner: a,
			doc:   doc,
			Panel: nil,
		}
	}
	output := &Container{
		owner: a,
		doc:   doc,
		Panel: root,
	}
	return output
}

// MakeDocumentRoot creates a new, blank Element and configures its panel.
// The resulting panel belongs to this Area's manager and parented to its root.
// A call to this method is made automatically by HTMLContainer() when building
// a layout from markup
func (a *Area) MakeDocumentRoot() *document.Element {
	a.assertInitialized()
	docRoot := a.Manager.Add().ToPanel()
	docRoot.Init(nil, ui.ElementTypePanel)
	docRoot.AllowClickThrough()
	docRoot.DontFitContent()
	docRoot.Base().Layout().Stylizer = ui.StretchCenterStylizer{Parent: weak.Make(a.Panel.Base())}
	docRoot.Base().Layout().SetZ(a.Panel.Base().Layout().Z() + 0.1)
	a.Panel.AddChild(docRoot.Base())
	return &document.Element{
		Type:     html.ElementNode,
		Data:     a.Type.ID + ".root",
		UI:       docRoot.Base(),
		UIPanel:  docRoot,
		Children: make([]*document.Element, 0),
	}
}

func imagePanel(owner *ui.Panel, manager *ui.Manager, ed EditorAreaInterface, texture string, ratio float32, blend bool) *ui.Panel {
	panel := manager.Add().ToPanel()
	tex, err := ed.Host().TextureCache().Texture(texture, textures.TextureSamplerModeClip, textures.TextureFilterLinear)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to load texture '%s'", texture), "error", err)
		return panel
	}
	panel.Init(tex, ui.ElementTypePanel)
	if blend {
		panel.SetUseBlending(true)
	}
	panel.SetColor(matrix.ColorWhite())
	panel.Base().Layout().SetPositioning(ui.PositioningRelative)
	panel.Base().Layout().SetOffset(0, 0)
	aspect := panel.Background().Size().Y() / panel.Background().Size().X()
	width := owner.Base().Layout().PixelSize().X() * ratio
	panel.Base().Layout().Scale(width, width*aspect)
	panel.DontFitContent()
	panel.Base().Layout().SetZ(owner.Base().Layout().Z() + float32(owner.Base().Entity().ChildCount()))
	return panel
}

// BackgroundImage creates an image element that inherits the styling of this Area's
// root panel.
func (a *Area) BackgroundImage(editor EditorAreaInterface, texture string, widthRatio float32) *Container {
	a.assertInitialized()
	panel := imagePanel(a.Panel, a.Manager, editor, texture, widthRatio, false)
	a.Panel.AddChild(panel.Base())
	panel.InheritBorderRadiusFrom(a.Panel)
	return &Container{owner: a, Panel: panel}
}

// LogoImage is a helper for drawing a floating, alpha-enabled logo in the top left of this Area.
func (a *Area) LogoImage(ed EditorAreaInterface, texture string, widthRatio float32) *Container {
	a.assertInitialized()
	panel := imagePanel(a.Panel, a.Manager, ed, texture, widthRatio, true)
	panel.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	a.Panel.AddChild(panel.Base())
	return &Container{owner: a, Panel: panel}
}

// ContextBar draws the common row-of-rows style ContextBar used everywhere in the Kaiju UI.
func (a *Area) ContextBar(ed EditorAreaInterface) ContextBarElement {
	a.assertInitialized()
	cbContent := a.Manager.Add().ToPanel()
	cbContent.Init(nil, ui.ElementTypePanel)
	cbContent.Base().Layout().SetPositioning(ui.PositioningRelative)
	cbContent.Base().Layout().SetOffset(0, 0)
	cbContent.Base().Layout().SetPadding(10, 3, 10, 3)
	cbContent.DontFitContent()
	cbContent.SetFlex()
	cbContent.SetFlexAlignContent(ui.FlexAlignStretch)
	cbContent.SetFlexDirection(ui.FlexDirectionRow)
	cbContent.Base().Layout().SetZ(a.Panel.Base().Layout().Z() + 1)
	cbContent.Base().Layout().Scale(a.Panel.Base().Layout().PixelSize().X(), ed.Theme().SizeContextBar)
	cbContent.SetColor(ed.Theme().ColorContextBar)
	a.Panel.AddChild(cbContent.Base())
	output := ContextBarElement{owner: a, Panel: cbContent}
	output.fields[LeftField] = output.newField(ui.FlexJustifyStart, ui.FlexAlignStart, ui.FlexJustifySelfStart)
	output.fields[CenterField] = output.newField(ui.FlexJustifyCenter, ui.FlexAlignCenter, ui.FlexJustifySelfCenter)
	output.fields[RightField] = output.newField(ui.FlexJustifyEnd, ui.FlexAlignEnd, ui.FlexJustifySelfEnd)
	for _, field := range output.fields {
		cbContent.AddChild(field.Panel.Base())
	}
	output.Panel.InheritBorderRadiusFrom(output.owner.Panel)
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
	return &Container{owner: cb.owner, Panel: panel}
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

// Annotions are a quick way to display important information. Their presence
// doesn't contribute to repaints or flex body alignments. They float above the
// UI and are attached absolutely to a corner of the associated Container.
func (c *Container) AnnotateTopLeft(editor EditorAreaInterface, text string) *ui.Panel {
	bg := c.owner.Manager.Add().ToPanel()
	bg.Init(nil, ui.ElementTypePanel)
	fs := editor.Theme().SizeTextGlobal
	ff := fs * 0.7
	bg.Base().Layout().SetPadding(ff, fs*0.3, ff, fs*0.3)
	bg.SetColor(editor.Theme().ColorBackground)
	bg.Base().Layout().SetZ(c.Panel.Base().Layout().Z() + 0.1)
	bg.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	content := c.owner.Manager.Add().ToLabel()
	content.Init(text)
	content.SetFontSize(fs)
	content.SetColor(editor.Theme().ColorTextPassive)
	content.EnforceBGColor(editor.Theme().ColorPanel)
	bg.SetBorderRadius(ff, ff, ff, ff)
	bg.AddChild(content.Base())
	c.Panel.AddChild(bg.Base())
	return bg
}

// Annotions are a quick way to display important information. Their presence
// doesn't contribute to repaints or flex body alignments. They float above the
// UI and are attached absolutely to a corner of the associated Container.
func (c *Container) AnnotateTopRight(ed EditorAreaInterface, text string) *ui.Panel {
	bg := c.owner.Manager.Add().ToPanel()
	bg.Init(nil, ui.ElementTypePanel)
	fs := ed.Theme().SizeTextGlobal
	ff := fs * 0.7
	bg.Base().Layout().SetPadding(ff, fs*0.3, ff, fs*0.3)
	bg.SetColor(ed.Theme().ColorBackground)
	bg.Base().Layout().SetZ(c.Panel.Base().Layout().Z() + 0.1)
	bg.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	content := c.owner.Manager.Add().ToLabel()
	content.Init(text)
	content.SetFontSize(ed.Theme().SizeTextGlobal)
	content.SetColor(ed.Theme().ColorTextPassive)
	content.EnforceBGColor(ed.Theme().ColorBackground)
	bg.SetBorderRadius(ff, ff, ff, ff)
	bg.Base().Layout().Stylizer = ui.RightStylizer{Parent: weak.Make(c.Panel.Base())}
	size := bg.Base().Layout().PixelSize()
	bg.Base().Layout().SetOffset(-size.X(), size.Y())
	bg.AddChild(content.Base())
	c.Panel.AddChild(bg.Base())
	return bg
}

// Annotions are a quick way to display important information. Their presence
// doesn't contribute to repaints or flex body alignments. They float above the
// UI and are attached absolutely to a corner of the associated Container.
func (c *Container) AnnotateBottomLeft(ed EditorAreaInterface, text string) *ui.Panel {
	bg := c.owner.Manager.Add().ToPanel()
	bg.Init(nil, ui.ElementTypePanel)
	fs := ed.Theme().SizeTextGlobal
	ff := fs * 0.7
	bg.Base().Layout().SetPadding(ff, fs*0.3, ff, fs*0.3)
	bg.SetColor(ed.Theme().ColorBackground)
	bg.Base().Layout().SetZ(c.Panel.Base().Layout().Z() + 0.1)
	bg.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	content := c.owner.Manager.Add().ToLabel()
	content.Init(text)
	content.SetFontSize(ed.Theme().SizeTextGlobal)
	content.SetColor(ed.Theme().ColorTextPassive)
	content.EnforceBGColor(ed.Theme().ColorBackground)
	bg.SetBorderRadius(ff, ff, ff, ff)
	bg.Base().Layout().Stylizer = BottomLeftStylizer{Parent: weak.Make(c.Panel.Base())}
	size := bg.Base().Layout().PixelSize()
	bg.Base().Layout().SetOffset(size.X(), -size.Y())
	bg.AddChild(content.Base())
	c.Panel.AddChild(bg.Base())
	return bg
}

// Annotions are a quick way to display important information. Their presence
// doesn't contribute to repaints or flex body alignments. They float above the
// UI and are attached absolutely to a corner of the associated Container.
func (c *Container) AnnotateBottomRight(ed EditorAreaInterface, text string) *ui.Panel {
	bg := c.owner.Manager.Add().ToPanel()
	bg.Init(nil, ui.ElementTypePanel)
	fs := ed.Theme().SizeTextGlobal
	ff := fs * 0.7
	bg.Base().Layout().SetPadding(ff, fs*0.3, ff, fs*0.3)
	bg.SetColor(ed.Theme().ColorBackground)
	bg.Base().Layout().SetZ(c.Panel.Base().Layout().Z() + 0.1)
	bg.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	content := c.owner.Manager.Add().ToLabel()
	content.Init(text)
	content.SetFontSize(ed.Theme().SizeTextGlobal)
	content.SetColor(ed.Theme().ColorTextPassive)
	content.EnforceBGColor(ed.Theme().ColorBackground)
	bg.SetBorderRadius(ff, ff, ff, ff)
	bg.Base().Layout().Stylizer = BottomRightStylizer{Parent: weak.Make(c.Panel.Base())}
	size := bg.Base().Layout().PixelSize()
	bg.Base().Layout().SetOffset(-size.X(), -size.Y())
	bg.AddChild(content.Base())
	c.Panel.AddChild(bg.Base())
	return bg
}

// Creates a transparent-to-black overlay gradient whose darkest band is aligned
// with the top of this container. This overlay is configured to respect the
// styles of the underlying panel and conform to it without affecting the layout
// of its children.
func (c *Container) TopShadow(editor EditorAreaInterface, darkness, ratio float32) *ui.Panel {
	ps := c.Panel.Base().Layout().PixelSize()
	overlay := imagePanel(c.Panel, c.owner.Manager, editor, "gradient.png", 1, true)
	overlay.SetColor(matrix.ColorBlack().WithAlpha(darkness))
	c.Panel.AddChild(overlay.Base())
	height := ps.Y() * ratio
	overlay.Base().Layout().Scale(ps.X(), height)
	// we can effectively flip the shadow upside down by sampling the UVs backwards
	overlay.Base().ShaderData().Size2D = matrix.Vec4{0, 0, ps.X(), ps.Y()}
	overlay.Base().ShaderData().UVs = matrix.Vec4{0, 1, 1, -1}
	overlay.InheritBorderRadiusFrom(c.Panel)
	return overlay
}

// Creates a transparent-to-black overlay gradient whose darkest band is aligned
// with the bottom of this container. This overlay is configured to respect the
// styles of the underlying panel and conform to it without affecting the layout
// of its children.
func (c *Container) BottomShadow(editor EditorAreaInterface, darkness, ratio float32) *ui.Panel {
	ps := c.Panel.Base().Layout().PixelSize()
	overlay := imagePanel(c.Panel, c.owner.Manager, editor, "gradient.png", 1, true)
	overlay.SetColor(matrix.ColorBlack().WithAlpha(darkness))
	c.Panel.AddChild(overlay.Base())
	height := ps.Y() * ratio
	overlay.Base().Layout().Scale(ps.X(), height)
	overlay.Base().Layout().SetOffset(0, ps.Y()-height)
	overlay.Base().ShaderData().Size2D = matrix.Vec4{0, 0, ps.X(), ps.Y()}
	overlay.InheritBorderRadiusFrom(c.Panel)
	return overlay
}

// SetMinimumWindowSize ensures that the editor window cannot be resized to
// be smaller than this Area's dimensions. This only really works if the Area
// is constrained to the center of the Editor in some way.
func (a *Area) SetMinimunWindowSize(ed EditorAreaInterface) {
	ed.SetMinimumWindowSize(int(a.Width()*1.1), int(a.Height()*1.1))
}

func (a *Area) Width() float32 {
	if a.Panel == nil {
		return 0
	}
	return a.Panel.Base().Entity().Transform.Scale().X()
}

func (a *Area) Height() float32 {
	if a.Panel == nil {
		return 0
	}
	return a.Panel.Base().Entity().Transform.Scale().Y()
}

func (a *Area) TakedownLayout() {
	a.Manager.Shutdown()
	a.Panel = nil
}

func (a *Area) SetBackdropColor(color matrix.Color) {
	a.assertInitialized()
	a.Panel.SetColor(color)
}

func (a *Area) BackdropColor() matrix.Color {
	a.assertInitialized()
	return a.Panel.Base().ShaderData().BgColor
}

func (a *Area) SetOutlineColor(color matrix.Color) {
	a.assertInitialized()
	a.Panel.SetOutline(2, 0, color)
}

func (a *Area) TopLeftCorner() matrix.Vec2 {
	return a.Panel.Base().Layout().PixelPosition()
}

func (a *Area) TopRightCorner() matrix.Vec2 {
	return a.TopLeftCorner().Add(matrix.Vec2{a.Panel.Base().Layout().PixelSize().X(), 0})
}

func (a *Area) BottomLeftCorner() matrix.Vec2 {
	return a.TopLeftCorner().Add(matrix.Vec2{0, a.Panel.Base().Layout().PixelSize().Y()})
}

func (a *Area) BottomRightCorner() matrix.Vec2 {
	return a.TopLeftCorner().Add(matrix.Vec2{a.Panel.Base().Layout().PixelSize().X(), a.Panel.Base().Layout().PixelSize().Y()})
}

// ConstrainAndScale sets the offset and dimensions of this Area's
// panel while ensuring that the panel cannot leave the window.
// The dimensions provided here are normalized against the DPMM of the window.
func (a *Area) ConstrainAndScale(editor EditorAreaInterface, posx, posy, width, height float32) {
	if a.Type.Handler == nil {
		return
	}
	a.assertInitialized()
	maxW := float32(editor.Host().Window.Width())
	maxH := float32(editor.Host().Window.Height())
	wx, wy := a.Type.Handler.GetOverlayDimensions()
	wx = min(max(wx, 256), maxW-24)
	wy = min(max(wy, 256), maxH-24)
	// TODO actually run impl on this
}

func (a *Area) assertInitialized() {
	if a.Panel == nil || a.Manager == nil {
		panic("UI audit on uninitialized area")
	}
}

func (a *Area) IsOverlay() bool {
	return a.OverlayPX > -1 && a.OverlayPY > -1
}

func (a *Area) Update(editor EditorAreaInterface, deltaTime float64, posx, posy, width, height float32) {
	if a.Panel != nil {
		a.Panel.Base().Layout().SetOffset(posx, posy)
		a.Panel.Base().Layout().Scale(width, height)
	}
	if a.Type.Handler != nil {
		a.Type.Handler.Update(a, editor, deltaTime, posx, posy, width, height)
		ox, oy, ds, dy := a.Type.Handler.GetOffsets(editor)
		posx += ox
		posy += oy
		width += ds
		height += dy
	}
	if a.ChildA != nil && a.ChildB != nil {
		if a.ChildA == a || a.ChildB == a {
			panic("Cycle in Area hierarchy")
		}
		switch a.SplitDirection {
		case SplitHorizontal:
			split := width * a.Ratio
			a.ChildA.Update(editor, deltaTime, posx, posy, split, height)
			a.ChildB.Update(editor, deltaTime, split, posy, width-split, height)
		case SplitVertical:
			split := height * a.Ratio
			a.ChildA.Update(editor, deltaTime, posx, posy, width, height-split)
			a.ChildB.Update(editor, deltaTime, posx, split, width, height-split)
		}
	}
}

// PerformAsChildren selectively recurses into child Areas and runs the
// supplied function, provided the Area either has its own children or a
// valid handler.
func (a *Area) PerformAsChildren(operation func(AreaHandler), editor EditorAreaInterface) {
	if a.Type.Handler != nil {
		operation(a.Type.Handler)
	}
	if a.ChildA != nil && a.ChildB != nil {
		if a.ChildA == a || a.ChildB == a {
			panic("Cycle in Area hierarchy")
		}
		a.ChildA.PerformAsChildren(operation, editor)
		a.ChildB.PerformAsChildren(operation, editor)
	}
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
