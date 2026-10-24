package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/LCmaster/coreutils-go/pkg/fsutils"
	"github.com/LCmaster/coreutils-go/pkg/ioutils"
	"github.com/LCmaster/coreutils-go/pkg/strutils"
)

const bufferSize = ioutils.DefaultBufferSize // Memory buffer for optimal I/O throughput

type Options struct {
	NumberNonBlank  bool
	ShowEnds        bool
	Number          bool
	SqueezeBlank    bool
	ShowTabs        bool
	ShowNonPrinting bool
	Unbuffered      bool
}

func (o Options) NeedsFormatting() bool {
	return o.NumberNonBlank || o.ShowEnds || o.Number ||
		o.SqueezeBlank || o.ShowTabs || o.ShowNonPrinting
}

type StreamState struct {
	LineNumber int
	PrevBlank  bool
}

func copyStream(out io.Writer, r io.Reader) error {
	_, err := io.Copy(out, r)
	return err
}

func processFile(filename string, stdin io.Reader, stdout io.Writer, opts Options, state *StreamState, pipeline LineTransformer, w *bufio.Writer) error {
	var r io.Reader = stdin
	if filename != "-" {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		stat, err := f.Stat()
		if err == nil && stat.IsDir() {
			return fmt.Errorf("Is a directory")
		}

		if fsutils.IsSameAsOutput(f, stdout) {
			return fmt.Errorf("input file is also the output")
		}
		r = f
	}

	if opts.NeedsFormatting() {
		return processStream(w, r, opts, state, pipeline)
	}

	var out io.Writer = w
	if opts.Unbuffered {
		_ = w.Flush()
		out = stdout
	}
	return copyStream(out, r)
}

func processStream(w *bufio.Writer, r io.Reader, opts Options, state *StreamState, transform LineTransformer) error {
	reader := bufio.NewReaderSize(r, bufferSize)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF && len(line) == 0 {
			break
		}
		if err != nil && err != io.EOF {
			return err
		}

		if opts.SqueezeBlank {
			if state.PrevBlank && strutils.IsEmptyLine(line) {
				if err == io.EOF {
					break
				}
				continue
			}
			state.PrevBlank = strutils.IsEmptyLine(line)
		}

		if opts.NumberNonBlank {
			if strutils.IsEmptyLine(line) {
				if _, writeErr := w.Write(line); writeErr != nil {
					return writeErr
				}
				if opts.Unbuffered {
					if flushErr := w.Flush(); flushErr != nil {
						return flushErr
					}
				}
				if err == io.EOF {
					break
				}
				continue
			}
			if _, writeErr := fmt.Fprintf(w, "%6d\t", state.LineNumber); writeErr != nil {
				return writeErr
			}
		} else if opts.Number {
			if _, writeErr := fmt.Fprintf(w, "%6d\t", state.LineNumber); writeErr != nil {
				return writeErr
			}
		}

		line = transform(line)

		if _, writeErr := w.Write(line); writeErr != nil {
			return writeErr
		}
		if opts.Unbuffered {
			if flushErr := w.Flush(); flushErr != nil {
				return flushErr
			}
		}
		state.LineNumber++

		if err == io.EOF {
			break
		}
	}
	return nil
}
