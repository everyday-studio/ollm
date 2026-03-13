const GCS_BASE = 'https://storage.googleapis.com/ollm-assets-prod';

export const DEFAULT_GAME_THUMBNAIL = `${GCS_BASE}/default/game_thumbnail.png`;
export const DEFAULT_USER_PROFILE = `${GCS_BASE}/default/user_profile.png`;

export function handleImageError(fallbackUrl: string) {
	return (e: Event) => {
		const el = e.currentTarget as HTMLImageElement;
		el.onerror = null;
		el.src = fallbackUrl;
	};
}
