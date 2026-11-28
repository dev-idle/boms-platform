package store

import domainuser "github.com/boms/backend/internal/domain/user"

const (
	AuditActionManagerUpdatedSettings   domainuser.AuditAction = "manager.updated_store_settings"
	AuditActionManagerAddedClosedDate   domainuser.AuditAction = "manager.added_closed_date"
	AuditActionManagerRemovedClosedDate domainuser.AuditAction = "manager.removed_closed_date"
	// AuditActionManagerCancelledStrandedOrder is a manager cancelling an order a
	// closed day or new opening hours would strand.
	AuditActionManagerCancelledStrandedOrder domainuser.AuditAction = "manager.cancelled_stranded_order"
)
