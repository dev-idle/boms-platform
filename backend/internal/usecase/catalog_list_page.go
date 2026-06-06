package usecase

import (
	"strconv"

	"github.com/boms/backend/internal/shared/utils"
)

const (
	CatalogListDefaultPageSize int32 = 20
	CatalogListMaxPageSize     int32 = 100
	catalogSearchMaxLen              = 100
)

// CatalogListDefaultPageSizeQuery is the Fiber query default for page_size.
var CatalogListDefaultPageSizeQuery = strconv.FormatInt(int64(CatalogListDefaultPageSize), 10)

func normalizeCatalogListPage(page, pageSize int32) (int32, int32) {
	return utils.NormalizePageParams(page, pageSize, CatalogListDefaultPageSize, CatalogListMaxPageSize)
}
