package user

// AuditAction is a stable action code for audit trail records.
type AuditAction string

const (
	AuditActionMeUpdatedProfile         AuditAction = "me.updated_profile"
	AuditActionMeChangedPassword        AuditAction = "me.changed_password"
	AuditActionMeSoftDeleted            AuditAction = "me.soft_deleted"
	AuditActionAdminCreatedUser         AuditAction = "admin.created_user"
	AuditActionAdminUpdatedProfile      AuditAction = "admin.updated_profile"
	AuditActionAdminUpdatedRole         AuditAction = "admin.updated_role"
	AuditActionAdminDisabledUser        AuditAction = "admin.disabled_user"
	AuditActionAdminEnabledUser         AuditAction = "admin.enabled_user"
	AuditActionAdminResetUserPassword   AuditAction = "admin.reset_user_password"
	AuditActionAdminRevokedUserSessions AuditAction = "admin.revoked_user_sessions"
)
