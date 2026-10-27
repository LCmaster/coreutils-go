package strutils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var sizeSuffixes = []struct {
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

// ParseByteSize parses a signed or unsigned integer string with an optional GNU coreutils
// multiplier suffix (e.g. b=512, kB/KB=1000, K/k/KiB=1024, MB=10^6, M/m/MiB=2^20, etc.).
// It is shared across utilities such as head, tail, split, and truncate.
func ParseByteSize(s string) (int, error) {
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
	for _, entry := range sizeSuffixes {
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

	if base > 0 && mult > math.MaxInt64/base {
		return 0, fmt.Errorf("value too large")
	}

	val := base * mult
	if neg {
		val = -val
	}
	return int(val), nil
}

