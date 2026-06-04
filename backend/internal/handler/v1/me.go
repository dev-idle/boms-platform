package v1

import (
	"errors"

	domainprofile "github.com/boms/backend/internal/domain/profile"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type MeHandler struct {
	usecase *usecase.MeUsecase
}

func NewMeHandler(uc *usecase.MeUsecase) *MeHandler {
	return &MeHandler{usecase: uc}
}

func (h *MeHandler) Get(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	user, profile, err := h.usecase.Get(c.UserContext(), userID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, toMeResponse(user, profile))
}

func (h *MeHandler) Patch(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	var req dto.UpdateMeRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	if role, ok := middleware.GetRole(c); ok {
		sanitizeSelfProfileUpdate(role, &req)
	}
	user, profile, err := h.usecase.UpdateProfile(c.UserContext(), userID, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, toMeResponse(user, profile))
}

func (h *MeHandler) PatchPassword(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	var req dto.ChangeMyPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	if err := h.usecase.ChangePassword(c.UserContext(), userID, req.OldPassword, req.NewPassword); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}

func (h *MeHandler) Delete(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	if err := h.usecase.SoftDeleteSelf(c.UserContext(), userID); err != nil {
		if errors.Is(err, apperrors.ErrForbidden) {
			return writeAppError(c, apperrors.New(
				apperrors.ErrForbidden.StatusCode,
				apperrors.ErrForbidden.Code,
				"Only customers can self-delete",
			))
		}
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}

func toMeResponse(user *domainuser.User, profile any) dto.MeResponse {
	res := dto.MeResponse{
		ID:                 user.ID.String(),
		Email:              user.Email,
		Role:               string(user.Role),
		EmailVerified:      user.EmailVerified,
		MustChangePassword: user.MustChangePassword,
		Disabled:           user.Disabled(),
		CreatedAt:          user.CreatedAt,
	}
	switch p := profile.(type) {
	case *domainprofile.Customer:
		res.Profile = dto.MeCustomerProfileResponse{
			Type:        "customer",
			DisplayName: p.DisplayName,
			Phone:       p.Phone,
		}
	case *domainprofile.Staff:
		res.Profile = dto.MeStaffProfileResponse{
			Type:         "staff",
			FullName:     p.FullName,
			Phone:        p.Phone,
			EmployeeCode: p.EmployeeCode,
			HireDate:     p.HireDate.Format("2006-01-02"),
		}
	case *domainprofile.Admin:
		res.Profile = dto.MeAdminProfileResponse{
			Type:     "admin",
			FullName: p.FullName,
			Phone:    p.Phone,
		}
	default:
		res.Profile = nil
	}
	return res
}

// sanitizeSelfProfileUpdate strips fields a role may not change via PATCH /me.
func sanitizeSelfProfileUpdate(role domainuser.Role, req *dto.UpdateMeRequest) {
	switch role {
	case domainuser.RoleCustomer:
		req.FullName = nil
		req.EmployeeCode = nil
		req.HireDate = nil
	case domainuser.RoleStaff, domainuser.RoleBaker, domainuser.RoleManager:
		req.DisplayName = nil
		req.EmployeeCode = nil
		req.HireDate = nil
	case domainuser.RoleAdmin:
		req.DisplayName = nil
		req.EmployeeCode = nil
		req.HireDate = nil
	default:
		req.DisplayName = nil
		req.FullName = nil
		req.EmployeeCode = nil
		req.HireDate = nil
	}
}
