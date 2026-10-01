package product

import (
	"github.com/google/uuid"

	domainevent "github.com/boms/backend/internal/domain/event"
)

// TopicSoldOutChanged announces a product running out for today, or coming
// back. Every open cart and pickup picker may hold it, and the storefront shows
// it, so every socket hears it.
const TopicSoldOutChanged domainevent.Topic = "product.sold_out_changed"

// SoldOutChangedEvent tells every open page to read the product again.
func SoldOutChangedEvent(productID uuid.UUID) domainevent.Event {
	return domainevent.New(TopicSoldOutChanged, domainevent.Audience{Public: true}, map[string]string{
		"product_id": productID.String(),
	})
}
