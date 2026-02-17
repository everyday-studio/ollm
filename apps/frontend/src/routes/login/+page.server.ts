import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
  // 1. 쿠키 상자를 열어봅니다.
  const refreshToken = cookies.get('refresh_token');

  // 2. "어? 이미 로그인한 손님이시네요?"
  if (refreshToken) {
    // 3. 로비로 바로 모십니다.
    throw redirect(303, '/lobby');
  }
  
  // 4. 쿠키가 없으면 그냥 로그인 페이지(+page.svelte)를 보여줍니다.
};