package internal

import (
	"image/color"
	"testing"
)

// The arc is opaque, the track faint, and the centre and corners empty, so
// rotating the square about its centre turns a ring that doesn't wobble.
func TestSpinnerImage_ArcOverTrackAroundAnEmptyCentre(t *testing.T) {
	img := SpinnerImage(40, color.NRGBA{A: 255})
	alpha := func(x, y int) uint8 { return img.NRGBAAt(x, y).A }
	if a := alpha(32, 8); a < 200 {
		t.Errorf("arc between 12 and 3 o'clock has alpha %d, want it opaque", a)
	}
	if a := alpha(8, 32); a == 0 || a > 100 {
		t.Errorf("track has alpha %d, want it faint but drawn", a)
	}
	if alpha(20, 20) != 0 || alpha(0, 0) != 0 {
		t.Error("centre or corner is drawn, want both empty")
	}
}
