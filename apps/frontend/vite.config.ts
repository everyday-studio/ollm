import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		host: true, // 도커 외부 접속 허용 (0.0.0.0)
		watch: {
			usePolling: true // [중요] 파일 변경 감지 강제 활성화
		}
	}
});
