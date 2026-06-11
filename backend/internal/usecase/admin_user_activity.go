package usecase

import (
	"context"
	"encoding/json"
	"strconv"

	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/google/uuid"
)

const (
	AdminUserActivityDefaultPageSize int32 = 20
	AdminUserActivityMaxPageSize     int32 = 100
)

var AdminUserActivityDefaultPageSizeQuery = strconv.FormatInt(int64(AdminUserActivityDefaultPageSize), 10)

func normalizeAdminUserActivityPage(page, pageSize int32) (int32, int32) {
	return utils.NormalizePageParams(page, pageSize, AdminUserActivityDefaultPageSize, AdminUserActivityMaxPageSize)
}

func (u *AdminUserUsecase) ListActivity(
	ctx context.Context,
	userID uuid.UUID,
	rawPage, rawPageSize int32,
) (items []dto.AdminUserActivityLogResponse, total int64, page, pageSize int32, err error) {
	if _, err := u.users.AdminGetByID(ctx, userID); err != nil {
		return nil, 0, 0, 0, err
	}

	page, pageSize = normalizeAdminUserActivityPage(rawPage, rawPageSize)
	total, err = u.auditLogs.CountByTargetID(ctx, userID)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	rows, err := u.auditLogs.ListByTargetID(ctx, port.ListAuditLogsByTargetParams{
		TargetID: userID,
		Limit:    pageSize,
		Offset:   utils.PageOffset(page, pageSize),
	})
	if err != nil {
		return nil, 0, 0, 0, err
	}

	out := make([]dto.AdminUserActivityLogResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.AdminUserActivityLogResponse{
			ID:         row.ID.String(),
			Action:     string(row.Action),
			Summary:    formatUserAuditSummary(row.Action, row.BeforeJSON, row.AfterJSON),
			ActorID:    row.ActorID.String(),
			ActorEmail: row.ActorEmail,
			ActorRole:  string(row.ActorRole),
			CreatedAt:  row.CreatedAt,
		})
	}
	return out, total, page, pageSize, nil
}

func formatUserAuditSummary(
	action domainuser.AuditAction,
	beforeJSON, afterJSON []byte,
) string {
	switch action {
	case domainuser.AuditActionAdminDisabledUser:
		return "Account disabled by administrator"
	case domainuser.AuditActionAdminEnabledUser:
		return "Account re-enabled by administrator"
	case domainuser.AuditActionAdminRevokedUserSessions:
		return "All active sessions revoked"
	case domainuser.AuditActionAdminResetUserPassword:
		return "Temporary password issued"
	case domainuser.AuditActionAdminCreatedUser:
		if role := jsonStringField(afterJSON, "role"); role != "" {
			return "Account created with role " + role
		}
		return "Account created"
	case domainuser.AuditActionAdminUpdatedRole:
		beforeRole := jsonStringField(beforeJSON, "role")
		afterRole := jsonStringField(afterJSON, "role")
		if beforeRole != "" && afterRole != "" && beforeRole != afterRole {
			return "Role changed from " + beforeRole + " to " + afterRole
		}
		return "Role updated"
	case domainuser.AuditActionAdminUpdatedProfile:
		return "Profile updated by administrator"
	case domainuser.AuditActionMeUpdatedProfile:
		return "Profile updated by user"
	case domainuser.AuditActionMeChangedPassword:
		return "Password changed by user"
	case domainuser.AuditActionMeSoftDeleted:
		return "Account deleted by user"
	default:
		return string(action)
	}
}

func jsonStringField(raw []byte, key string) string {
	if len(raw) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	value, ok := payload[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}
