package strutils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const unitPowers = "kmgtpe"

func matchSuffixMultiplier(s string) (trimmed string, mult int64) {
	if strings.HasSuffix(s, "b") && !strings.HasSuffix(s, "ib") && !strings.HasSuffix(s, "kb") && !strings.HasSuffix(s, "mb") && !strings.HasSuffix(s, "gb") && !strings.HasSuffix(s, "tb") && !strings.HasSuffix(s, "pb") && !strings.HasSuffix(s, "eb") {
		return s[:len(s)-1], 512
	}

	var base int64 = 1024
	rest := s
	if strings.HasSuffix(rest, "iB") {
		rest = rest[:len(rest)-2]
	} else if strings.HasSuffix(rest, "B") {
		base = 1000
		rest = rest[:len(rest)-1]
	}

	if len(rest) == 0 {
		return s, 1
	}
	last := rest[len(rest)-1]
	idx := strings.IndexByte(unitPowers, last|0x20)
	if idx == -1 || (base == 1000 && idx == -1) {
		return s, 1
	}

	mult = 1
	for range idx + 1 {
		mult *= base
	}
	return rest[:len(rest)-1], mult
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

	s, mult := matchSuffixMultiplier(s)
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
