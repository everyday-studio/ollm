package contexts

import (
	"context"
	"errors"
	"log/slog"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

type contextKey string

var loggerContextKey contextKey = "logger_context_key"

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}

func GetLogger(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerContextKey).(*slog.Logger); ok {
		return logger
	}

	logger := slog.Default()
	logger.Error("Default logger used",
		slog.String("reason", "no logger found in context"),
	)
	return logger
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, echo.HeaderXRequestID, requestID)
}

func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(echo.HeaderXRequestID).(string); ok {
		return requestID
	}
	return "no-request-id-in-context"
}

func TokenToUser(c echo.Context) (int64, string, string, error) {
	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return 0, "", "", errors.New("invalid or missing user token")
	}

	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", "", errors.New("invalid token claims")
	}

	idFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, "", "", errors.New("invalid user_id in token")
	}
	id := int64(idFloat)

	email, ok := claims["email"].(string)
	if !ok {
		return 0, "", "", errors.New("invalid email in token")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return 0, "", "", errors.New("invalid roles in token")
	}

	return id, email, role, nil
}
