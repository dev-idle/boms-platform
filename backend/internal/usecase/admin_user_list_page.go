package usecase

import (
	"strconv"

	"github.com/boms/backend/internal/shared/utils"
)

const (
	// AdminUserListDefaultPageSize is the page size when the client omits or sends an invalid page_size.
	AdminUserListDefaultPageSize int32 = 20
	// AdminUserListMaxPageSize caps admin user list queries (limits DB load and response size).
	AdminUserListMaxPageSize int32 = 100
)

// AdminUserListDefaultPageSizeQuery is the Fiber query default for page_size (derived from AdminUserListDefaultPageSize).
var AdminUserListDefaultPageSizeQuery = strconv.FormatInt(int64(AdminUserListDefaultPageSize), 10)

func normalizeAdminUserListPage(page, pageSize int32) (int32, int32) {
	return utils.NormalizePageParams(page, pageSize, AdminUserListDefaultPageSize, AdminUserListMaxPageSize)
}
