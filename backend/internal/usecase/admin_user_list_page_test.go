package usecase

import "testing"

func TestAdminUserListDefaultPageSizeQuery_matchesIntConstant(t *testing.T) {
	t.Parallel()
	if AdminUserListDefaultPageSizeQuery != "20" {
		t.Fatalf("query default %q != decimal value of %d", AdminUserListDefaultPageSizeQuery, AdminUserListDefaultPageSize)
	}
}

func TestNormalizeAdminUserListPage(t *testing.T) {
	t.Parallel()
	page, size := normalizeAdminUserListPage(0, 0)
	if page != 1 || size != AdminUserListDefaultPageSize {
		t.Fatalf("got page=%d size=%d, want page=1 size=%d", page, size, AdminUserListDefaultPageSize)
	}
	page, size = normalizeAdminUserListPage(2, 500)
	if page != 2 || size != AdminUserListMaxPageSize {
		t.Fatalf("got page=%d size=%d, want page=2 size=%d", page, size, AdminUserListMaxPageSize)
	}
}
