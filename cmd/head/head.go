package main

import (
	"bufio"
	"io"

	"github.com/LCmaster/coreutils-go/pkg/ioutils"
	"github.com/LCmaster/coreutils-go/pkg/strutils"
)

func parseNum(s string) (int, error) {
	return strutils.ParseByteSize(s)
}

func copyFirstLines(w io.Writer, r io.Reader, delimiter byte, count int) error {
	if count <= 0 {
		return nil
	}
	reader := bufio.NewReader(r)
	for range count {
		line, err := reader.ReadBytes(delimiter)
		if len(line) > 0 {
			if _, wErr := w.Write(line); wErr != nil {
				return wErr
			}
		}
		if err != nil {
			break
		}
	}
	return nil
}

func copyAllExceptLastLines(w io.Writer, r io.Reader, delimiter byte, k int) error {
	if k <= 0 {
		_, err := io.Copy(w, r)
		return err
	}

	reader := bufio.NewReader(r)
	ring := make([][]byte, k)
	head := 0
	count := 0

	for {
		line, err := reader.ReadBytes(delimiter)
		if len(line) > 0 {
			lineCopy := make([]byte, len(line))
			copy(lineCopy, line)
			if count == k {
				if _, wErr := w.Write(ring[head]); wErr != nil {
					return wErr
				}
				ring[head] = lineCopy
				head = (head + 1) % k
			} else {
				ring[(head+count)%k] = lineCopy
				count++
			}
		}
		if err != nil {
			break
		}
	}
	return nil
}

func copyFirstBytes(w io.Writer, r io.Reader, count int) error {
	if count <= 0 {
		return nil
	}
	_, err := io.CopyN(w, r, int64(count))
	if err == io.EOF {
		return nil
	}
	return err
}

func copyAllExceptLastBytes(w io.Writer, r io.Reader, k int) error {
	if k <= 0 {
		_, err := io.Copy(w, r)
		return err
	}

	if seeker, ok := r.(io.ReadSeeker); ok {
		curr, err := seeker.Seek(0, io.SeekCurrent)
		if err == nil {
			end, err := seeker.Seek(0, io.SeekEnd)
			if err == nil {
				_, err = seeker.Seek(curr, io.SeekStart)
				if err == nil {
					total := end - curr
					toCopy := total - int64(k)
					if toCopy > 0 {
						_, err := io.CopyN(w, seeker, toCopy)
						if err != nil && err != io.EOF {
							return err
						}
					}
					return nil
				}
			}
		}
	}

	// Non-seekable stream (e.g. pipe or stdin)
	ring := make([]byte, k)
	head := 0
	count := 0
	buf := make([]byte, ioutils.DefaultBufferSize)

	bw, hasByteWriter := w.(io.ByteWriter)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			for _, b := range buf[:n] {
				if count == k {
					if hasByteWriter {
						if err := bw.WriteByte(ring[head]); err != nil {
							return err
						}
					} else {
						if _, err := w.Write([]byte{ring[head]}); err != nil {
							return err
						}
					}
					ring[head] = b
					head = (head + 1) % k
				} else {
					ring[(head+count)%k] = b
					count++
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func headFile(r io.Reader, w io.Writer, byteMode bool, bytes int, lines int, delimiter byte) error {
	if byteMode {
		if bytes >= 0 {
			return copyFirstBytes(w, r, bytes)
		}
		return copyAllExceptLastBytes(w, r, -bytes)
	}
	if lines >= 0 {
		return copyFirstLines(w, r, delimiter, lines)
	}
	return copyAllExceptLastLines(w, r, delimiter, -lines)
}
