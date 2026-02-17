import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

// 이 함수는 사용자가 /lobby 로 들어올 때 '서버'에서 가장 먼저 실행됩니다.
export const load: LayoutServerLoad = async ({ cookies }) => {
  // 1. 브라우저의 쿠키 상자를 열어서 'refresh_token'이 있는지 봅니다.
  const refreshToken = cookies.get('refresh_token');

  // 2. 쿠키가 없다면? (로그인 안 한 사람)
  if (!refreshToken) {
    console.log('⛔ 접근 거부: 로그인이 필요합니다.');
    // 로그인 페이지로 강제 추방 (303 See Other)
    throw redirect(303, '/login');
  }

  // 3. 쿠키가 있다면? 통과!
  console.log('✅ 접근 허용: 인증된 사용자입니다.');
};