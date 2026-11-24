package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// Dollars writes cents as US dollars, the currency the bakery prices in, the
// way the storefront does: $1,234.50.
func Dollars(cents int64) string {
	dollars := strconv.FormatInt(cents/100, 10)
	var grouped strings.Builder
	for i, digit := range dollars {
		if i > 0 && (len(dollars)-i)%3 == 0 {
			grouped.WriteByte(',')
		}
		grouped.WriteRune(digit)
	}
	return fmt.Sprintf("$%s.%02d", grouped.String(), cents%100)
}
