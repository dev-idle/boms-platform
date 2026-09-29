package order

import domainuser "github.com/boms/backend/internal/domain/user"

const (
	AuditActionStaffUpdatedOrderStatus  domainuser.AuditAction = "staff.updated_order_status"
	AuditActionStaffUpdatedTicketStatus domainuser.AuditAction = "staff.updated_ticket_status"
	AuditActionBakerUpdatedTicketStatus domainuser.AuditAction = "baker.updated_ticket_status"
	AuditActionStaffMovedTicket         domainuser.AuditAction = "staff.moved_ticket"
)
