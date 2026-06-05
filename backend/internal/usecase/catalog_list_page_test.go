package usecase

import "testing"

func TestNormalizeCatalogListPage(t *testing.T) {
	t.Parallel()

	page, pageSize := normalizeCatalogListPage(0, 500)
	if page != 1 {
		t.Fatalf("page = %d", page)
	}
	if pageSize != CatalogListMaxPageSize {
		t.Fatalf("pageSize = %d", pageSize)
	}
}
