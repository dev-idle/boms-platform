package product

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOption(t *testing.T) {
	t.Parallel()
	option, err := NewOption(uuid.Nil, OptionSize, "  20 cm ", 1000, true)
	require.NoError(t, err)
	assert.Equal(t, "20 cm", option.Label)

	for name, tc := range map[string]struct {
		group OptionGroup
		label string
		delta int64
	}{
		"unknown_group":   {group: "shape", label: "Heart", delta: 0},
		"empty_label":     {group: OptionFlavor, label: "  ", delta: 0},
		"long_label":      {group: OptionFlavor, label: strings.Repeat("a", 61), delta: 0},
		"formatted_label": {group: OptionFlavor, label: "Mat" + string(rune(0x200b)) + "cha", delta: 0},
		"negative_price":  {group: OptionDecoration, label: "Berries", delta: -1},
		"price_over_cap":  {group: OptionDecoration, label: "Gold leaf", delta: MaxPriceDeltaCents + 1},
	} {
		_, err := NewOption(uuid.Nil, tc.group, tc.label, tc.delta, true)
		assert.ErrorIs(t, err, ErrInvalidOption, name)
	}
}

func TestPlace(t *testing.T) {
	t.Parallel()
	options := []Option{{Group: OptionSize}, {Group: OptionFlavor}, {Group: OptionSize}}
	Place(options)
	assert.Equal(t, []int16{0, 0, 1}, []int16{options[0].SortOrder, options[1].SortOrder, options[2].SortOrder})
}

func TestChoose(t *testing.T) {
	t.Parallel()
	small := Option{ID: uuid.New(), Group: OptionSize, Label: "16 cm", IsActive: true}
	large := Option{ID: uuid.New(), Group: OptionSize, Label: "20 cm", IsActive: true}
	matcha := Option{ID: uuid.New(), Group: OptionFlavor, Label: "Matcha", IsActive: true}
	durian := Option{ID: uuid.New(), Group: OptionFlavor, Label: "Durian", IsActive: false}
	offered := []Option{small, large, matcha, durian}

	chosen, err := Choose(offered, []uuid.UUID{matcha.ID, large.ID})
	require.NoError(t, err)
	assert.Equal(t, []Option{large, matcha}, chosen, "in the order offered")

	for name, ids := range map[string][]uuid.UUID{
		"a_group_left_out":       {large.ID},
		"two_from_one_group":     {small.ID, large.ID, matcha.ID},
		"an_option_not_on_offer": {large.ID, durian.ID},
		"another_products":       {large.ID, uuid.New()},
		"the_same_twice":         {large.ID, large.ID, matcha.ID},
	} {
		_, err := Choose(offered, ids)
		assert.ErrorIs(t, err, ErrInvalidChoice, name)
	}

	none, err := Choose(nil, nil)
	require.NoError(t, err, "a product without options takes only a message and a photo")
	assert.Empty(t, none)
}

func TestNewMessage(t *testing.T) {
	t.Parallel()
	message, err := NewMessage("  Happy birthday Mai  ")
	require.NoError(t, err)
	assert.Equal(t, "Happy birthday Mai", message)
	empty, err := NewMessage("")
	require.NoError(t, err)
	assert.Empty(t, empty)

	_, err = NewMessage(strings.Repeat("a", MaxMessageLength+1))
	require.ErrorIs(t, err, ErrInvalidMessage)
	_, err = NewMessage("Happy\nbirthday")
	require.ErrorIs(t, err, ErrInvalidMessage)
}
