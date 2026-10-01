// Package saved is a customer's lists of products to come back to.
package saved

// List names one of a customer's lists: favorites they order again, and a
// wishlist of what they want to try.
type List string

const (
	ListFavorite List = "favorite"
	ListWishlist List = "wishlist"
)

// MaxPerList is the most products one list holds.
const MaxPerList = 100

// Valid reports whether l is a list this package defines.
func (l List) Valid() bool {
	return l == ListFavorite || l == ListWishlist
}
