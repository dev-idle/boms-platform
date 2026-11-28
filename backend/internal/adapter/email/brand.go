// Package email writes and sends customer email: templates in the bakery's
// style, and delivery over SMTP.
package email

// brand is the bakery as its emails present it. The storefront shows the same
// details; both sides are tested against contracts/brand.json.
type brand struct {
	Name         string
	AddressLine  string
	ContactEmail string
	ContactPhone string
}

var choux = brand{
	Name:         "Choux",
	AddressLine:  "Ninh Kieu, Can Tho",
	ContactEmail: "nghiantgcc19224@fpt.edu.vn",
	ContactPhone: "0123 456 789",
}
