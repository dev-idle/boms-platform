package utils

// StringPtr returns a pointer to s. Useful for optional JSON fields in DTOs.
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to n.
func IntPtr(n int) *int {
	return &n
}
