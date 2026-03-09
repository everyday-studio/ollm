---
description: 커밋 및 브랜치 워크플로우 규칙
---

# Git Workflow Guidelines

## 1. Branch Naming
- **Format**: `Type/description`
- `Type`은 첫 글자를 대문자로 작성합니다 (e.g., `Feat`, `Fix`, `Hotfix`, `Refactor`, `Chore`, `Docs`).
- `description`은 소문자 `kebab-case`로 작성합니다.
- 예시: `Feat/login-page`, `Fix/typo-error`, `Hotfix/payment-bug`

## 2. Commit Messages
커밋 메시지는 Conventional Commits 규칙을 따르며 소문자로 시작합니다.
- **Format**: `type: subject`
- **Types**: 
  - `feat`: 새로운 기능 추가
  - `fix`: 버그 수정
  - `docs`: 문서 수정
  - `style`: 코드 포맷팅 등 로직 변경 없는 사항
  - `refactor`: 코드 리팩토링
  - `test`: 테스트 코드 추가/수정
  - `chore`: 빌드, 패키지 파일 변경, 등 기타
  - `design`: CSS, UI 변경

- **Subject Rule**:
  - 간결하게 작성
  - 동사 원형 시작의 명령문 형태로 작성 (e.g., `add user`, `use invoke`)
  - 마지막에 마침표(.) 사용 안 함

## 3. Pull Request Policy
- 머지는 반드시 Pull Request(PR)를 통해 수행합니다.
- 머지 전략은 **Squash Merge**만 허용합니다.

## 4. Agent Output Rule
- 터미널 자동 실행은 금지되며, 다음과 같이 `bash` 블록으로 제안만 할 것:
```bash
git add .
git commit -m "type: subject"
```