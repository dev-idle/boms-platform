package order

import domainuser "github.com/boms/backend/internal/domain/user"

const (
	AuditActionStaffUpdatedOrderStatus  domainuser.AuditAction = "staff.updated_order_status"
	AuditActionStaffUpdatedTicketStatus domainuser.AuditAction = "staff.updated_ticket_status"
	AuditActionBakerUpdatedTicketStatus domainuser.AuditAction = "baker.updated_ticket_status"
	AuditActionStaffMovedTicket         domainuser.AuditAction = "staff.moved_ticket"
	AuditActionStaffMovedPickup         domainuser.AuditAction = "staff.moved_pickup"
	// AuditActionManagerHandedOverWithoutCode is a manager handing over an
	// online order whose customer could not give its pickup code.
	AuditActionManagerHandedOverWithoutCode domainuser.AuditAction = "manager.handed_over_without_code"
)
