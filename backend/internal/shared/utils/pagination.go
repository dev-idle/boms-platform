package utils

// NormalizePageParams clamps page to at least 1 and pageSize to [defaultPageSize, maxPageSize].
func NormalizePageParams(page, pageSize, defaultPageSize, maxPageSize int32) (int32, int32) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

// PageOffset returns the SQL/list offset for zero-based paging (page is 1-based).
func PageOffset(page, pageSize int32) int32 {
	return Int32FromInt64(int64(page-1) * int64(pageSize))
}
