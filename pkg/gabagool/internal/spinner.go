package internal

import (
	"image"
	"image/color"
	"math"
)

// SpinnerImage draws a ring of the given size: a quarter arc in c over a
// track of c at a quarter of its opacity, anti-aliased by supersampling. It
// is drawn upright so it can be rotated about its centre.
func SpinnerImage(size int, c color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	outer := float64(size) / 2
	inner := outer * 0.62
	const samples = 4
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			var arc, track float64
			for sy := 0; sy < samples; sy++ {
				for sx := 0; sx < samples; sx++ {
					dx := float64(x) + (float64(sx)+0.5)/samples - outer
					dy := float64(y) + (float64(sy)+0.5)/samples - outer
					r := math.Hypot(dx, dy)
					if r < inner || r > outer {
						continue
					}
					// Angle clockwise from 12 o'clock: the arc runs from 12 to 3.
					if a := math.Atan2(dx, -dy); a >= 0 && a <= math.Pi/2 {
						arc++
					} else {
						track++
					}
				}
			}
			coverage := (arc + track*0.25) / (samples * samples)
			if coverage > 0 {
				img.SetNRGBA(x, y, color.NRGBA{R: c.R, G: c.G, B: c.B, A: uint8(math.Round(float64(c.A) * coverage))})
			}
		}
	}
	return img
}
