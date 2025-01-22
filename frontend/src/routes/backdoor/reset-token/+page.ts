import { browser } from '$app/environment'
import { save } from '$lib/../routes/token.js'
import { redirect } from '@sveltejs/kit'

export function load({ url }) {
	if (!browser) {
		return {} as never
	}
	if (url.searchParams.has('token')) {
		save(url.searchParams.get('token')!)
		throw redirect(307, '/')
	}
}
