import { defineConfig } from "vitepress";

// https://vitepress.dev/reference/site-config
export default defineConfig({
	title: "lsgo",
	description: "A modern, colourful replacement for `ls`, written in Go with no external dependencies.",
	head: [
		["link", { rel: "icon", type: "image/svg+xml", href: "/logo.svg" }],
		[
			"meta",
			{
				name: "description",
				content:
					"lsgo is a modern, colourful ls replacement written in Go with zero external dependencies. Grid, one-per-line, long-table, and tree views, Git status, LS_COLORS theming, and Nerd Font icons.",
			},
		],
		["meta", { property: "og:title", content: "lsgo" }],
		[
			"meta",
			{
				property: "og:description",
				content:
					"A modern, colourful ls replacement written in Go with zero external dependencies.",
			},
		],
		["meta", { property: "og:type", content: "website" }],
		["meta", { property: "og:url", content: "https://lsgo.jc21.com/" }],
		["meta", { name: "twitter:card", content: "summary" }],
		["meta", { name: "twitter:title", content: "lsgo" }],
		[
			"meta",
			{
				name: "twitter:description",
				content:
					"A modern, colourful ls replacement written in Go with zero external dependencies.",
			},
		],
		["meta", { name: "twitter:alt", content: "lsgo" }],
		[
			"script",
			{
				async: "true",
				src: "https://www.googletagmanager.com/gtag/js?id=G-TWRY418VLV",
			},
		],
	],
	sitemap: {
		hostname: "https://lsgo.jc21.com",
	},
	metaChunk: true,
	srcDir: "./src",
	outDir: "./dist",
	themeConfig: {
		// https://vitepress.dev/reference/default-theme-config
		logo: { src: "/logo.svg", width: 24, height: 24 },
		nav: [{ text: "Guide", link: "/guide/" }],
		sidebar: [
			{
				items: [
					{ text: "Guide", link: "/guide/" },
					{ text: "Installation", link: "/installation/" },
					{ text: "Usage & Flags", link: "/usage/" },
					{ text: "View Modes & Sorting", link: "/view-modes/" },
					{ text: "Colours & Theming", link: "/colours/" },
					{ text: "Frequently Asked Questions", link: "/faq/" },
				],
			},
		],
		socialLinks: [
			{
				icon: "github",
				link: "https://github.com/jc21/lsgo",
			},
		],
		search: {
			provider: "local",
		},
		footer: {
			copyright: "Copyright © 2026-present jc21.com",
		},
	},
});
