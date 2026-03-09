---
trigger: always_on
---

# Frontend Guidelines (Svelte 5 + Tailwind v4)

## Architecture & Structure
- **Framework**: SvelteKit (Svelte 5 Runes)
- **Language**: TypeScript (Strict Mode)
- **Styling**: Tailwind CSS v4
- **Routing**: `src/routes/` (File-based routing) 
  - `+page.svelte`: UI 비즈니스 로직 최소화
  - `+page.ts/server.ts`: 데이터 로딩 (`server.ts`는 민감 정보/DB 전용)
  - `+layout.svelte`: 공통 레이아웃
  - `+layout.server.ts`: 쿠키 기반 세션 로드 및 리다이렉트

## Svelte 5 Runes - Strict Rules
1. **State**: `$state(0)` 사용. 값 할당으로 상태 변경 (e.g., `count = count + 1`)
2. **Derived**: `$derived(expression)` 사용. (`$:` 사용 금지)
3. **Effects**: 꼭 필요한 경우에만 `$effect(() => {})` 사용.
4. **Props**: `let { data } = $props();` 로딩. (`export let` 금지)
5. **Event Handling**: `onclick={handler}` (표준 HTML 방식).
6. **Snippets**: 반복되는 UI 구조는 외부 컴포넌트 분리 전에 `<snippet>`을 먼저 고려.

## Auth & Session
- **Tokens**: Access Token은 인메모리 스토어(`authStore`) 보관, Refresh Token은 httpOnly Secure 픽스된 쿠키 사용.
- **Refresh Flow**: `onMount` 시점이나 Axios 401/403 응답 시에 `authApi.refresh()`로 재시도 진행.
- **Logout Flow**: 쿠키 무효화 후 `authStore.logout()` 진행, `invalidateAll()` 호출 및 `/login`으로 리다이렉트.

## UI/UX & Styling
- **Modal Pattern**: 모달의 백드롭은 `backdrop-blur-sm z-50 fixed inset-0` 컨벤션 준수. 닫기 액션에 Click propagation 방지 필수 (`on:click|stopPropagation`).
- **Class Ordering**: `Layout` → `Box Model` → `Typography` → `Visual` → `Misc` 순서로 작성.
- **Dynamic Classes**: 리터럴 템플릿 스트링 혹은 `clsx`, `tailwind-merge` 등을 활용. (`@apply` 사용 금지)
- **Error Handling**: Form 에러는 필드별 하단 메시지로 표기하며, Toast (`svelte-french-toast`) 사용 적극 권장.

## Error Handling & Security
- **Axios**: 응답(401, 403, 4xx, 5xx)별 로직 분기. (catch block 내부)
- **XSS Prevention**: `{@html}` 사용을 피하고 필수 시 `DOMPurify` 사용. 일반 텍스트는 표준 바인딩 권장.
- **Sensitive Data**: API 키는 서버 영역(`+page.server.ts` 혹은 Backend Server)에서 관리.
