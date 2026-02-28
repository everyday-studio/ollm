---
trigger: always_on
---

# Go Coding Standards
1. Error Handling: `try-catch` 대신 `if err != nil` 준수. 에러 무시 금지. 도메인 에러 우선 사용. `%w`로 에러 래핑.
2. Variable Naming: CamelCase 사용 (상수는 PascalCase). 인터페이스명은 동사형으로 끝나지 않게 작성.
3. Context: 모든 Repo/UseCase 메서드는 첫 파라미터로 `context.Context` 수신. Handler에서 `c.Request().Context()` 추출하여 전달.
4. DI (Uber FX): `main.go`에서 `fx.Provide()`, `fx.Invoke()` 사용. 생성자는 `New` 접두사 사용.
5. Documentation: Public 함수/타입은 `godoc` 스타일로 주석 필수 작성(영어로).

# Libraries & Database
- Core: Echo v4, PostgreSQL + `database/sql` + `lib/pq` (ORM 금지), Uber FX, Viper + godotenv.
- Testing: `testify`, `testcontainers-go` (PostgreSQL 기반 통합 테스트).
- Database Pool: MaxOpenConns 25, MaxIdleConns 25, ConnMaxLifetime 5m.
- Query: `QueryRowContext()`/`QueryContext()` 사용. `defer rows.Close()` 필수.