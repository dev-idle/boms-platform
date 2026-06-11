package usecase

import (
	"testing"

	domainuser "github.com/boms/backend/internal/domain/user"
)

func TestFormatUserAuditSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action domainuser.AuditAction
		before []byte
		after  []byte
		want   string
	}{
		{
			name:   "disabled",
			action: domainuser.AuditActionAdminDisabledUser,
			want:   "Account disabled by administrator",
		},
		{
			name:   "role change",
			action: domainuser.AuditActionAdminUpdatedRole,
			before: []byte(`{"role":"staff"}`),
			after:  []byte(`{"role":"manager"}`),
			want:   "Role changed from staff to manager",
		},
		{
			name:   "created with role",
			action: domainuser.AuditActionAdminCreatedUser,
			after:  []byte(`{"role":"baker"}`),
			want:   "Account created with role baker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatUserAuditSummary(tt.action, tt.before, tt.after)
			if got != tt.want {
				t.Fatalf("formatUserAuditSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}
