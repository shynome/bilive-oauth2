import { sveltekit } from '@sveltejs/kit/vite'
import { defineConfig } from 'vite'
import { isoImport } from 'vite-plugin-iso-import'

export default defineConfig({
	build: {
		target: 'es2015',
	},
	plugins: [isoImport(), sveltekit()],
	server: {
		proxy: {
			'/bilive': {
				target: 'http://127.0.0.1:9096',
				ws: true,
			},
			'/oauth': 'http://127.0.0.1:9096',
		},
	},
})
