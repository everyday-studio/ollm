---
trigger: always_on
---

# Security Guidelines
1. Password: 평문 저장 절대 금지. Argon2id (`internal/kit/security/password.go`) 사용.
2. JWT: `golang-jwt/jwt/v4` (RSA 키 기반, Access/Refresh 분리). Refresh Token은 Secure/HttpOnly 쿠키로 전송.
3. Info Leakage: 에러 메시지나 로그에 비밀번호/토큰 등 민감 정보 포함 금지.
4. Environment: 보안 정보는 환경변수로 로드 (`.env`는 gitignore 처리됨).

# Middleware Guidelines
- RBAC: `middleware.AllowRoles()` 사용. (Admin(3) > Manager(2) > User(1) > Public(0))
- Setup (`internal/middleware/setup.go`): RequestID, Logger, Recover, Timeout(30초) 전역 등록.
- JWT Middleware: `echo-jwt/v4` 사용 (soft auth 방식). 토큰 없는 요청 통과 후 AllowRoles에서 제어. 잘못된 토큰은 `echo.NewHTTPError` 반환.