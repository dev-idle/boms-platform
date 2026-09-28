package order

import (
	"fmt"
	"time"
)

// Code is the reference customers and staff read an order by: CH, the bakery
// day it was placed (YYMMDD) and its number that day, at least three digits.
func Code(day time.Time, number int) string {
	return fmt.Sprintf("CH-%s-%03d", day.Format("060102"), number)
}
