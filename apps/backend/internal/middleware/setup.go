package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/everyday-studio/ollm/internal/config"
	"github.com/everyday-studio/ollm/internal/kit/contexts"
	"github.com/everyday-studio/ollm/internal/kit/security"
)

func Setup(cfg *config.Config, logger *slog.Logger, e *echo.Echo) error {
	// ✅ Trailing Slash 제거 및 301 리디렉트 설정
	e.Pre(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently, // 301 리디렉트
	}))

	// ✅ RequestID: 각 요청에 고유한 ID 부여 (추적 및 디버깅 목적)
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		RequestIDHandler: func(c echo.Context, requestID string) {
			req := c.Request()
			req.Header.Set(echo.HeaderXRequestID, requestID)

			ctx := contexts.WithRequestID(req.Context(), requestID)
			c.SetRequest(req.WithContext(ctx))
		},
	}))

	// ✅ Logger: 요청 및 응답 로깅 설정
	e.Use(LoggerMiddleware(logger))

	// ✅ Recover: 패닉 발생 시 복구 및 로그 출력
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10, // 스택 크기: 1KB
		LogLevel:  log.ERROR,
	}))

	// ✅ Gzip: 응답 압축 (성능 최적화) -> nginx + body limit + rate limit
	/*
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
			Level:     5,   // 압축 레벨 (1-9)
			MinLength: 256, // 최소 압축 크기 (256바이트 이상만 압축)
			Skipper: func(c echo.Context) bool {
				return strings.Contains(c.Path(), "metrics") // 특정 경로는 압축 제외
			},
		}))
	*/

	// ✅ 핸들러 실행 시간 제한
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second, // 핸들러 내부 실행 시간
	}))

	// ✅ 서버 자체 타임아웃 설정
	e.Server.ReadTimeout = 10 * time.Second  // 요청 읽기 타임아웃
	e.Server.WriteTimeout = 40 * time.Second // 응답 쓰기 타임아웃 (Handler Timeout보다 길게)
	e.Server.IdleTimeout = 120 * time.Second // 유휴 연결 타임아웃

	// ✅ CSRF: Cross-Site Request Forgery 방어
	if cfg.App.Env != "dev" {
		// CSRF token route handler
		e.GET("/csrf-token", func(c echo.Context) error {
			token := c.Get(middleware.DefaultCSRFConfig.ContextKey).(string)
			return c.JSON(http.StatusOK, map[string]string{
				"csrf_token": token,
			})
		})

		e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
			TokenLookup:    "header:" + echo.HeaderXCSRFToken,
			CookieSecure:   false,                // HTTPS에서만 쿠키 전송
			CookiePath:     "/",                  // 이 설정 추가
			CookieName:     "_csrf",              // 이 설정 추가
			CookieHTTPOnly: true,                 // JavaScript 접근 금지
			CookieSameSite: http.SameSiteLaxMode, // 동일 출처 외 요청 차단
		}))
	}

	// ✅ JWT Authentication
	// Parse the RSA public key at startup; fail fast if the key is invalid.
	publicKey, err := security.ParseRSAPublicKeyFromBase64(cfg.Secure.JWT.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	e.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey:    publicKey,
		SigningMethod: "RS256",
		TokenLookup:   "header:Authorization:Bearer",

		// Reject refresh tokens used as access tokens.
		// Setting "user" to nil makes AllowRoles treat the request as unauthenticated.
		SuccessHandler: func(c echo.Context) {
			token := c.Get("user").(*jwt.Token)
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				c.Set("user", nil)
				return
			}
			tokenType, _ := claims["type"].(string)
			if tokenType != string(security.AccessToken) {
				c.Set("user", nil)
			}
		},

		ErrorHandler: func(c echo.Context, err error) error {
			// Allow requests without token to pass through (public endpoints).
			// AllowRoles middleware handles authorization downstream.
			if errors.Is(err, echojwt.ErrJWTMissing) {
				return nil
			}

			// For invalid tokens, return echo.HTTPError so the middleware
			// stops the request. Using c.JSON() would return nil and
			// ContinueOnIgnoredError would let the request through.
			var statusCode int
			var errorMsg string

			switch {
			case errors.Is(err, jwt.ErrTokenExpired):
				statusCode = http.StatusUnauthorized
				errorMsg = "Token has expired"
				if cfg.App.Debug {
					logger.Info("JWT token expired",
						"path", c.Path(),
						"method", c.Request().Method)
				}
			case errors.Is(err, jwt.ErrTokenSignatureInvalid):
				statusCode = http.StatusUnauthorized
				errorMsg = "Invalid token signature"
				logger.Warn("JWT token has invalid signature",
					"path", c.Path(),
					"method", c.Request().Method)
			case errors.Is(err, jwt.ErrTokenNotValidYet):
				statusCode = http.StatusUnauthorized
				errorMsg = "Token not valid yet"
				if cfg.App.Debug {
					logger.Info("JWT token not valid yet",
						"path", c.Path(),
						"method", c.Request().Method)
				}
			default:
				statusCode = http.StatusUnauthorized
				errorMsg = "Invalid or malformed token"
				if cfg.App.Debug {
					logger.Warn("JWT validation failed",
						"error", err.Error(),
						"path", c.Path(),
						"method", c.Request().Method)
				}
			}

			return echo.NewHTTPError(statusCode, map[string]string{
				"error": errorMsg,
			})
		},
		ContinueOnIgnoredError: true,
	}))

	return nil
}
