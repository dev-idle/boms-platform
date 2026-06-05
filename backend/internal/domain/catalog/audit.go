package catalog

import domainuser "github.com/boms/backend/internal/domain/user"

// Manager catalog audit actions (actor role is always manager).
const (
	AuditActionManagerCreatedCategory domainuser.AuditAction = "manager.created_category"
	AuditActionManagerUpdatedCategory domainuser.AuditAction = "manager.updated_category"
	AuditActionManagerDeletedCategory domainuser.AuditAction = "manager.deleted_category"
	AuditActionManagerCreatedProduct  domainuser.AuditAction = "manager.created_product"
	AuditActionManagerUpdatedProduct  domainuser.AuditAction = "manager.updated_product"
	AuditActionManagerDeletedProduct  domainuser.AuditAction = "manager.deleted_product"
)
