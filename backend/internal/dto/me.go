package dto

import "time"

type UpdateMeRequest struct {
	DisplayName *string `json:"display_name,omitempty" validate:"omitempty,max=255"`
	Phone       *string `json:"phone,omitempty" validate:"omitempty,vn_phone"`
	// `omitnil`, not `omitempty`: leaving the field out keeps the stored name,
	// but sending an empty one is a request to erase an operational identity and
	// is refused rather than written through.
	FullName     *string `json:"full_name,omitempty" validate:"omitnil,min=1,max=255"`
	EmployeeCode *string `json:"employee_code,omitempty" validate:"omitempty,max=64"`
}

type ChangeMyPasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=1"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128,password_complexity"`
}

type MeCustomerProfileResponse struct {
	Type        string  `json:"type"`
	DisplayName *string `json:"display_name,omitempty"`
	Phone       *string `json:"phone,omitempty"`
}

type MeStaffProfileResponse struct {
	Type         string  `json:"type"`
	FullName     string  `json:"full_name"`
	Phone        *string `json:"phone,omitempty"`
	EmployeeCode string  `json:"employee_code"`
}

type MeAdminProfileResponse struct {
	Type     string  `json:"type"`
	FullName string  `json:"full_name"`
	Phone    *string `json:"phone,omitempty"`
}

type MeResponse struct {
	ID                 string    `json:"id"`
	Email              string    `json:"email"`
	Role               string    `json:"role"`
	EmailVerified      bool      `json:"email_verified"`
	MustChangePassword bool      `json:"must_change_password"`
	Disabled           bool      `json:"disabled"`
	CreatedAt          time.Time `json:"created_at"`
	Profile            any       `json:"profile"`
}
