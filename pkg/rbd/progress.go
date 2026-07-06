package rbd

import (
	"fmt"
	"io"
	"os"
)

// Progress reports long-running operation progress in the same format as native rbd CLI.
type Progress struct {
	// Operation is the label shown before the percentage, e.g. "Image migration".
	Operation string
	// Out is the writer used for progress output; stderr is used when nil.
	Out io.Writer
	// lastReportedPercent is the most recently printed completion percentage.
	lastReportedPercent int
}

// NewProgress creates a progress reporter. When out is nil, stderr is used.
func NewProgress(operation string, out io.Writer) *Progress {
	if out == nil {
		out = os.Stderr
	}
	return &Progress{Operation: operation, Out: out}
}

// Update prints progress when the percentage increases.
func (p *Progress) Update(percent int) {
	if p == nil || p.Out == nil {
		return
	}
	if percent > 100 {
		percent = 100
	}
	if percent <= p.lastReportedPercent {
		return
	}
	p.lastReportedPercent = percent
	fmt.Fprintf(p.Out, "\r%s: %d%% complete...", p.Operation, percent)
}

// Finish prints the final success line.
func (p *Progress) Finish() {
	if p == nil || p.Out == nil {
		return
	}
	fmt.Fprintf(p.Out, "\r%s: 100%% complete...done.\n", p.Operation)
}

// Fail prints the final failure line.
func (p *Progress) Fail() {
	if p == nil || p.Out == nil {
		return
	}
	fmt.Fprintf(p.Out, "\r%s: %d%% complete...failed.\n", p.Operation, p.lastReportedPercent)
}
