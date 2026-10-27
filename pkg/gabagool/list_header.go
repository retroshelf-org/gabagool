package gabagool

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"time"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/internal"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// ListHeader is a status block drawn under a list's title and above its
// items: an optional icon in a filled circle, a title line and smaller lines
// below it, on an optional panel. Items start below it.
type ListHeader struct {
	Icon            string    // A glyph drawn in the circle; no circle when empty and not Spinning
	Spinning        bool      // Draws a turning ring in the circle instead of Icon, for work in progress
	IconColor       sdl.Color // The circle's fill
	IconTextColor   sdl.Color // The glyph's colour; the theme's button label colour when unset
	Title           string
	Lines           []string  // Each wraps to the panel's width
	BackgroundColor sdl.Color // The panel's fill; no panel when its alpha is 0
}

// listHeaderLayout measures a header at a given width.
type listHeaderLayout struct {
	padding, iconSize, gap, textWidth, height int32
}

func headerFonts() (title, lines *ttf.Font) {
	return internal.Fonts.MediumFont, internal.Fonts.TinyFont
}

func layoutListHeader(header *ListHeader, width int32) listHeaderLayout {
	scale := internal.GetScaleFactor()
	titleFont, linesFont := headerFonts()
	l := listHeaderLayout{padding: int32(16 * scale), gap: int32(16 * scale)}
	if header.Icon != "" || header.Spinning {
		l.iconSize = int32(float32(titleFont.Height()) * 1.8)
	}
	l.textWidth = width - 2*l.padding
	if l.iconSize > 0 {
		l.textWidth -= l.iconSize + l.gap
	}
	textHeight := int32(titleFont.Height())
	for _, line := range header.Lines {
		textHeight += int32(4*scale) + wrappedHeight(line, linesFont, l.textWidth)
	}
	l.height = max(l.iconSize, textHeight) + 2*l.padding
	return l
}

// wrappedHeight is the height RenderMultilineText draws text at: each line
// is the font's height apart, plus its 5px leading.
func wrappedHeight(text string, font *ttf.Font, width int32) int32 {
	lineHeight := int32(font.Height())
	lines := internal.MultilineTextHeight(text, font, width) / max(lineHeight, 1)
	if lines == 0 {
		return 0
	}
	return lines*(lineHeight+5) - 5
}

// spinnerTurn is how long the ring takes to go round once.
const spinnerTurn = 900 * time.Millisecond

// renderListHeader draws a header at y and returns the y below it. cache
// keeps the spinner's ring between frames.
func renderListHeader(renderer *sdl.Renderer, header *ListHeader, x, y, width int32, cache *internal.TextureCache) int32 {
	l := layoutListHeader(header, width)
	scale := internal.GetScaleFactor()
	if header.BackgroundColor.A > 0 {
		internal.DrawRoundedRect(renderer, &sdl.Rect{X: x, Y: y, W: width, H: l.height}, int32(14*scale), header.BackgroundColor)
	}
	titleFont, linesFont := headerFonts()
	textX := x + l.padding
	if l.iconSize > 0 {
		iconY := y + (l.height-l.iconSize)/2
		internal.DrawRoundedRect(renderer, &sdl.Rect{X: textX, Y: iconY, W: l.iconSize, H: l.iconSize}, l.iconSize/2, header.IconColor)
		glyphColor := header.IconTextColor
		if glyphColor.A == 0 {
			glyphColor = internal.GetTheme().ButtonLabelColor
		}
		if header.Spinning {
			renderSpinner(renderer, cache, glyphColor, textX+l.iconSize/2, iconY+l.iconSize/2, l.iconSize*52/100)
		} else if texture, w, h := internal.RenderTextCached(renderer, titleFont, header.Icon, glyphColor); texture != nil {
			renderer.Copy(texture, nil, &sdl.Rect{X: textX + (l.iconSize-w)/2, Y: iconY + (l.iconSize-h)/2, W: w, H: h})
		}
		textX += l.iconSize + l.gap
	}
	textHeight := l.height - 2*l.padding
	contentHeight := int32(titleFont.Height())
	for _, line := range header.Lines {
		contentHeight += int32(4*scale) + wrappedHeight(line, linesFont, l.textWidth)
	}
	textY := y + l.padding + (textHeight-contentHeight)/2
	if texture, w, h := internal.RenderTextCached(renderer, titleFont, header.Title, internal.GetTheme().TextColor); texture != nil {
		// A title wider than the panel is clipped, not squeezed.
		w = internal.Min32(w, l.textWidth)
		renderer.Copy(texture, &sdl.Rect{W: w, H: h}, &sdl.Rect{X: textX, Y: textY, W: w, H: h})
	}
	textY += int32(titleFont.Height())
	for _, line := range header.Lines {
		textY += int32(4 * scale)
		internal.RenderMultilineText(renderer, line, linesFont, l.textWidth, textX, textY, internal.MutedTextColor(), constants.TextAlignLeft)
		textY += wrappedHeight(line, linesFont, l.textWidth)
	}
	return y + l.height
}

// renderSpinner draws the ring centred on (cx, cy), turned by the clock, so
// it keeps moving as the list redraws each frame.
func renderSpinner(renderer *sdl.Renderer, cache *internal.TextureCache, c sdl.Color, cx, cy, size int32) {
	key := fmt.Sprintf("spinner:%d:%d,%d,%d,%d", size, c.R, c.G, c.B, c.A)
	texture := cache.Get(key)
	if texture == nil {
		var data bytes.Buffer
		if err := png.Encode(&data, internal.SpinnerImage(int(size), color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A})); err != nil {
			return
		}
		var err error
		if texture, err = loadImageTexture(renderer, data.Bytes(), size, size); err != nil {
			return
		}
		cache.Set(key, texture)
	}
	angle := float64(time.Now().UnixNano()%int64(spinnerTurn)) / float64(spinnerTurn) * 360
	dst := sdl.Rect{X: cx - size/2, Y: cy - size/2, W: size, H: size}
	renderer.CopyEx(texture, nil, &dst, angle, &sdl.Point{X: size / 2, Y: size / 2}, sdl.FLIP_NONE)
}
