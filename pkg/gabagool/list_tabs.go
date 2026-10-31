package gabagool

import (
	"fmt"

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
		color := internal.MutedTextColor()
		accent := tab.Color
		if accent.A == 0 {
			accent = theme.AccentColor
		}
		if tab.Color.A > 0 {
			color = tab.Color
			color.A = 150
		}
		if i == selected {
			color = theme.TextColor
			if tab.Color.A > 0 {
				color = tab.Color
			}
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
// left/right, ending the list so its caller shows that tab. It reports
// whether the button was a tab switch, which a list without tabs never has.
func (lc *listController) switchTab(button constants.VirtualButton, running *bool, result *ListResult) bool {
	if len(lc.Options.Tabs) == 0 || lc.MultiSelect || lc.ReorderMode {
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
	next := lc.Options.SelectedTab + step
	if next < 0 || next >= len(lc.Options.Tabs) {
		return true // the end of the strip: nothing to move to, and no paging either
	}
	*running = false
	result.Action = ListActionTabChanged
	result.Tab = next
	if len(lc.Options.Items) > 0 {
		result.Selected = []int{lc.Options.SelectedIndex}
	}
	return true
}

// hasSubtitles is whether any row has a second line, which makes every row
// tall enough for one, so the rows stay even.
func (lc *listController) hasSubtitles() bool {
	for _, item := range lc.Options.Items {
		if item.Subtitle != "" {
			return true
		}
	}
	return false
}

func (lc *listController) subtitleHeight() int32 {
	return int32(internal.Fonts.TinyFont.Height())
}

// rowHeight is a row's height: a pill, taller when rows have second lines.
func (lc *listController) rowHeight() int32 {
	height := int32(60 * internal.GetScaleFactor())
	if lc.hasSubtitles() {
		height += lc.subtitleHeight()
	}
	return height
}

// renderItemDetail draws a row's detail at its right edge, centred on its
// first line, and returns the width it takes from the row's text.
func (lc *listController) renderItemDetail(renderer *sdl.Renderer, item MenuItem, itemY, pillHeight, pillWidth int32) int32 {
	scale := internal.GetScaleFactor()
	color := internal.MutedTextColor()
	if item.Focused {
		color = internal.GetTheme().HighlightedTextColor
		color.A = 170
	}
	texture, w, h := internal.RenderTextCached(renderer, internal.Fonts.TinyFont, item.Detail, color)
	if texture == nil {
		return 0
	}
	padding := int32(20 * scale)
	lineHeight := pillHeight
	if item.Subtitle != "" {
		lineHeight -= lc.subtitleHeight()
	}
	renderer.Copy(texture, nil, &sdl.Rect{X: lc.Options.Margins.Left + pillWidth - padding - w, Y: itemY + (lineHeight-h)/2, W: w, H: h})
	return w + padding
}

// renderItemSubtitle draws a row's second line at y, clipped to its width.
func (lc *listController) renderItemSubtitle(renderer *sdl.Renderer, item MenuItem, y, maxWidth int32) {
	color := internal.MutedTextColor()
	if item.Focused {
		color = internal.GetTheme().HighlightedTextColor
		color.A = 170
	}
	font := internal.Fonts.TinyFont
	text := lc.truncateText(font, item.Subtitle, maxWidth)
	texture, w, h := internal.RenderTextCached(renderer, font, text, color)
	if texture == nil {
		return
	}
	renderer.Copy(texture, nil, &sdl.Rect{X: lc.Options.Margins.Left + int32(20*internal.GetScaleFactor()), Y: y, W: w, H: h})
}

// fitRowsBelow sizes a tabbed list to the rows that fit between where they
// start and the footer, which calculateMaxVisibleItems only estimates: with
// a header and a tab strip above them its margin costs a row.
func (lc *listController) fitRowsBelow(renderer *sdl.Renderer, top int32) {
	scale := internal.GetScaleFactor()
	_, screenHeight, _ := renderer.GetOutputSize()
	footerTop := screenHeight - lc.Options.Margins.Bottom - int32(50*scale) - int32(10*scale)
	rows := max(1, int((footerTop-top)/(lc.rowHeight()+lc.Options.ItemSpacing)))
	lc.fittedRows = rows
	if rows == lc.Options.MaxVisibleItems {
		return
	}
	lc.Options.MaxVisibleItems = rows
	lc.scrollTo(lc.Options.SelectedIndex)
}
