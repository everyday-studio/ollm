---
description: DB 마이그레이션 작성 규칙
---

# Database Migration (Goose)
- Tool/Path: Goose (`internal/db/migrations/`)
- Execution: 사용자가 터미널에서 직접 실행함. AI는 쿼리 파일만 작성.
- Rules:
  1. 안전 우선: 명시적 요청 없이 DROP, TRUNCATE 등 파괴적 쿼리 금지.
  2. 하위 호환성: RENAME COLUMN 대신 새 칼럼 추가 후 데이터 복사 제안.
  3. 잠금 방지: 인덱스 생성 시 `-- +goose NO TRANSACTION` 및 `CONCURRENTLY` 사용.
  4. 멱등성: `IF EXISTS`, `IF NOT EXISTS` 적극 활용.
  5. 구조: 반드시 `-- +goose Up` 과 `-- +goose Down` 섹션 모두 작성.