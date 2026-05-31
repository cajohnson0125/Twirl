package tui

import "testing"

// --- computeDims ---

func TestComputeDims_ViewportContentWidth(t *testing.T) {
	for _, w := range []int{20, 40, 80, 120, 200} {
		d := computeDims(w, 24)
		if d.vpContentW != w-2 {
			t.Errorf("w=%d: vpContentW=%d, want %d",
				w, d.vpContentW, w-2)
		}
	}
}

func TestComputeDims_FooterPresence(t *testing.T) {
	tests := []struct {
		h    int
		want int // expected footerH
	}{
		{1, 0},
		{3, 0},
		{5, 0},
		{6, 1},
		{7, 1},
		{24, 1},
		{50, 1},
		{100, 1},
	}
	for _, tt := range tests {
		d := computeDims(80, tt.h)
		if d.footerH != tt.want {
			t.Errorf("h=%d: footerH=%d, want %d",
				tt.h, d.footerH, tt.want)
		}
	}
}

func TestComputeDims_TotalHeightFits(t *testing.T) {
	// For terminals tall enough to fit the minimum layout (info + vp +
	// input + footer), the total rendered height must not exceed the
	// terminal. Minimum that fits: infoH(3) + vpTotalH_min(4) +
	// inputH(3) + footerH(1) = 11 lines.
	for _, h := range []int{11, 12, 24, 50, 100} {
		d := computeDims(80, h)
		total := d.infoH + (d.vpContentH + 3) + d.inputH + d.footerH
		if total > h {
			t.Errorf("h=%d: total rendered=%d exceeds terminal",
				h, total)
		}
	}
}

func TestComputeDims_TinyOverflowAccepted(t *testing.T) {
	// For very small terminals, computeDims clamps the viewport to a
	// minimum height. The rendered layout may exceed the terminal —
	// that's expected. Just verify vpContentH >= 1 (never zero/neg).
	for _, h := range []int{1, 3, 5, 7, 9} {
		d := computeDims(80, h)
		if d.vpContentH < 1 {
			t.Errorf("h=%d: vpContentH=%d, want >= 1", h, d.vpContentH)
		}
	}
}

func TestComputeDims_FixedPanelHeights(t *testing.T) {
	d := computeDims(80, 24)
	if d.infoH != 3 {
		t.Errorf("infoH=%d, want 3", d.infoH)
	}
	if d.inputH != 3 {
		t.Errorf("inputH=%d, want 3", d.inputH)
	}
}

func TestComputeDims_WidthStored(t *testing.T) {
	for _, w := range []int{10, 40, 80, 200} {
		d := computeDims(w, 24)
		if d.w != w {
			t.Errorf("w=%d: d.w=%d, want %d", w, d.w, w)
		}
	}
}

func TestComputeDims_TinyTerminal(t *testing.T) {
	d := computeDims(20, 3)
	// Viewport should get minimum height, never negative.
	if d.vpContentH < 1 {
		t.Errorf("tiny terminal: vpContentH=%d, want >= 1",
			d.vpContentH)
	}
	if d.footerH != 0 {
		t.Errorf("tiny terminal: footerH=%d, want 0",
			d.footerH)
	}
}

func TestComputeDims_LargeTerminal(t *testing.T) {
	d := computeDims(300, 120)
	if d.footerH != 1 {
		t.Errorf("large terminal: footerH=%d, want 1",
			d.footerH)
	}
	if d.vpContentW != 298 {
		t.Errorf("large terminal: vpContentW=%d, want 298",
			d.vpContentW)
	}
	// Most of the height should go to the viewport.
	expectedVP := 120 - 3 - 3 - 1 - 3 // h - info - input - footer - vp border/title
	if d.vpContentH != expectedVP {
		t.Errorf("large terminal: vpContentH=%d, want %d",
			d.vpContentH, expectedVP)
	}
}

// --- countActive ---

func TestCountActive_AllIdle(t *testing.T) {
	agents := []agent{
		{name: "A", status: statusIdle},
		{name: "B", status: statusIdle},
	}
	if n := countActive(agents); n != 0 {
		t.Errorf("got %d, want 0", n)
	}
}

func TestCountActive_AllActive(t *testing.T) {
	agents := []agent{
		{name: "A", status: statusActive},
		{name: "B", status: statusActive},
	}
	if n := countActive(agents); n != 2 {
		t.Errorf("got %d, want 2", n)
	}
}

func TestCountActive_Mixed(t *testing.T) {
	agents := []agent{
		{name: "A", status: statusActive},
		{name: "B", status: statusIdle},
		{name: "C", status: statusActive},
		{name: "D", status: statusIdle},
	}
	if n := countActive(agents); n != 2 {
		t.Errorf("got %d, want 2", n)
	}
}

func TestCountActive_Empty(t *testing.T) {
	if n := countActive(nil); n != 0 {
		t.Errorf("got %d, want 0", n)
	}
}

func TestCountActive_SingleActive(t *testing.T) {
	agents := []agent{
		{name: "A", status: statusActive},
	}
	if n := countActive(agents); n != 1 {
		t.Errorf("got %d, want 1", n)
	}
}
