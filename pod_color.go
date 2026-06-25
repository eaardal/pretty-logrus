package main

import "github.com/fatih/color"

// PodColorizer assigns a distinct color to each pod ID it encounters, in order
// of first appearance, so logs interleaved from multiple pods (via
// `kubectl logs -l <selector> --prefix`) are visually separable. Colors are
// reused (wrapped) once the palette is exhausted.
//
// A PodColorizer is not safe for concurrent use; it is owned by the single
// printer goroutine.
type PodColorizer struct {
	palette  []*color.Color
	assigned map[string]*color.Color
	next     int
}

func newPodColorizer() *PodColorizer {
	return &PodColorizer{
		palette:  defaultPodPalette(),
		assigned: make(map[string]*color.Color),
	}
}

// colorFor returns the color assigned to podID, assigning the next palette color
// the first time a pod is seen.
func (p *PodColorizer) colorFor(podID string) *color.Color {
	if c, ok := p.assigned[podID]; ok {
		return c
	}

	c := p.palette[p.next%len(p.palette)]
	p.assigned[podID] = c
	p.next++
	return c
}

// Colorize renders the pod ID as a bracketed, colored prefix segment.
func (p *PodColorizer) Colorize(podID string) string {
	return p.colorFor(podID).Sprintf("[%s]", podID)
}

// defaultPodPalette is the set of foreground colors cycled through when
// assigning colors to pods. Level/timestamp/message colors are deliberately
// avoided where possible to keep the pod segment distinguishable.
func defaultPodPalette() []*color.Color {
	return []*color.Color{
		color.New(color.FgHiCyan),
		color.New(color.FgHiMagenta),
		color.New(color.FgHiYellow),
		color.New(color.FgHiBlue),
		color.New(color.FgHiGreen),
		color.New(color.FgHiRed),
		color.New(color.FgCyan),
		color.New(color.FgMagenta),
	}
}
