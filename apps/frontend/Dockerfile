# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# 의존성 설치 (package-lock.json 캐싱 활용)
COPY package.json package-lock.json ./
RUN npm ci --legacy-peer-deps

# 소스코드 복사 (.env.production 포함)
# .env.production은 git에 없으므로 프로덕션 서버에 직접 파일을 두어야 함
# vite build는 기본 mode가 production이므로 .env.production을 자동으로 읽음
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
