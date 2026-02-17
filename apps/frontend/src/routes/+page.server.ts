import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ cookies }) => {
  const refreshToken = cookies.get('refresh_token');

  if (refreshToken) {
    // 로그인 되어 있으면 -> 로비로
    throw redirect(303, '/lobby');
  } else {
    // 로그인 안 되어 있으면 -> 로그인 페이지로
    throw redirect(303, '/login');
  }
};