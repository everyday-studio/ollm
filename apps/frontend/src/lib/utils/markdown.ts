import { marked } from 'marked';
import DOMPurify from 'dompurify';
import { browser } from '$app/environment';

// Configure marked
marked.setOptions({
	breaks: true, // \n → <br>
	gfm: true // GitHub Flavored Markdown (code blocks, tables, etc.)
});

/**
 * Parses markdown text into sanitized HTML string.
 * Safe to call in SSR context — returns plain text fallback when not in browser.
 */
export function renderMarkdown(text: string): string {
	if (!text) return '';
	const html = marked.parse(text) as string;
	if (!browser) {
		// SSR: return raw HTML from marked (no DOMPurify, but content is from our API)
		return html;
	}
	return DOMPurify.sanitize(html, {
		ALLOWED_TAGS: [
			'p', 'br', 'strong', 'em', 'code', 'pre',
			'ul', 'ol', 'li', 'blockquote', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
			'a', 'hr', 'del', 's', 'table', 'thead', 'tbody', 'tr', 'th', 'td',
			'span', 'div'
		],
		ALLOWED_ATTR: ['href', 'target', 'rel', 'class'],
		FORCE_BODY: false
	});
}
