package gabagool

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/internal"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// ListTab is one tab of a list's tab strip: its label, a count drawn after
// it when there is one, and a colour that marks it out (the theme's accent
// when unset).
type ListTab struct {
	Label string
	Count int
	Color sdl.Color
}

func listTabsFont() *ttf.Font {
	return internal.Fonts.TinyFont
}

// listTabsHeight is the strip's height: a label, the gap to its underline
// and the underline.
func listTabsHeight() int32 {
	scale := internal.GetScaleFactor()
	return int32(listTabsFont().Height()) + int32(18*scale)
}

// renderListTabs draws the strip at y and returns the y below it: an L1
// pill, the tabs, and an R1 pill at the right edge. The shown tab is in the
// text colour and underlined; the others are muted.
func renderListTabs(renderer *sdl.Renderer, tabs []ListTab, selected int, x, y, width int32) int32 {
	scale := internal.GetScaleFactor()
	labelFont := listTabsFont()
	height := listTabsHeight()
	theme := internal.GetTheme()

	pill := func(text string, px int32, alignRight bool) int32 {
		texture, w, h := internal.RenderTextCached(renderer, internal.Fonts.MicroFont, text, theme.ButtonLabelColor)
		if texture == nil {
			return 0
		}
		padX, padY := int32(10*scale), int32(4*scale)
		pw, ph := w+2*padX, h+2*padY
		if alignRight {
			px -= pw
		}
		py := y + (height-int32(6*scale)-ph)/2
		internal.DrawRoundedRect(renderer, &sdl.Rect{X: px, Y: py, W: pw, H: ph}, ph/2, theme.AccentColor)
		renderer.Copy(texture, nil, &sdl.Rect{X: px + padX, Y: py + padY, W: w, H: h})
		return pw
	}

	cursor := x + pill("L1", x, false) + int32(8*scale)
	right := x + width - pill("R1", x+width, true) - int32(8*scale)
	gap := int32(12 * scale)
	underline := int32(4 * scale)
	for i, tab := range tabs {
		accent := cmp.Or(tab.Color, theme.AccentColor)
		color := internal.MutedTextColor()
		switch {
		case i == selected && tab.Color.A > 0:
			color = tab.Color
		case i == selected:
			color = theme.TextColor
		case tab.Color.A > 0:
			color = tab.Color
			color.A = internal.MutedTextColor().A
		}
		text := tab.Label
		if tab.Count > 0 {
			text = fmt.Sprintf("%s %d", tab.Label, tab.Count)
		}
		texture, w, h := internal.RenderTextCached(renderer, labelFont, text, color)
		if texture == nil {
			continue
		}
		if cursor+w+2*gap > right {
			break // the rest don't fit; the shown tab comes first in practice
		}
		renderer.Copy(texture, nil, &sdl.Rect{X: cursor + gap, Y: y, W: w, H: h})
		if i == selected {
			internal.DrawRoundedRect(renderer, &sdl.Rect{X: cursor, Y: y + height - underline, W: w + 2*gap, H: underline}, underline/2, accent)
		}
		cursor += w + 2*gap
	}
	return y + height
}

// switchTab moves to the tab before or after the shown one on L1/R1 or
// left/right, in place: the list's OnTabChange gives the tab's items, which
// show on the next frame. It reports whether the button switches tabs,
// which it never does in a list without them.
func (lc *listController) switchTab(button constants.VirtualButton) bool {
	if len(lc.Options.Tabs) == 0 || lc.Options.OnTabChange == nil || lc.MultiSelect || lc.ReorderMode {
		return false
	}
	step := 0
	switch button {
	case constants.VirtualButtonL1, constants.VirtualButtonLeft:
		step = -1
	case constants.VirtualButtonR1, constants.VirtualButtonRight:
		step = 1
	default:
		return false
	}
	// At the end of the strip there is nothing to move to, and no paging.
	if next := lc.Options.SelectedTab + step; next >= 0 && next < len(lc.Options.Tabs) {
		lc.pendingTab = &next
	}
	return true
}

// showPendingTab shows the tab switchTab moved to, its cursor on its first
// row.
func (lc *listController) showPendingTab(window *internal.Window) {
	if lc.pendingTab == nil {
		return
	}
	update := lc.Options.OnTabChange(*lc.pendingTab)
	lc.pendingTab = nil
	lc.Options.SelectedIndex, lc.Options.VisibleStartIndex = 0, 0
	lc.applyUpdate(window, update, nil)
}

// hasSubtitles is whether any row has a second line.
func hasSubtitles(items []MenuItem) bool {
	return slices.ContainsFunc(items, func(item MenuItem) bool { return item.Subtitle != "" })
}

func (lc *listController) subtitleHeight() int32 {
	return int32(internal.Fonts.TinyFont.Height())
}

// rowHeight is a row's height: a pill, taller when rows have second lines,
// so the rows stay even.
func (lc *listController) rowHeight() int32 {
	height := int32(60 * internal.GetScaleFactor())
	if lc.subtitled {
		height += lc.subtitleHeight()
	}
	return height
}

// secondaryTextColor is a row's detail and second line: muted, or the
// highlighted text's colour, softened, on the focused pill.
func secondaryTextColor(item MenuItem) sdl.Color {
	if !item.Focused {
		return internal.MutedTextColor()
	}
	color := internal.GetTheme().HighlightedTextColor
	color.A = 170
	return color
}

// renderItemDetail draws a row's detail at its right edge, centred on its
// first line, and returns the width it takes from the row's text.
func (lc *listController) renderItemDetail(renderer *sdl.Renderer, item MenuItem, itemY, lineHeight, pillWidth int32) int32 {
	texture, w, h := internal.RenderTextCached(renderer, internal.Fonts.TinyFont, item.Detail, secondaryTextColor(item))
	if texture == nil {
		return 0
	}
	padding := int32(20 * internal.GetScaleFactor())
	renderer.Copy(texture, nil, &sdl.Rect{X: lc.Options.Margins.Left + pillWidth - padding - w, Y: itemY + (lineHeight-h)/2, W: w, H: h})
	return w + padding
}

// renderItemSubtitle draws a row's second line at y, cut to its width once
// rather than every frame.
func (lc *listController) renderItemSubtitle(renderer *sdl.Renderer, item MenuItem, y, maxWidth int32) {
	font := internal.Fonts.TinyFont
	key := fmt.Sprintf("%d\x00%s", maxWidth, item.Subtitle)
	text, ok := lc.subtitleCuts[key]
	if !ok {
		text = lc.truncateText(font, item.Subtitle, maxWidth)
		lc.subtitleCuts[key] = text
	}
	texture, w, h := internal.RenderTextCached(renderer, font, text, secondaryTextColor(item))
	if texture == nil {
		return
	}
	renderer.Copy(texture, nil, &sdl.Rect{X: lc.Options.Margins.Left + int32(20*internal.GetScaleFactor()), Y: y, W: w, H: h})
}

// fitRowsBelow sizes a tabbed list to the rows that fit between where they
// start and the footer, which calculateMaxVisibleItems only estimates: with
// a header and a tab strip above them its margin costs a row.
func (lc *listController) fitRowsBelow(top int32) {
	gap := int32(10 * internal.GetScaleFactor())
	rows := max(1, int((footerY(lc.Options.Margins.Bottom)-gap-top)/(lc.rowHeight()+lc.Options.ItemSpacing)))
	lc.fittedRows = rows
	if rows == lc.Options.MaxVisibleItems {
		return
	}
	lc.Options.MaxVisibleItems = rows
	lc.scrollTo(lc.Options.SelectedIndex)
}
