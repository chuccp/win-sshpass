package main

import (
	"fmt"
	"io"
	"sync"

	"github.com/schollz/progressbar/v3"
)

// cliProgress wraps a schollz/progressbar instance and adapts it to the SDK's
// ProgressFunc callback. Each transfer creates its own bar; the callback
// updates the bar with the cumulative byte count.
type cliProgress struct {
	mu   sync.Mutex
	out  io.Writer
	bar  *progressbar.ProgressBar
	desc string
}

func newCLIProgress(out io.Writer) *cliProgress {
	return &cliProgress{out: out}
}

func (p *cliProgress) progress(description string, sent, total int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// A new transfer starts: close the previous bar (if any) and create a new
	// one for this file.
	if p.bar == nil || p.desc != description {
		if p.bar != nil {
			_ = p.bar.Close()
		}
		p.desc = description
		p.bar = progressbar.NewOptions64(
			total,
			progressbar.OptionSetDescription(description),
			progressbar.OptionSetWriter(p.out),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(40),
			progressbar.OptionThrottle(65),
			progressbar.OptionShowCount(),
			progressbar.OptionOnCompletion(func() { fmt.Fprint(p.out, "\n") }),
			progressbar.OptionFullWidth(),
			progressbar.OptionUseANSICodes(true),
		)
		return
	}

	// Update existing bar to the cumulative byte count.
	p.bar.Set64(sent)
}
