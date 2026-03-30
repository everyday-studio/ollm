# 🕹️ Ollm : LLM 기반 프롬프트 인젝션 게임 플랫폼

<div align="center">
  <img width="1470" height="834" alt="lobby" src="https://github.com/user-attachments/assets/fb8f7f69-8c34-46a1-a081-f99f3f256536" />
</div>

</br>

**Ollm**은 생성형 AI를 활용한 프롬프트 인젝션 게임 플랫폼입니다.  
다양한 게임을 통해 AI의 방어를 뚫고 목표를 달성해보세요!

> 프롬프트 인젝션이란? AI 모델이 '설정된 지시를 무시하도록 유도하는 것'입니다.

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

## 🏗 Architecture & Structure
```txt
apps/
├── backend/internal/
│   ├── domain/     # 비즈니스 핵심 엔티티, 인터페이스, 도메인 에러 구조체 정의
│   ├── repository/ # PostgreSQL 쿼리 실행 및 DB 에러 → 도메인 에러 변환 계층
│   ├── usecase/    # 비즈니스 로직 및 외부 연동 (LLM/GCS) 오케스트레이션 계층
│   └── handler/    # HTTP 통신 (Echo v4), 파라미터 바인딩 및 무결성 검증 계층
└── frontend/src/
    ├── lib/        # API 클라이언트, 인메모리 캐시, Svelte Store 및 공통 컴포넌트
    └── routes/     # SvelteKit 파일 기반 라우팅
        ├── login/
        └── lobby/
            ├── match/
            ├── leaderboard/
            └── mypage/
```

---

## 🚀 Getting Started

### 1. 환경 변수 설정

`apps/backend` 디렉토리에 `.env` 파일을 생성합니다.
```env
# App
APP_ENV="dev"
APP_ADMIN_PATH="/api/admin"

# Database
DB_USER=root
DB_PASSWORD=root1!
GOOSE_DRIVER="postgres"
GOOSE_MIGRATION_DIR="your/path/to/migrations"
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

### 2. Backend & DB 실행
```bash
cd apps/backend
docker compose up -d
```

### 3. Frontend 실행
```bash
cd apps/frontend
docker compose up -d
```
