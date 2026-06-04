package dto

import "time"

type CreateOperationalUserRequest struct {
	Email        string  `json:"email" validate:"required,email,max=255"`
	Role         string  `json:"role" validate:"required,oneof=staff baker manager"`
	FullName     string  `json:"full_name" validate:"required,max=255"`
	Phone        *string `json:"phone,omitempty" validate:"omitempty,max=50"`
	EmployeeCode *string `json:"employee_code,omitempty" validate:"omitempty,max=64"`
	HireDate     *string `json:"hire_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

type CreateOperationalUserResponse struct {
	User         AdminUserResponse `json:"user"`
	TempPassword string            `json:"temp_password"`
}

type UpdateOperationalProfileRequest struct {
	FullName     string  `json:"full_name" validate:"required,max=255"`
	Phone        *string `json:"phone,omitempty" validate:"omitempty,max=50"`
	EmployeeCode *string `json:"employee_code,omitempty" validate:"omitempty,max=64"`
	HireDate     *string `json:"hire_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

type UpdateUserRoleRequest struct {
	Role         string  `json:"role" validate:"required,oneof=staff baker manager"`
	FullName     string  `json:"full_name,omitempty" validate:"omitempty,max=255"`
	Phone        *string `json:"phone,omitempty" validate:"omitempty,max=50"`
	EmployeeCode *string `json:"employee_code,omitempty" validate:"omitempty,max=64"`
	HireDate     *string `json:"hire_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

type AdminUserResponse struct {
	ID                 string     `json:"id"`
	Email              string     `json:"email"`
	Role               string     `json:"role"`
	EmailVerified      bool       `json:"email_verified"`
	MustChangePassword bool       `json:"must_change_password"`
	Disabled           bool       `json:"disabled"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DisplayName        *string    `json:"display_name,omitempty"`
	FullName           *string    `json:"full_name,omitempty"`
	Phone              *string    `json:"phone,omitempty"`
	EmployeeCode       *string    `json:"employee_code,omitempty"`
	HireDate           *time.Time `json:"hire_date,omitempty"`
}
