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
			if ae.StatusCode >= fiber.StatusInternalServerError {
				logError(log, c, err, ae.Code, ae.StatusCode)
			}
			return response.Error(c, ae.StatusCode, &response.ErrorBody{
				Code:    ae.Code,
				Message: ae.Message,
			})
		}

		if ae, ok := errors.AsAppError(err); ok {
			if ae.StatusCode >= fiber.StatusInternalServerError {
				logError(log, c, err, ae.Code, ae.StatusCode)
			}
			return response.Error(c, ae.StatusCode, &response.ErrorBody{
				Code:    ae.Code,
				Message: ae.Message,
				Details: ae.Details,
			})
		}

		logError(log, c, err, errors.ErrInternal.Code, fiber.StatusInternalServerError)
		return response.Error(c, fiber.StatusInternalServerError, &response.ErrorBody{
			Code:    errors.ErrInternal.Code,
			Message: errors.ErrInternal.Message,
		})
	}
}

func logError(log *zap.Logger, c *fiber.Ctx, err error, code string, status int) {
	fields := []zap.Field{
		zap.Error(err),
		zap.String("path", c.Path()),
		zap.String("method", c.Method()),
		zap.String("code", code),
		zap.Int("status", status),
	}
	fields = append(fields, response.ZapCorrelationFields(c)...)
	log.Error("request_failed", fields...)
}
