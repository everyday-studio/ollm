---
trigger: always_on
---

# Project Context
- Working Directory: 주로 `apps/backend` 내부에서 작업합니다. 프론트엔드 코드는 명시적 요청이 없는 한 수정하지 마십시오.
- Project Type: LLM 텍스트 게임 플랫폼 (Ollm) 백엔드
- Language: Go (Golang) 1.25.4
- Framework: Echo v4
- Architecture: Clean Architecture (Handler -> Usecase -> Repository -> Domain)
- DI Framework: Uber FX (의존성 주입 및 생명주기 관리)

# Project Structure Compliance
1. Domain Layer (`internal/domain/`): 핵심 비즈니스 엔티티, Repo/UseCase 인터페이스, 도메인 에러, DTO 정의. 외부 의존성 없음.
2. Repository Layer (`internal/repository/postgres/`): `database/sql` 직접 사용 (ORM 미사용), DB 에러를 도메인 에러로 변환, Context 타임아웃 지원.
3. UseCase Layer (`internal/usecase/`): 실제 비즈니스 로직, Repo를 통한 데이터 접근, 보안 로직.
4. Handler Layer (`internal/handler/`): HTTP 요청/응답 처리, 파라미터 바인딩 및 검증. 비즈니스 로직 포함 금지.