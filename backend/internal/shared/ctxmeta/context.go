package ctxmeta

import "context"

type contextKey string

const (
	ipKey        contextKey = "request_ip"
	userAgentKey contextKey = "request_user_agent"
	inTxKey      contextKey = "in_transaction"
)

// WithinTransaction marks a context as running inside a database transaction.
// The transaction handle itself stays in the adapter; this is only the fact,
// so layers above can refuse work a single transaction connection cannot do.
func WithinTransaction(ctx context.Context) context.Context {
	return context.WithValue(ctx, inTxKey, true)
}

// InTransaction reports whether ctx is running inside a database transaction.
func InTransaction(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	within, _ := ctx.Value(inTxKey).(bool)
	return within
}

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
