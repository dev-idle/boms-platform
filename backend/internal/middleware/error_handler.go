package middleware

import (
	stderrors "errors"

	"github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ErrorHandler maps errors to the standard JSON envelope. Wire into fiber.Config.ErrorHandler.
func ErrorHandler(log *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		if err == nil {
			return nil
		}

		response.EnsureRequestID(c)
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)

		var fe *fiber.Error
		if stderrors.As(err, &fe) {
			ae := errors.FromFiberError(fe)
			return response.Error(c, ae.StatusCode, &response.ErrorBody{
				Code:    ae.Code,
				Message: ae.Message,
			})
		}

		if ae, ok := errors.AsAppError(err); ok {
			return response.Error(c, ae.StatusCode, &response.ErrorBody{
				Code:    ae.Code,
				Message: ae.Message,
				Details: ae.Details,
			})
		}

		log.Error("unhandled_error",
			zap.Error(err),
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
			zap.String("request_id", response.RequestIDFromCtx(c)),
		)

		return response.Error(c, fiber.StatusInternalServerError, &response.ErrorBody{
			Code:    errors.ErrInternal.Code,
			Message: errors.ErrInternal.Message,
		})
	}
}
