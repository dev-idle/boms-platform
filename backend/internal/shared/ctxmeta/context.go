package ctxmeta

import "context"

type contextKey string

const (
	ipKey        contextKey = "request_ip"
	userAgentKey contextKey = "request_user_agent"
)

func WithRequestMeta(ctx context.Context, ip, userAgent string) context.Context {
	ctx = context.WithValue(ctx, ipKey, ip)
	ctx = context.WithValue(ctx, userAgentKey, userAgent)
	return ctx
}

func IP(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(ipKey).(string)
	return v
}

func UserAgent(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(userAgentKey).(string)
	return v
}
