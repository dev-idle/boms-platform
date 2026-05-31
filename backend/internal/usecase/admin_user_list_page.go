package usecase

import "github.com/boms/backend/internal/shared/utils"

const (
	// AdminUserListDefaultPageSize is the page size when the client omits or sends an invalid page_size.
	AdminUserListDefaultPageSize int32 = 20
	// AdminUserListDefaultPageSizeQuery is the Fiber query default for page_size (must match AdminUserListDefaultPageSize).
	AdminUserListDefaultPageSizeQuery = "20"
	// AdminUserListMaxPageSize caps admin user list queries (limits DB load and response size).
	AdminUserListMaxPageSize int32 = 100
)

// NormalizeAdminUserListPage applies admin user list pagination rules.
func NormalizeAdminUserListPage(page, pageSize int32) (int32, int32) {
	return utils.NormalizePageParams(page, pageSize, AdminUserListDefaultPageSize, AdminUserListMaxPageSize)
}
