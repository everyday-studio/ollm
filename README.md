# 🕹️ Ollm : LLM 기반 프롬프트 인젝션 게임 플랫폼
<div align="center">
  <img width="735" alt="Screenshot 2026-03-12 at 9 48 01 PM" src="https://github.com/user-attachments/assets/5f768ee2-0284-4631-8039-faddaf1cccfb" />
</div>
</br>

- **Ollm**은 생성형 AI를 활용한 프롬프트 인젝션 게임 플랫폼입니다. 다양한 게임을 통해 AI의 방어를 뚫고 목표를 달성해보세요!
> 프롬프트 인젝션이란? AI 모델이 '설정된 지시를 무시하도록 유도하는 것' 입니다.

---

## 🛠 Tech Stack

**Frontend**  
![Svelte](https://img.shields.io/badge/Svelte_5-FF3E00?style=for-the-badge&logo=svelte&logoColor=white) ![SvelteKit](https://img.shields.io/badge/SvelteKit-FF3E00?style=for-the-badge&logo=svelte&logoColor=white) ![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white) ![TailwindCSS](https://img.shields.io/badge/Tailwind_v4-06B6D4?style=for-the-badge&logo=tailwindcss&logoColor=white)

**Backend**  
![Go](https://img.shields.io/badge/Go_1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white) ![Echo](https://img.shields.io/badge/Echo_v4-000000?style=for-the-badge&logo=go&logoColor=white) ![Uber_FX](https://img.shields.io/badge/Uber_FX-000000?style=for-the-badge)

**Database & Infrastructure**  
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?style=for-the-badge&logo=postgresql&logoColor=white) ![GCP_GCS](https://img.shields.io/badge/Google_Cloud_Storage-4285F4?style=for-the-badge&logo=google-cloud&logoColor=white) ![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)

**AI & LLM**  
![OpenAI](https://img.shields.io/badge/OpenAI-412991?style=for-the-badge&logo=openai&logoColor=white) ![Groq](https://img.shields.io/badge/Groq_Llama_3-F55036?style=for-the-badge)

---

## ✨ Key Technical Highlights

안정적인 트래픽 처리와 유지보수성 향상을 위해 다음과 같은 기술적 고민을 코드에 담았습니다.

- **Clean Architecture & DI**
  Handler - UseCase - Repository로 이어지는 계층을 명확히 분리하고, **Uber FX**를 통해 의존성 주입(DI) 생명주기를 중앙 집중화했습니다. 이를 통해 LLM 클라이언트의 다형성(Polymorphism)을 확장성 있게 확보하고 비즈니스 로직의 결합도를 최소화했습니다.

- **Concurrency & Resource Control**
  무분별한 리소스 낭비를 막기 위해 활성화 처리된(Active) 유령 매치(게임) 생성 개수를 서버 단에서 엄격하게 제한했습니다. 또한, 다건의 매치 데이터 조회 시 발생할 수 있는 성능 저하를 방어하기 위해 쿼리 실행 계획 기반의 **복합 인덱스(Composite Index)** 튜닝을 설계했습니다.

- **Stateless Asset Management (GCS)**
  에셋 관리를 위해 클라이언트 다이렉트 업로드 대신 서버를 경유하는 GCP GCS 이미지 통신 로직을 구축했습니다.

- **Modern Frontend Architecture & Security**
  **Svelte 5 Runes**를 도입해 선언적이고 직관적인 상태 관리를 구현했습니다. **Tailwind CSS v4**로 유연하고 일관된 UI를 구성하고, **Axios Interceptor**를 활용한 Access/Refresh Token 기반 Soft Auth 로직으로 매끄럽고 안전한 사용자 세션 연장을 보장합니다. 더불어 Promise Deduplication을 적용해 여러 컴포넌트 마운트 시에도 토큰 갱신 요청(`ensureSession`)이 중복 발생하지 않도록 네트워크를 최적화했습니다.

- **Frontend Performance & Caching**
  프론트엔드 단에 자체적인 TTL 기반 인메모리 캐시 레이어(`lib/cache`)를 구축해 무분별한 데이터 페칭을 제한했습니다. 매치 생성이나 완료 등 실제 상태 변경이 발생하는 특정 시점에만 데이터 캐시 무효화(Invalidate)를 수행해, 서버 부하와 사용자 체감 속도의 균형을 맞췄습니다.

- **Admin Dashboard & RBAC Authorization**
  유저 및 게임 데이터의 원활한 라이브 서비스를 위해 자체적인 Admin 백오피스를 구축했습니다. 백엔드에서 Role-Based 미들웨어를 통해 권한별 접근을 엄격하게 제어하며, 프론트엔드에서도 관리자 전용 라우팅과 API를 완전히 분리하여 안전한 운영 환경을 확보했습니다.

---

## 🏗 Architecture & Structure

초기 설계부터 외부 프레임워크와의 결합도를 낮추고 비즈니스 로직을 보호하기 위해 백엔드에는 **Clean Architecture** 원칙을 도입했고, 프론트엔드에는 SvelteKit 기반의 **File-based Routing**을 결합했습니다. (`apps/`)

```txt
apps/
├── backend/internal/
│   ├── domain/     # 비즈니스 핵심 엔티티, 인터페이스, 도메인 에러 구조체 정의 (외부 의존성 제로)
│   ├── repository/ # PostgreSQL (database/sql) 쿼리 실행 및 DB 에러 -> 도메인 에러 변환 계층
│   ├── usecase/    # 실제 비즈니스, 권한 판별 로직 및 외부 연동 (LLM/GCS) 오케스트레이션 계층
│   └── handler/    # HTTP 통신 (Echo v4), 파라미터 바인딩 및 무결성 검증 계층
└── frontend/src/
    ├── lib/        # API 클라이언트, 인메모리 캐시, Svelte Store 및 공통 컴포넌트 계층
    └── routes/     # SvelteKit 파일 기반 라우팅 및 UI/상태 관리 계층 (page/layout 분리)
        ├── login/      # 로그인/회원가입 라우팅 (+page.server.ts 기반 서버 검증)
        └── lobby/      # 'refresh_token' 쿠키 기반 보호 라우팅 (+layout.server.ts)
            ├── match/      # 실시간 대화 플레이 및 상태 전환
            ├── leaderboard/# 게임별 랭킹 및 한글 자모 검색
            └── mypage/     # 유저 프로필 관리
```

---

## 🚀 Getting Started

프로젝트 구동을 위한 로컬 환경 가이드입니다.

### 1. 환경 변수 설정
`apps/backend` 디렉토리에 `.env` 파일을 생성하고 아래 필수 환경변수를 설정합니다.

```env
# App
APP_ENV="dev"
APP_ADMIN_PATH="/api/admin"

# Database
DB_USER=root
DB_PASSWORD=root1!
GOOSE_DRIVER="postgres"
GOOSE_MIGRATION_DIR="/Users/hojoong/ollm/apps/backend/internal/db/migrations"
GOOSE_DBSTRING="host=localhost port=5432 user=root password=root1! dbname=mydb sslmode=disable"

# RSA Keys for JWT Auth (Base64 Encoded)
SECURE_JWT_PRIVATE_KEY_BASE64="..."
SECURE_JWT_PUBLIC_KEY_BASE64="..."

# AI & LLM Platforms
LLM_OPENAI_API_KEY="sk-..."
LLM_GROQ_API_KEY="gsk_..."

# Infrastructure
GOOGLE_APPLICATION_CREDENTIALS="./gcp-key.json"
```

### 2. 프로젝트 실행 (Backend & DB)
Docker Compose와 Go 명령어를 이용해 서버를 기동합니다.

```bash
cd apps/backend
docker compose up -d
```

### 3. 프로젝트 실행 (Frontend)
Svelte 최신 환경과 Tailwind v4를 기반으로 프론트엔드 개발 서버를 기동합니다.

```bash
cd apps/frontend
docker compose up -d
```

---

## 🛡️ Admin Dashboard (백오피스)
<div align="center">
  <img width="317" alt="Screenshot 2026-03-12 at 9 42 34 PM" src="https://github.com/user-attachments/assets/60da7fd6-259b-4bd9-8a2a-e0566912656e" />
  <img width="317" alt="Screenshot 2026-03-12 at 9 43 27 PM" src="https://github.com/user-attachments/assets/256d8581-fee7-4e4a-86ff-212719dc4f00" />
  <img width="317" alt="Screenshot 2026-03-12 at 9 44 01 PM" src="https://github.com/user-attachments/assets/798d39e5-738c-4361-9d57-299a2e1d5633" />
</div>

- 라이브 서비스의 게임과 유저를 원활하게 관리하기 위해, 외부 접속을 철저히 통제하는 **자체 Admin 웹페이지**를 구축했습니다.
- **HTMX + a-h/templ** 조합을 통해 별도의 무거운 프론트엔드 프레임워크 없이 백엔드(Echo) 자체 SSR 백오피스를 구현했습니다.

### ✨ 관리자 페이지 주요 기능
- **통합 대시보드 (Dashboard):** 전체 유저 수, 생성된 게임 수 등 핵심 운영 지표 렌더링
- **게임 관리 (Game Management):** 
  - 신규 게임 시나리오 및 시스템 프롬프트(심판 조건 등) 생성/수정 가능 
  - 유저들에게 노출할지 결정하는 **Public/Private Visibility Toggle** 상태 제어
- **보안 및 권한 (Security & RBAC):** 
  - 단순 UI 분리가 아닌 백엔드의 `RoleAdmin` 미들웨어 단에서 API 접근 원천 차단
  - 성공적인 로그인 시 일반 유저와 동일하게 Access/Refresh Token을 발급하여 세션 관리

### 🔐 접속 및 실행 가이드
관리자 페이지는 보안을 위해 **환경변수(`APP_ADMIN_PATH`)로 난독화된 URL 경로**를 사용해야만 접근할 수 있습니다.

1. `.env` 파일에 설정된 `APP_ADMIN_PATH` 값을 확인합니다. (예: `APP_ADMIN_PATH="/7N2kQxP"`)
2. API 서버(포트 `8080`) 기동 후, 해당 경로 아래의 `/login`으로 접속합니다.
   - 📍 **접속 URL:** `http://localhost:8080/7N2kQxP/login`
3. 시드 데이터나 DB를 통해 생성된 **Admin 등급(Role: 3)** 계정 이메일과 비밀번호로 로그인하여 대시보드에 진입합니다.

---

## 📖 Detailed Portfolio

> 💡 **아키텍처 설계 배경, 트러블슈팅, 동시성 제어에 대한 깊은 고민은 [팀 기술블로그]에서 확인하실 수 있습니다.**
