package main

import (
	"strings"
	"testing"

	"github.com/fatih/color"
)

func newTestPodColorizer() *PodColorizer {
	return &PodColorizer{
		palette: []*color.Color{
			color.New(color.FgRed),
			color.New(color.FgGreen),
			color.New(color.FgBlue),
		},
		assigned: make(map[string]*color.Color),
	}
}

func TestPodColorizer_AssignsSameColorToSamePod(t *testing.T) {
	colorizer := newTestPodColorizer()

	first := colorizer.colorFor("svc-1")
	second := colorizer.colorFor("svc-1")

	if first != second {
		t.Errorf("expected the same pod to keep its color, got two different colors")
	}
}

func TestPodColorizer_AssignsPaletteColorsInOrderOfFirstAppearance(t *testing.T) {
	colorizer := newTestPodColorizer()

	got := []*color.Color{
		colorizer.colorFor("svc-1"),
		colorizer.colorFor("svc-2"),
		colorizer.colorFor("svc-3"),
	}

	for i, c := range got {
		if c != colorizer.palette[i] {
			t.Errorf("pod %d got palette color %v, want palette[%d]", i, c, i)
		}
	}
}

func TestPodColorizer_WrapsAroundWhenPaletteExhausted(t *testing.T) {
	colorizer := newTestPodColorizer()

	colorizer.colorFor("svc-1")
	colorizer.colorFor("svc-2")
	colorizer.colorFor("svc-3")
	fourth := colorizer.colorFor("svc-4")

	if fourth != colorizer.palette[0] {
		t.Errorf("expected the fourth distinct pod to wrap to palette[0]")
	}
}

func TestPodColorizer_ColorizeIncludesBracketedPodID(t *testing.T) {
	colorizer := newTestPodColorizer()

	got := colorizer.Colorize("svc-1")

	if !strings.Contains(got, "[svc-1]") {
		t.Errorf("Colorize(%q) = %q, want it to contain %q", "svc-1", got, "[svc-1]")
	}
}
