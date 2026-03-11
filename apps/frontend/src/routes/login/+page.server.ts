import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies, url }) => {
  // 1. 쿠키 상자를 열어봅니다.
  const refreshToken = cookies.get('refresh_token');
  const clearSession = url.searchParams.get('clear') === 'true';

  // 2. "어? 이미 로그인한 손님이시네요?" (단, 강제로 세션을 지우고 들어온게 아니라면)
  if (refreshToken && !clearSession) {
    // 3. 로비로 바로 모십니다.
    throw redirect(303, '/lobby');
  }

  // 4. 쿠키가 없거나 리셋되었다면 그냥 로그인 페이지(+page.svelte)를 보여줍니다.
};