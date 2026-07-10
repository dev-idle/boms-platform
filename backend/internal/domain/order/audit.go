package order

import domainuser "github.com/boms/backend/internal/domain/user"

const (
	AuditActionStaffUpdatedOrderStatus domainuser.AuditAction = "staff.updated_order_status"
	AuditActionBakerUpdatedOrderStatus domainuser.AuditAction = "baker.updated_order_status"
)
