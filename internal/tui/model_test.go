package tui

import "testing"

func TestClamp(t *testing.T) {
	tests := []struct {
		v, lo, hi, want int
	}{
		{5, 0, 10, 5},
		{-1, 0, 10, 0},
		{15, 0, 10, 10},
		{3, 3, 3, 3},
		{0, 1, 5, 1},
	}
	for _, tt := range tests {
		got := clamp(tt.v, tt.lo, tt.hi)
		if got != tt.want {
			t.Errorf("clamp(%d, %d, %d) = %d, want %d",
				tt.v, tt.lo, tt.hi, got, tt.want)
		}
	}
}

func TestTrunc(t *testing.T) {
	tests := []struct {
		s    string
		n    int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 5, "hell…"},
		{"hi", 0, ""},
		{"test", -1, ""},
		{"", 5, ""},
	}
	for _, tt := range tests {
		got := trunc(tt.s, tt.n)
		if got != tt.want {
			t.Errorf("trunc(%q, %d) = %q, want %q",
				tt.s, tt.n, got, tt.want)
		}
	}
}

func TestComputeDims_Width(t *testing.T) {
	tests := []struct {
		w              int
		wantCenterMaj  bool   // centerW > leftW && centerW > rightW
		wantExactFit   bool   // total rendered = w
		wantMinPanel   int    // minimum panel width >= 1
	}{
		{10, true, true, 1},
		{20, true, true, 1},
		{40, true, true, 1},
		{60, true, true, 1},
		{80, true, true, 1},
		{120, true, true, 1},
		{200, true, true, 1},
	}

	for _, tt := range tests {
		d := computeDims(tt.w, 24)

		if d.leftW < tt.wantMinPanel {
			t.Errorf("w=%d: leftW=%d < min %d", tt.w, d.leftW, tt.wantMinPanel)
		}
		if d.rightW < tt.wantMinPanel {
			t.Errorf("w=%d: rightW=%d < min %d", tt.w, d.rightW, tt.wantMinPanel)
		}
		if d.centerW < tt.wantMinPanel {
			t.Errorf("w=%d: centerW=%d < min %d", tt.w, d.centerW, tt.wantMinPanel)
		}

		if tt.wantCenterMaj && (d.centerW <= d.leftW || d.centerW <= d.rightW) {
			t.Errorf("w=%d: center=%d not majority vs left=%d right=%d",
				tt.w, d.centerW, d.leftW, d.rightW)
		}

		if tt.wantExactFit {
			total := d.leftW + 2 + d.centerW + 2 + d.rightW + 2
			if total != tt.w {
				t.Errorf("w=%d: total=%d (left=%d center=%d right=%d)",
					tt.w, total, d.leftW, d.centerW, d.rightW)
			}
		}
	}
}

func TestComputeDims_Height(t *testing.T) {
	tests := []struct {
		h             int
		wantFooter    bool
		wantBordered  bool
	}{
		{4, false, false},  // tiny: no footer, plain input
		{5, false, false},  // tiny: no footer, plain input
		{6, false, true},   // small: no footer, bordered input
		{7, false, true},   // small: no footer, bordered input
		{8, true, true},    // normal: footer + bordered input
		{24, true, true},   // standard: footer + bordered input
	}

	for _, tt := range tests {
		d := computeDims(80, tt.h)
		if d.showFooter != tt.wantFooter {
			t.Errorf("h=%d: showFooter=%v, want %v",
				tt.h, d.showFooter, tt.wantFooter)
		}
		isBordered := d.inputBarH > 1
		if isBordered != tt.wantBordered {
			t.Errorf("h=%d: inputBarH=%d (bordered=%v), want bordered=%v",
				tt.h, d.inputBarH, isBordered, tt.wantBordered)
		}

		// Verify total height fits.
		total := d.panelInnerH + 2 + d.inputBarH
		if d.showFooter {
			total++
		}
		if total > tt.h {
			t.Errorf("h=%d: total=%d exceeds terminal height", tt.h, total)
		}
	}
}

func TestComputeDims_SidePanelMax(t *testing.T) {
	// At very wide terminals, side panels should cap at maxPanelW.
	d := computeDims(400, 24)
	if d.leftW > maxPanelW {
		t.Errorf("leftW=%d exceeds maxPanelW=%d", d.leftW, maxPanelW)
	}
	if d.rightW > maxPanelW {
		t.Errorf("rightW=%d exceeds maxPanelW=%d", d.rightW, maxPanelW)
	}
}
