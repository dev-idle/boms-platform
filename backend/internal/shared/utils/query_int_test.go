package utils

import (
	"math"
	"testing"
)

func TestParseQueryInt32(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		def  int32
		want int32
	}{
		{name: "empty", s: "", def: 20, want: 20},
		{name: "trimmed", s: "  3  ", def: 1, want: 3},
		{name: "invalid", s: "not-a-number", def: 5, want: 5},
		{name: "overflow int32", s: "2147483648", def: 1, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ParseQueryInt32(tt.s, tt.def); got != tt.want {
				t.Errorf("ParseQueryInt32(%q, %d) = %d, want %d", tt.s, tt.def, got, tt.want)
			}
		})
	}
}

func TestInt32FromInt64(t *testing.T) {
	t.Parallel()
	if got := Int32FromInt64(int64(math.MaxInt32) + 1); got != math.MaxInt32 {
		t.Fatalf("overflow clamp: got %d", got)
	}
	if got := Int32FromInt64(int64(math.MinInt32) - 1); got != math.MinInt32 {
		t.Fatalf("underflow clamp: got %d", got)
	}
	if got := Int32FromInt64(42); got != 42 {
		t.Fatalf("in-range: got %d", got)
	}
}
