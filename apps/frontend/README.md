# 🎮 Ollm Frontend

**Svelte 5 + SvelteKit** 기반의 LLM 프롬프트 인젝션 게임 프론트엔드입니다.

## Tech Stack

| 기술 | 버전 | 용도 |
|------|------|------|
| Svelte | 5 (Runes) | 선언적 반응형 UI |
| SvelteKit | 2 | 파일 기반 라우팅, SSR, 서버 가드 |
| TypeScript | 5+ | 타입 안전성 |
| Tailwind CSS | v4 | 유틸리티 퍼스트 스타일링 |
| Axios | — | HTTP 클라이언트 + 인터셉터 |

## 프로젝트 구조

```
src/
├── lib/
│   ├── api/            # Axios 클라이언트 (withCredentials, 401 인터셉터)
│   ├── cache/          # TTL 기반 인메모리 캐시 (게임/매치 데이터)
│   ├── features/       # 도메인별 API·Store·타입 모듈
│   │   ├── auth/       # 로그인, 토큰 갱신, 세션 관리
│   │   ├── game/       # 게임·매치·메시지 API
│   │   ├── upload/     # GCS 이미지 업로드
│   │   └── user/       # 유저 프로필 API
│   ├── utils/          # 공통 유틸리티 (게임 헬퍼, 이미지 폴백 등)
│   └── components/ui/  # 공통 UI 컴포넌트 (GameCard 등)
├── routes/
│   ├── +page.server.ts     # 루트 → /lobby 또는 /login 리다이렉트
│   ├── login/              # 로그인·회원가입
│   └── lobby/              # 인증 보호 라우팅 (layout.server.ts)
│       ├── match/[id]/     # 실시간 AI 대화 플레이
│       ├── leaderboard/    # 게임별 랭킹
│       ├── mypage/         # 프로필·닉네임·아바타 관리
│       ├── guide/          # 이용 가이드
│       └── achievements/   # 업적 (준비 중)
└── app.html
```

## 주요 기술 포인트

### 인증 & 세션
- **httpOnly 쿠키** 기반 Refresh Token + 메모리 Access Token 분리
- **Promise Deduplication** (`ensureSession`): 여러 컴포넌트가 동시에 마운트되어도 토큰 갱신은 단 1회
- **Axios 인터셉터**: 401 응답 시 자동 토큰 갱신 후 원본 요청 재시도
- **서버 가드**: `+layout.server.ts`에서 쿠키 검증, 미인증 시 `/login` 리다이렉트

### 프론트엔드 캐싱
- `lib/cache/` 에 TTL 기반 인메모리 캐시 레이어 구축
- 게임 목록·매치 목록 등 반복 페칭 방지
- 매치 생성·완료 등 상태 변경 시점에서만 캐시 무효화 (`invalidateMatchesCache`)

### 매치 페이지 안정성
- **Race Condition 방어**: `latestLoadToken`으로 비동기 응답이 현재 매치에 해당하는지 검증
- **Per-Match 전송 상태**: `sendingMatchId`로 매치 전환 시 이전 전송 상태가 오염되지 않도록 격리

### 반응형 상태 관리
- Svelte 5 Runes (`$state`, `$derived`, `$effect`) 기반 선언적 반응성
- `SvelteMap`을 활용한 매치 그룹핑 및 파생 상태 계산

## 개발 서버 실행

```bash
npm install
npm run dev
```

## 빌드

```bash
npm run build
npm run preview   # 프로덕션 빌드 미리보기
```

## 환경 변수

```env
VITE_API_URL=http://localhost:8080   # 백엔드 API 주소
```
