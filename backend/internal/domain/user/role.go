package user

// Role identifies a user's application role.
type Role string

const (
	RoleCustomer Role = "customer"
	RoleStaff    Role = "staff"
	RoleBaker    Role = "baker"
	RoleManager  Role = "manager"
	RoleAdmin    Role = "admin"
)

// IsOperational reports whether the role belongs to internal operators.
func (r Role) IsOperational() bool {
	switch r {
	case RoleStaff, RoleBaker, RoleManager, RoleAdmin:
		return true
	default:
		return false
	}
}

// CanBeAssigned reports whether an admin may assign this role to another user (excludes admin).
func (r Role) CanBeAssigned() bool {
	switch r {
	case RoleCustomer, RoleStaff, RoleBaker, RoleManager:
		return true
	default:
		return false
	}
}

// IsAdmin reports platform administrator role.
func (r Role) IsAdmin() bool {
	return r == RoleAdmin
}

// ProfileType maps a role to the backing profile table type.
func (r Role) ProfileType() string {
	switch r {
	case RoleCustomer:
		return "customer"
	case RoleAdmin:
		return "admin"
	case RoleStaff, RoleBaker, RoleManager:
		return "staff"
	default:
		return ""
	}
}
