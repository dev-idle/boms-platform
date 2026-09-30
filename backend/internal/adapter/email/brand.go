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
	AddressLine:  "123 Greige Lane",
	ContactEmail: "hello@chouxbakery.example",
	ContactPhone: "(555) 123-4567",
}
