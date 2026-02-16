// src/lib/api.ts
import axios from 'axios';
import { get } from 'svelte/store';
import { authStore } from './stores/auth';

// 환경변수에서 API 주소 가져오기 (없으면 로컬호스트)
const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Axios 인스턴스 생성
const api = axios.create({
  baseURL: BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // 쿠키(Refresh Token)를 주고받기 위해 필수!
});

// [요청 인터셉터] : 모든 요청 헤더에 Access Token 자동 주입
api.interceptors.request.use(
  (config) => {
    const state = get(authStore);
    if (state.accessToken) {
      config.headers.Authorization = `Bearer ${state.accessToken}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// [응답 인터셉터] : 401 에러 발생 시 토큰 갱신 시도
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // 401 에러이고, 아직 재시도를 안 했다면?
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        // 1. 토큰 갱신 요청 (쿠키에 있는 Refresh Token 사용)
        // 주의: 이 경로는 백엔드 명세서의 Refresh 엔드포인트와 일치해야 합니다.
        const { data } = await axios.post(
            `${BASE_URL}/auth/refresh`, 
            {}, 
            { withCredentials: true } // 중요: 쿠키 전송
        );

        // 2. 새 토큰을 스토어에 저장
        authStore.updateToken(data.access_token);

        // 3. 실패했던 요청의 헤더를 새 토큰으로 교체
        originalRequest.headers.Authorization = `Bearer ${data.access_token}`;

        // 4. 원래 요청 재실행
        return api(originalRequest);

      } catch (refreshError) {
        // 갱신마저 실패하면 진짜 로그아웃 처리
        console.error('Session expired, logging out...');
        authStore.logout();
        
        // 로그인 페이지로 리다이렉트 (필요 시)
        // window.location.href = '/login'; 
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

export default api;