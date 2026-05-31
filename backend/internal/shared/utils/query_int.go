package utils

import (
	"math"
	"strconv"
	"strings"
)

// ParseQueryInt32 parses s as a base-10 int32. On empty input or parse error, returns defaultVal.
func ParseQueryInt32(s string, defaultVal int32) int32 {
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultVal
	}
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return defaultVal
	}
	return int32(n)
}

// Int32FromInt64 narrows v to int32 after clamping to [math.MinInt32, math.MaxInt32].
func Int32FromInt64(v int64) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}
