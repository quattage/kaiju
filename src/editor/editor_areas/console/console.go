/******************************************************************************/
/* terminal.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package console

import (
	"fmt"
	"strings"
	"sync/atomic"
	"weak"

	"kaijuengine.com/editor/editor_areas"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/systems/console"
	"kaijuengine.com/engine/systems/logging"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

var ConsoleArea = editor_areas.AreaType{
	Key:            klib.NewKey("com.kaiju.area_console", "Console"),
	HandlerFactory: func() editor_areas.AreaHandlerInterface { return &ConsoleHandler{} },
	Flags:          editor_areas.AreaCapabilityFlagDockable | editor_areas.AreaCapabilityFlagDiscoverable,
}

func init() {
	editor_areas.Register(ConsoleArea)
}

type ConsoleHandler struct {
	backer        *console.Console
	ed            editor_areas.EditorAreaInterface
	list          *ui.VirtualList
	tcd           *document.Element
	lineIndicator *document.Element
	smallMode     bool
	incoming      bool
	wasAtBottom   atomic.Bool
	evtid         logging.EventId
}

func (h *ConsoleHandler) Open(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
	h.backer = console.For(ed.Host())
	h.ed = ed
	content, _ := area.HTMLContainer(ed, "console.html", nil, nil)
	h.list = area.VirtualList(ed)
	tcd, ok := content.GetDocument().GetElementById("terminalContent")
	if !ok {
		panic("Failed to open terminal - terminal.html file doesn't contain a definition for 'terminalContent'")
	}
	prompt, ok := content.GetDocument().GetElementById("prompt")
	if !ok {
		panic("Failed to open terminal - terminal.html file doesn't contain a definition for 'prompt'")
	}
	h.lineIndicator, _ = content.GetDocument().GetElementById("lineIndicator")
	prompt.UI.AddEvent(ui.EventTypeSubmit, func() {
		h.runCommand(prompt.UI.ToInput(), ed.Host())
	})
	tcd.UIPanel.AddChild(h.list.Base())
	h.tcd = tcd
	h.list.SetDelegate(h)
	h.list.SetOverscan(2)
	h.list.SetRowHeightFunc(h.CalculateRowHeight)
	h.evtid = console.For(ed.Host()).AddOnMessageSent(func(msg string) {
		h.incoming = true
		if h.IsAtBottom() && !h.wasAtBottom.Load() {
			h.wasAtBottom.Store(true)
		}
	})
	h.list.Base().ToPanel().SetScrollY(matrix.FloatMax)
	h.list.Base().ToPanel().PanelData().OnAfterLayout.Add(func() {
		if h.wasAtBottom.Load() {
			h.list.Base().ToPanel().SetScrollY(matrix.FloatMax)
			if !h.incoming {
				h.wasAtBottom.Store(false)
			}
		}
	})
}

func (h *ConsoleHandler) Close(area *editor_areas.Area, ed editor_areas.EditorAreaInterface) {
	h.ed = nil
}

func (h *ConsoleHandler) FocusInterface(ed editor_areas.EditorAreaInterface) {
}

func (h *ConsoleHandler) BlurInterface(ed editor_areas.EditorAreaInterface) {
}

func (h *ConsoleHandler) Update(area *editor_areas.Area, ed editor_areas.EditorAreaInterface, deltaTime float64, posx, posy, width, height float32) {
	if h.list == nil {
		return
	}
	ps := h.tcd.UIPanel.Base().Layout().PixelSize()
	if ed.Host().Window.Mouse.ScrollY > 0 {
		h.wasAtBottom.Store(false)
	}
	if h.incoming {
		h.incoming = false
		h.checkIfTooSmall(ed, width)
		h.list.Base().Layout().Scale(ps.X(), ps.Y())
		h.list.ReloadData()
	} else if h.list.Base().Layout().Scale(ps.X(), ps.Y()) {
		h.checkIfTooSmall(ed, width)
		if h.IsAtBottom() {
			h.wasAtBottom.Store(true)
		}
		h.list.ReloadData()
	}
	if h.lineIndicator != nil {
		f, l := h.list.ViewportRange()
		ln := h.lineIndicator.UI.ToPanel().Child(0)
		if ln != nil && !ln.IsDisabled() {
			ln.ToLabel().SetText(fmt.Sprintf("line [%v - %v] of %v", f+1, l+1, h.RowCount()))
		}
	}
}

func (h *ConsoleHandler) runCommand(input *ui.Input, host *engine.Host) {
	cmd := strings.TrimSpace(input.Text())
	if len(cmd) <= 0 {
		return
	}
	if !input.Base().Host().Window.Keyboard.HasShift() {
		input.SetText("")
	}
	h.list.Content().SetScrollY(matrix.FloatMax)
	host.RunOnMainThread(func() {
		console.For(host).Execute(cmd)
	})
}

func (h *ConsoleHandler) checkIfTooSmall(ed editor_areas.EditorAreaInterface, width float32) {
	h.smallMode = ed.Theme().SizeTextGlobal*24 > width
}

func (h *ConsoleHandler) GetOverlayDimensions() (float32, float32) {
	return -1, -1
}

func (h *ConsoleHandler) IsAtBottom() bool {
	if h.list == nil {
		return true
	}
	panel := h.list.Base().ToPanel()
	return matrix.Abs(panel.ScrollY()-panel.MaxScroll().Y()) < 50
}

func (h *ConsoleHandler) RowCount() int {
	return h.backer.MessageCount()
}

func (h *ConsoleHandler) CreateRow(manager *ui.Manager) *ui.UI {
	container := manager.Add().ToPanel()
	container.Init(nil, ui.ElementTypePanel)
	container.Base().Layout().SetPadding(4, 1, 4, 1)
	container.Base().Layout().Stylizer = editor_areas.MaxWidthStylizer{Parent: weak.Make(h.list.Content().Base())}
	container.SetFlex()
	container.SetFlexDirection(ui.FlexDirectionRow)
	container.SetFlexAlignItems(ui.FlexAlignStart)
	container.DontFitContentWidth()
	fs := float32(12)
	if h.ed != nil {
		fs = h.ed.Theme().SizeTextGlobal
	}
	addField(manager, container, fs)
	addField(manager, container, fs)
	msgp := manager.Add().ToPanel()
	msgp.Init(nil, ui.ElementTypePanel)
	msgp.Base().Layout().SetFlexGrow(1)
	msgp.DontFitContentWidth()
	container.AddChild(msgp.Base())
	addField(manager, msgp, fs)
	return container.Base()
}

func addField(manager *ui.Manager, panel *ui.Panel, fs float32) *ui.Label {
	field := manager.Add().ToLabel()
	field.Init("")
	field.SetFontFace(rendering.FontMono)
	field.SetFontSize(fs)
	field.Base().Layout().SetMargin(4, 1, 4, 1)
	panel.AddChild(field.Base())
	return field
}

func (h *ConsoleHandler) BindRow(index int, row *ui.UI) {
	if h.ed == nil {
		return
	}
	dc := h.backer
	if index < 0 || index >= dc.MessageCount() {
		return
	}
	entry := dc.MessageAt(index)
	var color matrix.Color
	var lvl string
	switch entry.Level {
	case console.LogLevelInfo:
		color = h.ed.Theme().ColorConsoleInfo
		lvl = "[NFO]"
	case console.LogLevelWarn:
		color = h.ed.Theme().ColorConsoleWarn
		lvl = "[WRN]"
	case console.LogLevelError:
		color = h.ed.Theme().ColorConsoleError
		lvl = "[ERR]"
	case console.LogLevelDebug:
		color = h.ed.Theme().ColorTextPassive
		lvl = "[DBG]"
	}
	darkColor := matrix.ColorMix(color, matrix.ColorBlack(), 0.25)
	time := row.ToPanel().Child(0).ToLabel()
	time.SetText(entry.Time)
	time.SetColor(darkColor)
	level := row.ToPanel().Child(1).ToLabel()
	level.SetText(lvl)
	level.SetColor(darkColor)
	msg := row.ToPanel().Child(2).ToPanel().Child(0).ToLabel()
	msg.SetText(entry.Message)
	msg.SetColor(color)
	if index%2 == 0 {
		row.ToPanel().SetColor(matrix.ColorMix(h.ed.Theme().ColorPanel, matrix.ColorWhite(), 0.02))
	} else {
		row.ToPanel().SetColor(matrix.ColorMix(h.ed.Theme().ColorPanel, matrix.ColorBlack(), 0.1))
	}
	if h.smallMode {
		time.Base().SetVisibility(false)
		level.Base().SetVisibility(false)
	} else {
		time.Base().SetVisibility(true)
		level.Base().SetVisibility(true)
	}
	// dc.MessageAt(index).SetUI(row)
}

func (h *ConsoleHandler) UnbindRow(index int, row *ui.UI) {
	row.ToPanel().Child(0).ToLabel().SetText("")
	row.ToPanel().Child(1).ToLabel().SetText("")
	row.ToPanel().Child(2).ToPanel().Child(0).ToLabel().SetText("")
	// h.ed.Host().DevConsole().MessageAt[index].SetUI(nil)
}

func (h *ConsoleHandler) CalculateRowHeight(index int) float32 {
	entry := h.backer.MessageAt(index)
	// if entry.ui != nil {
	// 	return entry.ui.Layout().PixelSize().Y() + 4
	// }
	fs := h.ed.Theme().SizeTextGlobal
	width := h.list.Content().Base().Layout().PixelSize().X() - 28
	output := h.ed.Host().FontCache().MeasureStringWithinWithLetterSpacing(rendering.FontMono,
		entry.Time+"  [---]  "+entry.Message, fs, width, 0, 0)
	return output.Y() + 8
}
