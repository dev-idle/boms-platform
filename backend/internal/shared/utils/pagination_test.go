package utils

import "testing"

func TestNormalizePageParams(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		page, pageSize   int32
		defaultSize, max int32
		wantPage         int32
		wantSize         int32
	}{
		{
			name: "defaults invalid page and size", page: 0, pageSize: 0,
			defaultSize: 20, max: 100, wantPage: 1, wantSize: 20,
		},
		{
			name: "caps oversized page size", page: 2, pageSize: 500,
			defaultSize: 20, max: 100, wantPage: 2, wantSize: 100,
		},
		{
			name: "in range unchanged", page: 3, pageSize: 50,
			defaultSize: 20, max: 100, wantPage: 3, wantSize: 50,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotPage, gotSize := NormalizePageParams(tt.page, tt.pageSize, tt.defaultSize, tt.max)
			if gotPage != tt.wantPage || gotSize != tt.wantSize {
				t.Fatalf("NormalizePageParams() = (%d, %d), want (%d, %d)", gotPage, gotSize, tt.wantPage, tt.wantSize)
			}
		})
	}
}

func TestPageOffset(t *testing.T) {
	t.Parallel()
	if got := PageOffset(1, 20); got != 0 {
		t.Fatalf("first page offset: got %d", got)
	}
	if got := PageOffset(3, 20); got != 40 {
		t.Fatalf("third page offset: got %d", got)
	}
}
