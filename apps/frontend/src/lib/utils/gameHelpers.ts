import type { MatchStatus } from '$lib/features/game/types';

export function getJudgeBadgeStyle(judgeType: string): { label: string; classes: string } {
	switch (judgeType) {
		case 'target_word':
			return {
				label: 'Target Word',
				classes: 'bg-purple-500/90 text-white border border-purple-300/40'
			};
		case 'llm_judge':
			return {
				label: 'LLM Judge',
				classes: 'bg-blue-500/90 text-white border border-blue-300/40'
			};
		case 'format_break':
			return {
				label: 'Format Break',
				classes: 'bg-orange-500/90 text-white border border-orange-300/40'
			};
		default:
			return {
				label: 'Unknown',
				classes: 'bg-gray-500/90 text-white border border-gray-300/40'
			};
	}
}

export function getStatusLabel(status: MatchStatus | undefined): string {
	switch (status) {
		case 'active':
			return '진행 중';
		case 'generating':
			return 'AI 응답 중...';
		case 'won':
			return '승리';
		case 'lost':
			return '패배';
		case 'resigned':
			return '기권';
		case 'expired':
			return '만료';
		case 'error':
			return '오류';
		default:
			return '—';
	}
}

export function getStatusColor(status: MatchStatus | undefined): string {
	switch (status) {
		case 'active':
			return 'bg-gray-500/20 text-gray-400 border-gray-500/30';
		case 'generating':
			return 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30';
		case 'won':
			return 'bg-green-500/20 text-green-400 border-green-500/30';
		case 'lost':
			return 'bg-red-500/20 text-red-400 border-red-500/30';
		case 'resigned':
			return 'bg-gray-500/20 text-gray-400 border-gray-500/30';
		case 'expired':
			return 'bg-orange-500/20 text-orange-400 border-orange-500/30';
		case 'error':
			return 'bg-red-500/20 text-red-400 border-red-500/30';
		default:
			return 'bg-gray-500/20 text-gray-400 border-gray-500/30';
	}
}

export function getShortStatusLabel(status: MatchStatus): string {
	switch (status) {
		case 'active':
			return '진행 중';
		case 'generating':
			return '생성 중';
		case 'won':
			return '승리';
		case 'lost':
			return '패배';
		case 'resigned':
			return '기권';
		case 'expired':
			return '만료';
		case 'error':
			return '오류';
		default:
			return status;
	}
}
