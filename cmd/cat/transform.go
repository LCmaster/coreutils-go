package main

import (
	"bytes"

	"github.com/LCmaster/coreutils-go/pkg/fn"
	"github.com/LCmaster/coreutils-go/pkg/strutils"
)

type LineTransformer func([]byte) []byte

func Pipe(transformers ...LineTransformer) LineTransformer {
	transforms := make([]func([]byte) []byte, len(transformers))
	for i, t := range transformers {
		transforms[i] = t
	}
	return fn.Pipe(transforms...)
}

func newFormatPipeline(opts Options) LineTransformer {
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
		return func(line []byte) []byte { return line } // Identity
	}
	return Pipe(stages...)
}

func formatTabs(line []byte) []byte {
	return bytes.ReplaceAll(line, []byte{'\t'}, []byte("^I"))
}

func formatEnds(line []byte) []byte {
	return bytes.ReplaceAll(line, []byte{'\n'}, []byte("$\n"))
}
