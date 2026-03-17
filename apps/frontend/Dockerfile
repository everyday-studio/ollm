# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# 빌드 타임 환경 변수 (Docker 빌드 시 --build-arg로 전달)
ARG VITE_GOOGLE_CLIENT_ID
ARG VITE_API_URL

# 환경 변수를 빌드 환경에 설정
ENV VITE_GOOGLE_CLIENT_ID=$VITE_GOOGLE_CLIENT_ID
ENV VITE_API_URL=$VITE_API_URL

# 의존성 설치 (package-lock.json 캐싱 활용)
COPY package.json package-lock.json ./
RUN npm ci --legacy-peer-deps

# 소스코드 복사 및 프로덕션 빌드 (이 과정에서 ENV 변수 사용됨)
COPY . .
RUN npm run build

# Production stage
FROM node:20-alpine AS runner

WORKDIR /app

# 빌드 결과물과 프로덕션 의존성만 복사하여 이미지 크기 최적화
COPY --from=builder /app/build ./build
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json

# SvelteKit Node 어댑터 환경변수 설정
ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=5173

EXPOSE 5173

# 빌드된 최적화 서버 실행
CMD ["node", "build/index.js"]
