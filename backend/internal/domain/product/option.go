package product

import (
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	domainstore "github.com/boms/backend/internal/domain/store"
)

// OptionGroup is what an option sets on a customized product. A customer
// picks one option from each group the product offers.
type OptionGroup string

const (
	OptionSize       OptionGroup = "size"
	OptionFlavor     OptionGroup = "flavor"
	OptionDecoration OptionGroup = "decoration"
)

// Valid reports whether g is a group this package defines.
func (g OptionGroup) Valid() bool {
	switch g {
	case OptionSize, OptionFlavor, OptionDecoration:
		return true
	default:
		return false
	}
}

const (
	// MaxOptions is how many options one product may offer.
	MaxOptions = 30
	// MaxPriceDeltaCents is the most one option adds to the price.
	MaxPriceDeltaCents = 100_000
	// MaxMessageLength is the longest message a customer may ask for on a cake.
	MaxMessageLength = 60

	maxOptionLabelLength = 60
)

// Option is one choice a customer makes on a customizable product, added to
// its price.
type Option struct {
	ID              uuid.UUID
	ProductID       uuid.UUID
	Group           OptionGroup
	Label           string
	PriceDeltaCents int64
	SortOrder       int16
	IsActive        bool
}

// NewOption checks an option a manager offers and returns it with its label
// trimmed. Its id is the one it already has, or zero for a new option.
func NewOption(id uuid.UUID, group OptionGroup, label string, priceDeltaCents int64, isActive bool) (Option, error) {
	label = strings.TrimSpace(label)
	if !group.Valid() || label == "" || utf8.RuneCountInString(label) > maxOptionLabelLength ||
		!domainstore.PlainText(label) || priceDeltaCents < 0 || priceDeltaCents > MaxPriceDeltaCents {
		return Option{}, ErrInvalidOption
	}
	return Option{ID: id, Group: group, Label: label, PriceDeltaCents: priceDeltaCents, IsActive: isActive}, nil
}

// Place numbers each option by its order within its group, as listed.
func Place(options []Option) {
	perGroup := make(map[OptionGroup]int16, 3)
	for i := range options {
		options[i].SortOrder = perGroup[options[i].Group]
		perGroup[options[i].Group]++
	}
}

// Choose resolves a customer's choice among the options a product offers: one
// active option from each group it offers, and nothing else. The chosen
// options come back in the order offered.
func Choose(offered []Option, ids []uuid.UUID) ([]Option, error) {
	picked := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if picked[id] {
			return nil, ErrInvalidChoice
		}
		picked[id] = true
	}
	chosen := make([]Option, 0, len(ids))
	perGroup := make(map[OptionGroup]int, 3)
	for _, option := range offered {
		if !option.IsActive {
			continue
		}
		// Every group offered needs its one pick, even one the customer skipped.
		perGroup[option.Group] += 0
		if picked[option.ID] {
			chosen = append(chosen, option)
			perGroup[option.Group]++
			delete(picked, option.ID)
		}
	}
	if len(picked) > 0 {
		return nil, ErrInvalidChoice
	}
	for _, count := range perGroup {
		if count != 1 {
			return nil, ErrInvalidChoice
		}
	}
	return chosen, nil
}

// NewMessage checks the message a customer asks to have written on a cake and
// returns it trimmed; empty asks for none. It is plain text a baker copies, so
// no control or invisible formatting characters.
func NewMessage(raw string) (string, error) {
	message := strings.TrimSpace(raw)
	if utf8.RuneCountInString(message) > MaxMessageLength || !domainstore.PlainText(message) {
		return "", ErrInvalidMessage
	}
	return message, nil
}
