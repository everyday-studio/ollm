package ctx

import (
	"context"

	"github.com/labstack/echo/v4"
)

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, echo.HeaderXRequestID, requestID)
}

func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(echo.HeaderXRequestID).(string); ok {
		return requestID
	}
	return "no-request-id-in-context"
}
