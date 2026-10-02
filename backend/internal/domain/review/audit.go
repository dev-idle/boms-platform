package review

import domainuser "github.com/boms/backend/internal/domain/user"

// AuditActionManagerModerated records a manager publishing or hiding a review.
const AuditActionManagerModerated domainuser.AuditAction = "manager.moderated_review"
