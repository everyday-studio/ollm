package middleware

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/everyday-studio/ollm/internal/kit/contexts"
)

func LoggerMiddleware(logger *slog.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{

		BeforeNextFunc: func(c echo.Context) {
			req := c.Request()
			requestID := contexts.GetRequestID(c.Request().Context())
			ctxLogger := logger.With(slog.String("request_id", requestID))
			reqCtx := contexts.WithLogger(req.Context(), ctxLogger)
			c.SetRequest(req.WithContext(reqCtx))
		},

		// 요청 및 응답에서 로깅할 값들
		LogRequestID: true, // 요청 ID 로깅
		LogStatus:    true, // HTTP 응답 상태 코드 로깅
		LogMethod:    true, // HTTP 메서드 (GET, POST 등) 로깅
		LogURIPath:   true, // 요청된 URI 경로 로깅
		LogRemoteIP:  true, // 클라이언트 IP 주소 로깅
		LogUserAgent: true, // 요청한 클라이언트의 User-Agent 로깅
		LogReferer:   true, // Referer 헤더(어디서 요청이 왔는지) 로깅
		LogLatency:   true, // 요청 완료까지 걸린 시간 로깅
		LogError:     true, // 에러 발생 시 로깅
		HandleError:  true, // 에러 발생 시 글로벌 에러 핸들러로 전달하여 적절한 응답을 반환

		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			// 공통 로깅 필드 설정
			baseLogger := logger.With(
				slog.String("request_id", v.RequestID),                           // 요청 고유 ID
				slog.Int("status", v.Status),                                     // HTTP 응답 상태 코드
				slog.String("method", v.Method),                                  // HTTP 요청 메서드
				slog.String("path", v.URIPath),                                   // 실제 요청된 URI 경로
				slog.String("remote_ip", v.RemoteIP),                             // 요청을 보낸 클라이언트의 IP 주소
				slog.String("user_agent", v.UserAgent),                           // 클라이언트의 브라우저, OS 등 식별 정보
				slog.String("referer", v.Referer),                                // 요청이 유입된 이전 페이지 주소
				slog.Float64("latency_ms", float64(v.Latency.Nanoseconds())/1e6), // 요청 처리에 소요된 시간
			)
			if v.Error != nil {
				baseLogger = baseLogger.With(slog.String("err", v.Error.Error()))
			}

			switch {
			case v.Status >= 500:
				baseLogger.Error("SERVER_ERROR")
			case v.Status >= 400:
				baseLogger.Info("CLIENT_ERROR")
			default:
				baseLogger.Info("REQUEST_SUCCESS")
			}
			return nil
		},
	})
}
