package main

import (
	"bytes"

	"github.com/LCmaster/coreutils-go/pkg/fn"
	"github.com/LCmaster/coreutils-go/pkg/strutils"
)

type LineTransformer = func([]byte) []byte

func newFormatPipeline(opts CatOptions) LineTransformer {
	var stages []LineTransformer
	if opts.ShowNonPrinting {
		stages = append(stages, strutils.FormatNonPrinting)
	}
	if opts.ShowTabs {
		stages = append(stages, formatTabs)
	}
	if opts.ShowEnds {
		stages = append(stages, formatEnds)
	}
	if len(stages) == 0 {
		return func(line []byte) []byte { return line }
	}
	return fn.Pipe(stages...)
}

func formatTabs(line []byte) []byte {
	return bytes.ReplaceAll(line, []byte{'\t'}, []byte("^I"))
}

func formatEnds(line []byte) []byte {
	return bytes.ReplaceAll(line, []byte{'\n'}, []byte("$\n"))
}
