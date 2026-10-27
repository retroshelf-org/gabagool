package internal

import "testing"

// Rows 0 and 3 are headings the cursor steps over.
func headings(i int) bool { return i != 0 && i != 3 }

func TestNextSelectable_StepsOverHeadingsInTheDirectionOfTravel(t *testing.T) {
	if got := NextSelectable(6, headings, 3, 1, true); got != 4 {
		t.Errorf("down onto a heading = %d, want the row after it", got)
	}
	if got := NextSelectable(6, headings, 3, -1, true); got != 2 {
		t.Errorf("up onto a heading = %d, want the row before it", got)
	}
}

func TestNextSelectable_WrapsPastAHeadingAtTheTop(t *testing.T) {
	if got := NextSelectable(6, headings, 0, -1, true); got != 5 {
		t.Errorf("up from the top heading = %d, want the last row", got)
	}
	if got := NextSelectable(6, headings, 6, 1, true); got != 1 {
		t.Errorf("down past the end = %d, want the first row under the heading", got)
	}
}

func TestNextSelectable_StopsAtTheEndsWithoutWrapping(t *testing.T) {
	last := func(i int) bool { return i < 4 }
	if got := NextSelectable(6, last, 4, 1, false); got != -1 {
		t.Errorf("forward from a trailing heading = %d, want none", got)
	}
}

// A list of nothing but headings must not loop forever.
func TestNextSelectable_NoneSelectable(t *testing.T) {
	if got := NextSelectable(4, func(int) bool { return false }, 0, 1, true); got != -1 {
		t.Errorf("got %d, want -1", got)
	}
}
