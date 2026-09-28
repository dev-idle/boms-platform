package postgres

// optionalSearch treats an empty search box the same as no search, so list
// queries skip the text filter instead of matching every row against "".
func optionalSearch(search *string) *string {
	if search == nil || *search == "" {
		return nil
	}
	return search
}
