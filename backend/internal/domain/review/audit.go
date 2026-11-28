package review

import domainuser "github.com/boms/backend/internal/domain/user"

// AuditActionManagerModerated records a manager hiding a review or publishing it again.
const AuditActionManagerModerated domainuser.AuditAction = "manager.moderated_review"
