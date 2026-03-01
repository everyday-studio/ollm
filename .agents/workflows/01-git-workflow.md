---
description: 커밋 메시지 규칙
---

# Git Commit Message Guidelines
- Format: `{type}/{kebab-case-description}`
- Types: feat, fix, refactor, chore, test, design
- Subject: 간략하게 작성.
- Output Format: 터미널 자동 실행 금지. 다음과 같이 bash 코드 블록으로 제안만 할 것.
  ```bash
  git add .
  git commit -m "type: subject"
  ```