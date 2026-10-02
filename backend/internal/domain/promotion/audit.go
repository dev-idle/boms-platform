package promotion

import domainuser "github.com/boms/backend/internal/domain/user"

const (
	// AuditActionManagerSent records a manager sending a promotion.
	AuditActionManagerSent domainuser.AuditAction = "manager.sent_promotion"
	// AuditActionMeUnsubscribed records a customer withdrawing their agreement
	// to promotions from a promotion email's link.
	AuditActionMeUnsubscribed domainuser.AuditAction = "me.unsubscribed_from_promotions"
)
