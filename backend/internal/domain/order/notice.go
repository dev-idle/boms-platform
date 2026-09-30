package order

import (
	"github.com/google/uuid"

	domainevent "github.com/boms/backend/internal/domain/event"
)

// Notice is an email that tells a customer where their order stands. Only the
// moments a customer acts on get one: the order was paid and received, it is
// ready to collect, it was cancelled, or it expired unpaid. Steps inside the bakery (confirmed, in
// production) are shown live on the order page instead of filling an inbox.
type Notice string

const (
	NoticePlaced    Notice = "placed"
	NoticeReady     Notice = "ready"
	NoticeCancelled Notice = "cancelled"
	NoticeExpired   Notice = "expired"
)

// Valid reports whether n is a notice this package defines.
func (n Notice) Valid() bool {
	switch n {
	case NoticePlaced, NoticeReady, NoticeCancelled, NoticeExpired:
		return true
	default:
		return false
	}
}

// NoticeFor names the notice an order event calls for, and the order it is
// about. Any other event calls for none.
func NoticeFor(e domainevent.Event) (uuid.UUID, Notice, bool) {
	var notice Notice
	switch {
	case e.Topic == TopicOrderStatusChanged && Status(e.Data["from"]) == StatusAwaitingPayment &&
		Status(e.Data["status"]) == StatusConfirmed:
		notice = NoticePlaced
	case e.Topic == TopicOrderStatusChanged && Status(e.Data["status"]) == StatusReady:
		notice = NoticeReady
	case e.Topic == TopicOrderStatusChanged && Status(e.Data["status"]) == StatusCancelled:
		notice = NoticeCancelled
	case e.Topic == TopicOrderStatusChanged && Status(e.Data["status"]) == StatusExpired:
		notice = NoticeExpired
	default:
		return uuid.Nil, "", false
	}
	orderID, err := uuid.Parse(e.Data["order_id"])
	if err != nil {
		return uuid.Nil, "", false
	}
	return orderID, notice, true
}

// StillApplies reports whether the notice still describes an order that is now
// in status current. Emails can go out late (a worker was down, the mail server
// asked to retry); one that would contradict what the customer sees on the
// order page is not sent: "ready" for an order already collected or
// cancelled, "received" for one already cancelled.
func (n Notice) StillApplies(current Status) bool {
	switch n {
	case NoticePlaced:
		return current != StatusCancelled
	case NoticeReady:
		return current == StatusReady
	case NoticeCancelled:
		return current == StatusCancelled
	case NoticeExpired:
		return current == StatusExpired
	default:
		return false
	}
}
