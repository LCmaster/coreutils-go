package main

import "bytes"

type LineTransformer func([]byte) []byte

func Pipe(transformers ...LineTransformer) LineTransformer {
	return func(line []byte) []byte {
		for _, transform := range transformers {
			line = transform(line)
		}
		return line
	}
}

func newFormatPipeline(opts Options) LineTransformer {
	var stages []LineTransformer
	if opts.ShowNonPrinting {
		stages = append(stages, formatNonPrinting)
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

func formatNonPrinting(line []byte) []byte {
	result := make([]byte, 0, len(line))
	for _, b := range line {
		switch {
		case b == '\n' || b == '\t':
			result = append(result, b)
		case b < 32:
			result = append(result, '^', b+64)
		case b < 127:
			result = append(result, b)
		case b == 127:
			result = append(result, '^', '?')
		case b < 160:
			result = append(result, 'M', '-', '^', b-128+64)
		case b < 255:
			result = append(result, 'M', '-', b-128)
		default: // b == 255
			result = append(result, 'M', '-', '^', '?')
		}
	}
	return result
}

func isEmptyLine(line []byte) bool {
	return len(line) == 0 || (len(line) == 1 && line[0] == '\n') || (len(line) == 2 && line[0] == '\r' && line[1] == '\n')
}
