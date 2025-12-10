package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/everyday-studio/ollm/internal/config"
)

func Setup(cfg *config.Config, e *echo.Echo) {
	// ✅ Trailing Slash 제거 및 301 리디렉트 설정
	e.Pre(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently, // 301 리디렉트
	}))

	// ✅ RequestID: 각 요청에 고유한 ID 부여 (추적 및 디버깅 목적)
	e.Use(middleware.RequestID())

	// ✅ Logger: 요청 및 응답 로깅 설정
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${time_rfc3339}] ${method} ${uri} ${status} request_id=${id}\n",
	}))

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
}
