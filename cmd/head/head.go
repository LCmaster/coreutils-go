package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/LCmaster/coreutils-go/pkg/ioutils"
)

var suffixes = []struct {
	suffix     string
	multiplier int64
}{
	{"KiB", 1024},
	{"MiB", 1024 * 1024},
	{"GiB", 1024 * 1024 * 1024},
	{"TiB", 1024 * 1024 * 1024 * 1024},
	{"PiB", 1024 * 1024 * 1024 * 1024 * 1024},
	{"EiB", 1024 * 1024 * 1024 * 1024 * 1024 * 1024},

	{"kB", 1000},
	{"KB", 1000},
	{"MB", 1000 * 1000},
	{"GB", 1000 * 1000 * 1000},
	{"TB", 1000 * 1000 * 1000 * 1000},
	{"PB", 1000 * 1000 * 1000 * 1000 * 1000},
	{"EB", 1000 * 1000 * 1000 * 1000 * 1000 * 1000},

	{"b", 512},
	{"k", 1024},
	{"K", 1024},
	{"m", 1024 * 1024},
	{"M", 1024 * 1024},
	{"g", 1024 * 1024 * 1024},
	{"G", 1024 * 1024 * 1024},
	{"t", 1024 * 1024 * 1024 * 1024},
	{"T", 1024 * 1024 * 1024 * 1024},
	{"p", 1024 * 1024 * 1024 * 1024 * 1024},
	{"P", 1024 * 1024 * 1024 * 1024 * 1024},
	{"e", 1024 * 1024 * 1024 * 1024 * 1024 * 1024},
	{"E", 1024 * 1024 * 1024 * 1024 * 1024 * 1024},
}

func parseNum(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty number")
	}
	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	} else if strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	if s == "" {
		return 0, fmt.Errorf("missing number")
	}

	var mult int64 = 1
	for _, entry := range suffixes {
		if strings.HasSuffix(s, entry.suffix) {
			mult = entry.multiplier
			s = s[:len(s)-len(entry.suffix)]
			break
		}
	}

	if s == "" {
		return 0, fmt.Errorf("missing number digits")
	}

	base, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	val := base * mult
	if neg {
		val = -val
	}
	return int(val), nil
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
