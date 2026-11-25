package store

import domainuser "github.com/boms/backend/internal/domain/user"

const (
	AuditActionAdminUpdatedSettings   domainuser.AuditAction = "admin.updated_store_settings"
	AuditActionAdminAddedClosedDate   domainuser.AuditAction = "admin.added_closed_date"
	AuditActionAdminRemovedClosedDate domainuser.AuditAction = "admin.removed_closed_date"
	// AuditActionAdminCancelledStrandedOrder is an admin cancelling an order a
	// closed day or new opening hours would strand.
	AuditActionAdminCancelledStrandedOrder domainuser.AuditAction = "admin.cancelled_stranded_order"
)
