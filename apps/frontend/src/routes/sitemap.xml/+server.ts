export function GET() {
	const pages = [
		{ url: '/', changefreq: 'daily', priority: 1.0 },
		{ url: '/login', changefreq: 'monthly', priority: 0.8 },
		{ url: '/lobby', changefreq: 'daily', priority: 0.9 },
		{ url: '/lobby/guide', changefreq: 'monthly', priority: 0.7 },
		{ url: '/lobby/leaderboard', changefreq: 'daily', priority: 0.8 },
	];

	const sitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${pages
	.map(
		(p) => `  <url>
    <loc>https://ollm.everydaystudio.xyz${p.url}</loc>
    <changefreq>${p.changefreq}</changefreq>
    <priority>${p.priority}</priority>
  </url>`
	)
	.join('\n')}
</urlset>`;

	return new Response(sitemap.trim(), {
		headers: {
			'Content-Type': 'application/xml',
			'Cache-Control': 'max-age=3600'
		}
	});
}
