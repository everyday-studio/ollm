# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Vite는 OS 환경변수가 .env 파일보다 우선순위가 높음
# docker-compose가 args로 전달 → ENV로 설정 → Vite가 .env.production 대신 이 값 사용
# 로컬: apps/frontend/.env의 VITE_API_URL=http://localhost:8080 이 docker-compose에 전달됨
# 프로덕션 서버: 서버의 .env에 VITE_API_URL=https://ollm.everydaystudio.xyz 설정
ARG VITE_API_URL
ARG VITE_GOOGLE_CLIENT_ID
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_GOOGLE_CLIENT_ID=$VITE_GOOGLE_CLIENT_ID

# 의존성 설치 (package-lock.json 캐싱 활용)
COPY package.json package-lock.json ./
RUN npm ci --legacy-peer-deps

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

LABEL org.opencontainers.image.source="https://github.com/everyday-studio/ollm"
