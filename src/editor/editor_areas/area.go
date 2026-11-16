/******************************************************************************/
/* area.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_areas

import (
	"fmt"
	"log/slog"
	"math"
	"reflect"
	"weak"

	"golang.org/x/net/html"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/textures"
)

type SplitDirection uint8
type AreaCapabilityFlags int
type DiscreteFieldRowAlignment uint8

// The smallest that an Area is allowed to be (unit is effectively millimeters
// in the smallest dimension, so width and height)
const minFunctionalWidth = 16

const (
	// Used to indicate invalid state
	SplitNone SplitDirection = iota
	// Areas will be aligned side-by-side
	SplitHorizontal
	// Areas will be stacked on top of each other
	SplitVertical
)

const (
	LeftField DiscreteFieldRowAlignment = iota
	CenterField
	RightField
)

const (
	// Dockable AreaTypes can be dropped into the primary Area hierarchy
	// and tiled by the workspace manager. Most Areas should be dockable, but
	// some may not desire this. (For example, the splash screen isn't dockable)
	AreaCapabilityFlagDockable AreaCapabilityFlags = 1 << iota
	// Overlayable AreaTypes are permitted to float above the layout.
	// Various context menus including the color picker are overlayable.
	AreaCapabilityFlagOverlayable
	// Discoverble AreaTypes are surfaced to the game developer in the editor.
	// The AreaType Selector only includes discoverable areas. If an AreaType
	// isn't marked as discoverable, the only way it may be opened is via
	// editor or plugin code.
	AreaCapabilityFlagDiscoverable
)

type BottomRightStylizer struct{ ui.BasicStylizer }
type BottomLeftStylizer struct{ ui.BasicStylizer }
type CenterStylizer struct{ ui.BasicStylizer }
type MaxWidthStylizer struct{ ui.BasicStylizer }

type ContextBarElement struct {
	owner  *Area
	fields [3]*Container
	Panel  *ui.Panel
}

// A hitbox used for cursor interactions when resizing Areas
type Gutter struct {
	composite                       *Area
	posx, posy, width, height, dist float32
	grabbed                         bool
}

type AreaHandlerInterface interface {
	Open(area *Area, editor EditorAreaInterface)
	Close(area *Area, editor EditorAreaInterface)
	FocusInterface(editor EditorAreaInterface)
	BlurInterface(editor EditorAreaInterface)
	Update(area *Area, editor EditorAreaInterface, deltaTime float64, posx, posy, width, height float32)
	// GetOverlayDimensions returns the size [width, height] that this Area
	// assumes by default when opened as an overlay.
	GetOverlayDimensions() (float32, float32)
}

// The AreaType is a registry construct that exists to convey implementation
// specifics to the Area itself. Each discrete Area type
// (such as the Stage viewport, shader editoGr, render graph, etc.)
// will exist as an AreaType.
type AreaType struct {
	Key            klib.RegistryKey
	HandlerFactory func() AreaHandlerInterface
	Flags          AreaCapabilityFlags
}

func (at AreaType) GetRegistryKey() klib.RegistryKey {
	return at.Key
}

func (at *AreaType) String() string {
	return at.Key.String()
}

type Container struct {
	owner *Area
	panel *ui.Panel
	doc   *document.Document
}

func (c Container) GetPanel() *ui.Panel {
	return c.panel
}

func (c Container) GetOwner() *Area {
	return c.owner
}

func (c Container) GetDocument() *document.Document {
	return c.doc
}

// The Area is the fundamental core of what Kaiju presents to the end user.
type Area struct {
	Container
	areaType        *AreaType
	manager         *ui.Manager
	childA          *Area
	childB          *Area
	internalHandler AreaHandlerInterface
	splitDirection  SplitDirection
	ratio           float32
	overlayPX       float32
	overlayPY       float32
	focused         bool
	rect            AreaRect
}

func (a *Area) String() string {
	return a.areaType.String()
}

// Returns true if this Area's backing type matches the provided AreaType.
func (a *Area) Is(at *AreaType) bool {
	return a.areaType == at
}

// Returns true if this Area's internal handler is the same instance as the
// provided interface. This is useful
func (a *Area) IsFocused() bool {
	return a.IsFunctional() && a.focused
}

// Returns true if this Area is composite. Composite Areas' sole responsibility
// is to hold exactly two child Areas. Composite areas have no UI of their own
// and propagate state updates/reads to their children.
func (a *Area) IsComposite() bool {
	return a.areaType == nil && a.childA != nil && a.childB != nil
}

// Returns true if this Area is functional. Functional Areas are responsible for
// presenting UI to the user.
func (a *Area) IsFunctional() bool {
	return a.areaType != nil && a.childA == nil && a.childB == nil && a.manager != nil && a.panel != nil
}

// Returns true if this Area is the MainArea, that is, it is the top-level
// composite currently being managed by the workspace
func (a *Area) IsMain(wm *WorkspaceManager) bool {
	return a.owner == nil && a == wm.main
}

// An overlay floats above the contents of the Area system. It has no owner
// or children and describes its location as an absolute position rather than
// existing within the hierarchy.
func (a *Area) IsOverlay() bool {
	return a.owner == nil && a.overlayPX > -1 && a.overlayPY > -1
}

func (a *Area) Host() *engine.Host {
	if a.manager == nil {
		return nil
	}
	return a.manager.Host
}

func (a *Area) Enable() {
	a.SetEnabled(true)
}

func (a *Area) Disable() {
	a.SetEnabled(false)
}

func (a *Area) SetEnabled(enabled bool) {
	if a.panel != nil {
		a.panel.Base().Entity().SetActive(enabled)
	}
}

func (a *Area) GetManager() *ui.Manager {
	return a.manager
}

func (a *Area) GetOwner() *Area {
	return a.owner
}

// Returns the SplitDirection (Horizontal or Vertical) for this Area.
func (a *Area) GetSplitDirection() SplitDirection {
	if a.splitDirection != SplitNone {
		return a.splitDirection
	}
	if a.owner != nil {
		return a.owner.GetSplitDirection()
	}
	return SplitNone
}

// GetHandler is internally (and can be called by editor or plugin code) to
// return the interhal Handler instance owned by this Area. This instance is
// provided initially by the AreaType's factory. The first call to this method
// will create the handler, and subsequent calls will return it. The handler
// should not be invalidated throughout this Area's lifespan - the Area will do
// that automatically when it is closed.
func (a *Area) GetHandler() AreaHandlerInterface {
	if a.internalHandler != nil {
		return a.internalHandler
	}
	if a.areaType.HandlerFactory != nil {
		a.internalHandler = a.areaType.HandlerFactory()
		if a.internalHandler == nil {
			panic(fmt.Sprintf("handler factory for '%s' produced a nil result", a.areaType))
		}
		return a.internalHandler
	}
	panic("handler requested for non-presented area")
}

// Aquire the lazy-loaded AreaHandlerInterface instance from this Area. This
// can be useful if you need the actual typed instance for handler-specific
// data access. If this area's handler can't be acquired or cast, this method
// will panic.
func (a *Area) GetHandlerAs[T AreaHandlerInterface]() T {
	if handler, ok := a.GetHandler().(T); ok {
		return handler
	}
	panic(fmt.Sprintf(
		"failed getting area handler - '%s' cannot conform to type '%s'",
		reflect.TypeOf(a.internalHandler),
		reflect.TypeFor[T](),
	))
}

func newBlankArea(areaType *AreaType) *Area {
	if areaType == nil {
		panic("nil areatype for new blank area")
	}
	output := &Area{}
	output.resetFields()
	output.areaType = areaType
	return output
}

func (a *Area) open(wm *WorkspaceManager, width, height float32) {
	a.manager = &ui.Manager{}
	a.manager.Init(wm.editor.Host())
	a.panel = a.manager.Add().ToPanel()
	a.panel.Init(nil, ui.ElementTypePanel)
	a.SetBackdropColor(wm.editor.Theme().ColorPanel)
	a.panel.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	a.panel.Base().Layout().SetOffset(0, 0)
	a.panel.Base().Layout().Scale(float32(width), float32(height))
	a.panel.DontFitContent()
	rad := wm.Editor().Theme().SizeRadiusGlobal
	a.panel.SetBorderRadius(rad, rad, rad, rad)
	a.Unhighlight(wm.editor)
	a.setupTemplate(wm)
	a.GetHandler().Open(a, wm.editor)
	a.panel.Base().SetDirty(ui.DirtyTypeGenerated)
}

func (a *Area) setupTemplate(wm *WorkspaceManager) {
	if a.IsOverlay() {
		return
	}
	css := wm.editor.Theme().GetGlobalCSS(wm.editor.Host())
	file, err := wm.editor.Host().AssetDatabase().ReadText("areaTemplate.html")
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open area '%s' - areaTemplate.html! couldn't be opened! (%s)", a.areaType.Key.ID(), err))
	}
	docRoot := a.MakeDocumentRoot()
	docRoot.UI.Layout().SetPositioning(ui.PositioningAbsolute)
	docRoot.Data = a.areaType.Key.ID() + ".template"
	docRoot.UI.Layout().SetZ(a.panel.Base().Layout().Z() + 0.1)
	markup.DocumentFromHTMLString(a.manager, file, css, nil, nil, docRoot)
}

func (a *Area) closeAndInvalidate(ed EditorAreaInterface) {
	a.GetHandler().Close(a, ed)
	if a.manager != nil {
		a.manager.Shutdown()
	}
	a.resetFields()
}

func (a *Area) resetFields() {
	a.areaType = nil
	a.manager = nil
	a.panel = nil
	a.owner = nil
	a.childA = nil
	a.childB = nil
	a.splitDirection = SplitHorizontal
	a.ratio = -1
	a.overlayPX = -1
	a.overlayPY = -1
	a.focused = false
	a.rect = AreaRect{-1, -1, -1, -1}
}

// The rest of the file contains layout helpers for building basic UI elements
// by describing them in Go code rather than with HTML/CSS. All the stuff here
// is styled automatically using the editor theme.

// StretchFlexColumn configures this Area's root panel to be a flex column with
// stretch alignment. This makes menu & status bar alignment an emergent
// property of the panel as a flex container, rather than requiring explicit
// placement.
func (a *Area) StretchFlexColumn() {
	a.assertInitialized()
	a.panel.SetFlex()
	a.panel.SetFlexDirection(ui.FlexDirectionColumn)
	a.panel.SetFlexAlignItems(ui.FlexAlignStretch)
	a.panel.SetFlexJustify(ui.FlexJustifySpaceBetween)
}

// MakeDocumentRoot creates a new, blank Element and configures its panel.
// The resulting panel belongs to this Area's manager and parented to its root.
// A call to this method is made automatically by HTMLContainer() when building
// a layout from markup
func (a *Area) MakeDocumentRoot() *document.Element {
	a.assertInitialized()
	docRoot := a.manager.Add().ToPanel()
	docRoot.Init(nil, ui.ElementTypePanel)
	docRoot.AllowClickThrough()
	docRoot.DontFitContent()
	docRoot.Base().Layout().Stylizer = ui.StretchCenterStylizer{Parent: weak.Make(a.panel.Base())}
	docRoot.Base().Layout().SetZ(a.panel.Base().Layout().Z() + 0.001)
	a.panel.AddChild(docRoot.Base())
	return &document.Element{
		Type:     html.ElementNode,
		Data:     a.areaType.Key.ID() + ".root",
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
	panel := imagePanel(a.panel, a.manager, editor, texture, widthRatio, false)
	a.panel.AddChild(panel.Base())
	panel.InheritBorderRadiusFrom(a.panel)
	return &Container{owner: a, panel: panel}
}

// LogoImage is a helper for drawing a floating, alpha-enabled logo in the top left of this Area.
func (a *Area) LogoImage(ed EditorAreaInterface, texture string, widthRatio float32) *Container {
	a.assertInitialized()
	panel := imagePanel(a.panel, a.manager, ed, texture, widthRatio, true)
	panel.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	a.panel.AddChild(panel.Base())
	return &Container{owner: a, panel: panel}
}

func (a *Area) EmptyContainer(editor EditorAreaInterface) *Container {
	a.assertInitialized()
	panel := a.manager.Add().ToPanel()
	panel.Init(nil, ui.ElementTypePanel)
	a.panel.AddChild(panel.Base())
	size := a.panel.Base().Layout().PixelSize()
	panel.Base().Layout().Scale(size.X(), size.Y())
	panel.Base().Layout().SetPositioning(ui.PositioningRelative)
	panel.Base().Layout().SetOffset(0, 0)
	panel.SetColor(editor.Theme().ColorPanel)
	panel.DontFitContent()
	panel.Base().Layout().SetZ(a.panel.Base().Layout().Z() + 0.001)
	return &Container{owner: a, panel: panel}
}

func (a *Area) VirtualList(editor EditorAreaInterface) *ui.VirtualList {
	a.assertInitialized()
	vl := a.manager.Add().ToVirtualList()
	vl.Init()
	size := a.panel.Base().Layout().PixelSize()
	vl.Base().Layout().Scale(size.X(), size.Y())
	vl.Base().Layout().SetPositioning(ui.PositioningRelative)
	vl.Base().Layout().SetOffset(0, 0)
	vl.Base().Layout().SetZ(a.panel.Base().Layout().Z() + 0.001)
	return vl
}

// ContextBar draws the common row-of-rows style ContextBar used everywhere in the Kaiju UI.
func (a *Area) ContextBar(ed EditorAreaInterface) ContextBarElement {
	a.assertInitialized()
	cbContent := a.manager.Add().ToPanel()
	cbContent.Init(nil, ui.ElementTypePanel)
	cbContent.Base().Layout().SetPositioning(ui.PositioningRelative)
	cbContent.Base().Layout().SetOffset(0, 0)
	cbContent.Base().Layout().SetPadding(10, 3, 10, 3)
	cbContent.DontFitContent()
	cbContent.SetFlex()
	cbContent.SetFlexAlignContent(ui.FlexAlignStretch)
	cbContent.SetFlexDirection(ui.FlexDirectionRow)
	cbContent.Base().Layout().SetZ(a.panel.Base().Layout().Z() + 1)
	cbContent.Base().Layout().Scale(a.panel.Base().Layout().PixelSize().X(), ed.Theme().SizeContextBar)
	cbContent.SetColor(ed.Theme().ColorContextBar)
	a.panel.AddChild(cbContent.Base())
	output := ContextBarElement{owner: a, Panel: cbContent}
	output.fields[LeftField] = output.newField(ui.FlexJustifyStart, ui.FlexAlignStart, ui.FlexJustifySelfStart)
	output.fields[CenterField] = output.newField(ui.FlexJustifyCenter, ui.FlexAlignCenter, ui.FlexJustifySelfCenter)
	output.fields[RightField] = output.newField(ui.FlexJustifyEnd, ui.FlexAlignEnd, ui.FlexJustifySelfEnd)
	for _, field := range output.fields {
		cbContent.AddChild(field.panel.Base())
	}
	output.Panel.InheritBorderRadiusFrom(output.owner.panel)
	return output
}

// HTMLContainer creates a document from the HTML file at the provided relative
// path. The resulting document is given the global CSS for the current editor
// theme, configured for this Area, and added to it. This is the only call you
// should need to get UI to appear in an area.
func (a *Area) HTMLContainer(ed EditorAreaInterface, relativePath string, withData any, functions ...func(*document.Element)) (*Container, error) {
	css := ed.Theme().GetGlobalCSS(ed.Host())
	doc, err := ed.Host().AssetDatabase().ReadText(relativePath)
	if err != nil {
		return &Container{owner: a, doc: nil, panel: nil}, err
	}
	slog.Debug(fmt.Sprintf("Opening document '%s'", relativePath))
	docRoot := a.MakeDocumentRoot()
	if len(functions) <= 0 {
		return a.applyDocument(docRoot.UIPanel, markup.DocumentFromHTMLString(a.manager, doc, css, withData, nil, docRoot))
	}
	var funcMap map[string]func(*document.Element) = mapFunctions(functions)
	return a.applyDocument(docRoot.UIPanel, markup.DocumentFromHTMLString(a.manager, doc, css, withData, funcMap, docRoot))
}

// LoadHTML loads and returns the HTML at the provided filepath without
// modyfing this Area's layout. This can be more flexible than [HTMLContainer()]
// if you need some procedural functionality. For example, the terminal uses
// this method to load its row template, which is duplicated by the terminal's
// handler for each row that's displayed.
func (a *Area) LoadHTML(ed EditorAreaInterface, relativePath string, withData any, functions ...func(*document.Element)) (*document.Document, error) {
	css := ed.Theme().GetGlobalCSS(ed.Host())
	doc, err := ed.Host().AssetDatabase().ReadText(relativePath)
	if err != nil {
		return nil, err
	}
	slog.Debug(fmt.Sprintf("Opening document '%s'", relativePath))
	if len(functions) <= 0 {
		return markup.DocumentFromHTMLString(a.manager, doc, css, withData, nil, nil), nil
	}
	var funcMap map[string]func(*document.Element) = mapFunctions(functions)
	return markup.DocumentFromHTMLString(a.manager, doc, css, withData, funcMap, nil), nil
}

func mapFunctions(functions []func(*document.Element)) map[string]func(*document.Element) {
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
	return funcMap
}

func (a *Area) applyDocument(root *ui.Panel, doc *document.Document) (*Container, error) {
	if doc == nil || len(doc.Elements) == 0 {
		return nil, fmt.Errorf("Couldn't apply HTML document for '%s': no elements", a.areaType.String())
	}
	return &Container{
		owner: a,
		doc:   doc,
		panel: root,
	}, nil
}

func (cb *ContextBarElement) newField(justify ui.FlexJustify, align ui.FlexAlign, justifySelf ui.FlexJustifySelf) *Container {
	panel := cb.owner.manager.Add().ToPanel()
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

// Annotions are a quick way to display important information. Their presence
// doesn't contribute to repaints or flex body alignments. They float above the
// UI and are attached absolutely to a corner of the associated Container.
func (c *Container) AnnotateTopLeft(editor EditorAreaInterface, text string) *ui.Panel {
	bg := c.owner.manager.Add().ToPanel()
	bg.Init(nil, ui.ElementTypePanel)
	fs := editor.Theme().SizeTextGlobal
	ff := fs * 0.7
	bg.Base().Layout().SetPadding(ff, fs*0.3, ff, fs*0.3)
	bg.SetColor(editor.Theme().ColorBackground)
	bg.Base().Layout().SetZ(c.panel.Base().Layout().Z() + 0.001)
	bg.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	content := c.owner.manager.Add().ToLabel()
	content.Init(text)
	content.SetFontSize(fs)
	content.SetColor(editor.Theme().ColorTextPassive)
	content.EnforceBGColor(editor.Theme().ColorPanel)
	bg.SetBorderRadius(ff, ff, ff, ff)
	bg.AddChild(content.Base())
	c.panel.AddChild(bg.Base())
	return bg
}

// Annotions are a quick way to display important information. Their presence
// doesn't contribute to repaints or flex body alignments. They float above the
// UI and are attached absolutely to a corner of the associated Container.
func (c *Container) AnnotateTopRight(ed EditorAreaInterface, text string) *ui.Panel {
	bg := c.owner.manager.Add().ToPanel()
	bg.Init(nil, ui.ElementTypePanel)
	fs := ed.Theme().SizeTextGlobal
	ff := fs * 0.7
	bg.Base().Layout().SetPadding(ff, fs*0.3, ff, fs*0.3)
	bg.SetColor(ed.Theme().ColorBackground)
	bg.Base().Layout().SetZ(c.panel.Base().Layout().Z() + 0.001)
	bg.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	content := c.owner.manager.Add().ToLabel()
	content.Init(text)
	content.SetFontSize(ed.Theme().SizeTextGlobal)
	content.SetColor(ed.Theme().ColorTextPassive)
	content.EnforceBGColor(ed.Theme().ColorBackground)
	bg.SetBorderRadius(ff, ff, ff, ff)
	bg.Base().Layout().Stylizer = ui.RightStylizer{Parent: weak.Make(c.panel.Base())}
	size := bg.Base().Layout().PixelSize()
	bg.Base().Layout().SetOffset(-size.X(), size.Y())
	bg.AddChild(content.Base())
	c.panel.AddChild(bg.Base())
	return bg
}

// Annotions are a quick way to display important information. Their presence
// doesn't contribute to repaints or flex body alignments. They float above the
// UI and are attached absolutely to a corner of the associated Container.
func (c *Container) AnnotateBottomLeft(ed EditorAreaInterface, text string) *ui.Panel {
	bg := c.owner.manager.Add().ToPanel()
	bg.Init(nil, ui.ElementTypePanel)
	fs := ed.Theme().SizeTextGlobal
	ff := fs * 0.7
	bg.Base().Layout().SetPadding(ff, fs*0.3, ff, fs*0.3)
	bg.SetColor(ed.Theme().ColorBackground)
	bg.Base().Layout().SetZ(c.panel.Base().Layout().Z() + 0.001)
	bg.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	content := c.owner.manager.Add().ToLabel()
	content.Init(text)
	content.SetFontSize(ed.Theme().SizeTextGlobal)
	content.SetColor(ed.Theme().ColorTextPassive)
	content.EnforceBGColor(ed.Theme().ColorBackground)
	bg.SetBorderRadius(ff, ff, ff, ff)
	bg.Base().Layout().Stylizer = BottomLeftStylizer{Parent: weak.Make(c.panel.Base())}
	size := bg.Base().Layout().PixelSize()
	bg.Base().Layout().SetOffset(size.X(), -size.Y())
	bg.AddChild(content.Base())
	c.panel.AddChild(bg.Base())
	return bg
}

// Annotions are a quick way to display important information. Their presence
// doesn't contribute to repaints or flex body alignments. They float above the
// UI and are attached absolutely to a corner of the associated Container.
func (c *Container) AnnotateBottomRight(ed EditorAreaInterface, text string) *ui.Panel {
	bg := c.owner.manager.Add().ToPanel()
	bg.Init(nil, ui.ElementTypePanel)
	fs := ed.Theme().SizeTextGlobal
	ff := fs * 0.7
	bg.Base().Layout().SetPadding(ff, fs*0.3, ff, fs*0.3)
	bg.SetColor(ed.Theme().ColorBackground)
	bg.Base().Layout().SetZ(c.panel.Base().Layout().Z() + 0.001)
	bg.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	content := c.owner.manager.Add().ToLabel()
	content.Init(text)
	content.SetFontSize(ed.Theme().SizeTextGlobal)
	content.SetColor(ed.Theme().ColorTextPassive)
	content.EnforceBGColor(ed.Theme().ColorBackground)
	bg.SetBorderRadius(ff, ff, ff, ff)
	bg.Base().Layout().Stylizer = BottomRightStylizer{Parent: weak.Make(c.panel.Base())}
	size := bg.Base().Layout().PixelSize()
	bg.Base().Layout().SetOffset(-size.X(), -size.Y())
	bg.AddChild(content.Base())
	c.panel.AddChild(bg.Base())
	return bg
}

func (a *Area) AnnotateCenter(ed EditorAreaInterface, text string) {
	content := a.manager.Add().ToLabel()
	content.Init(text)
	content.SetFontSize(ed.Theme().SizeTextGlobal * 2)
	content.SetFontFace(rendering.FontBold)
	content.SetColor(ed.Theme().ColorTextPassive)
	a.panel.AddChild(content.Base())
	content.Base().Layout().SetZ(a.panel.Base().Layout().Z() + 0.001)
	content.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	content.Base().Layout().Stylizer = CenterStylizer{Parent: weak.Make(a.panel.Base())}
}

// Creates a transparent-to-black overlay gradient whose darkest band is aligned
// with the top of this container. This overlay is configured to respect the
// styles of the underlying panel and conform to it without affecting the layout
// of its children.
func (c *Container) TopFade(editor EditorAreaInterface, darkness, ratio float32) *ui.Panel {
	ps := c.panel.Base().Layout().PixelSize()
	overlay := imagePanel(c.panel, c.owner.manager, editor, "gradient.png", 1, true)
	overlay.SetColor(matrix.ColorBlack().WithAlpha(darkness))
	c.panel.AddChild(overlay.Base())
	height := ps.Y() * ratio
	overlay.Base().Layout().Scale(ps.X(), height)
	// we can effectively flip the shadow upside down by sampling the UVs backwards
	overlay.Base().ShaderData().Size2D = matrix.Vec4{0, 0, ps.X(), ps.Y()}
	overlay.Base().ShaderData().UVs = matrix.Vec4{0, 1, 1, -1}
	overlay.InheritBorderRadiusFrom(c.panel)
	return overlay
}

// Creates a transparent-to-black overlay gradient whose darkest band is aligned
// with the bottom of this container. This overlay is configured to respect the
// styles of the underlying panel and conform to it without affecting the layout
// of its children.
func (c *Container) BottomFade(editor EditorAreaInterface, darkness, ratio float32) *ui.Panel {
	ps := c.panel.Base().Layout().PixelSize()
	overlay := imagePanel(c.panel, c.owner.manager, editor, "gradient.png", 1, true)
	overlay.SetColor(matrix.ColorBlack().WithAlpha(darkness))
	c.panel.AddChild(overlay.Base())
	height := ps.Y() * ratio
	overlay.Base().Layout().Scale(ps.X(), height)
	overlay.Base().Layout().SetOffset(0, ps.Y()-height)
	overlay.Base().ShaderData().Size2D = matrix.Vec4{0, 0, ps.X(), ps.Y()}
	overlay.InheritBorderRadiusFrom(c.panel)
	return overlay
}

func (a *Area) SetBackdropColor(color matrix.Color) {
	a.assertInitialized()
	a.panel.SetColor(color)
}

func (a *Area) Highlight(ed EditorAreaInterface) {
	a.assertInitialized()
	col := matrix.ColorMix(ed.Theme().ColorPanel, matrix.ColorWhite(), 0.1)
	a.panel.SetBorderColor(col, col, col, col)
	size := float32(math.Max(3, math.Ceil(float64(ed.Theme().SizeTrenchGlobal*0.25))))
	a.panel.SetBorderSize(size, size, size, size)
}

func (a *Area) Unhighlight(ed EditorAreaInterface) {
	a.assertInitialized()
	col := matrix.ColorMix(ed.Theme().ColorPanel, matrix.ColorWhite(), 0.04)
	a.panel.SetBorderColor(col, col, col, col)
	size := float32(math.Max(3, math.Ceil(float64(ed.Theme().SizeTrenchGlobal*0.25))))
	a.panel.SetBorderSize(size, size, size, size)
}

func (a *Area) SetBackdropTexture(ed EditorAreaInterface, tex *rendering.Texture) {
	a.assertInitialized()
	if tex == nil {
		a.panel.SetColor(ed.Theme().ColorPanel)
		a.panel.RemoveBackground()
	} else {
		a.panel.SetColor(matrix.ColorWhite())
		a.panel.SetBackground(tex)
	}
}

func (a *Area) HasBackdropTexture() bool {
	return a.panel != nil && a.panel.HasBackground()
}

func (a *Area) BackdropColor() matrix.Color {
	a.assertInitialized()
	return a.panel.Base().ShaderData().BgColor
}

func (a *Area) SetOutlineColor(color matrix.Color) {
	a.assertInitialized()
	a.panel.SetOutline(2, 0, color)
}

func (a *Area) assertInitialized() {
	if a.panel == nil || a.manager == nil {
		panic("UI audit on uninitialized area")
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

func (s CenterStylizer) ProcessStyle(layout *ui.Layout) []error {
	width := float32(layout.Ui().Host().Window.Width())
	height := float32(layout.Ui().Host().Window.Height())
	parent := s.Parent.Value()
	if parent != nil {
		width = parent.Layout().PixelSize().X()
		height = parent.Layout().PixelSize().Y()
	}
	layout.SetInnerOffsetTop(-height/2 + layout.PixelSize().Y())
	layout.SetInnerOffsetLeft(-width/2 + layout.PixelSize().X())
	return nil
}

func (s MaxWidthStylizer) ProcessStyle(layout *ui.Layout) []error {
	layout.ScaleWidth(s.Parent.Value().Layout().PixelSize().X() - 32)
	return nil
}
