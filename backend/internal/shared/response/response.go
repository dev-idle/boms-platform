package response

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Envelope is the standard API response shape for JSON endpoints.
type Envelope struct {
	Success bool           `json:"success"`
	Data    any            `json:"data,omitempty"`
	Error   *ErrorBody     `json:"error,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
}

// ErrorBody carries a stable machine code and a human-readable message.
type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// OK sends a 200 response with optional data and meta merged with request id.
func OK(c *fiber.Ctx, data any) error {
	return JSON(c, fiber.StatusOK, true, data, nil, MetaFromCtx(c, nil))
}

// MetaFromCtx builds meta including request_id for handlers outside middleware.
func MetaFromCtx(c *fiber.Ctx, base map[string]any) map[string]any {
	return metaWithRequestID(c, base)
}

// JSON writes a typed envelope with the given HTTP status.
func JSON(c *fiber.Ctx, status int, success bool, data any, err *ErrorBody, meta map[string]any) error {
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	return c.Status(status).JSON(Envelope{
		Success: success,
		Data:    data,
		Error:   err,
		Meta:    meta,
	})
}

// Error sends an error envelope. errBody may be nil (generic fallback applied by caller).
func Error(c *fiber.Ctx, status int, err *ErrorBody) error {
	if err == nil {
		err = &ErrorBody{Code: "internal_error", Message: "An unexpected error occurred"}
	}
	return JSON(c, status, false, nil, err, metaWithRequestID(c, nil))
}

func metaWithRequestID(c *fiber.Ctx, base map[string]any) map[string]any {
	rid := RequestIDFromCtx(c)
	if rid == "" {
		return base
	}
	out := map[string]any{"request_id": rid}
	for k, v := range base {
		out[k] = v
	}
	return out
}

// RequestIDFromCtx returns the request correlation id if present.
func RequestIDFromCtx(c *fiber.Ctx) string {
	if v := c.Locals("requestid"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if v := c.Locals("request_id"); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return c.Get(fiber.HeaderXRequestID)
}

// EnsureRequestID sets a request id on the context when middleware did not.
func EnsureRequestID(c *fiber.Ctx) string {
	if id := RequestIDFromCtx(c); id != "" {
		return id
	}
	id := uuid.NewString()
	c.Locals("request_id", id)
	c.Set(fiber.HeaderXRequestID, id)
	return id
}
