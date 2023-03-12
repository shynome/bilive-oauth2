import type { Load } from "@sveltejs/kit"

export const ssr = false

export type Data = {
	whoami: string
}

export const load: Load = ({ fetch, depends }) => {
	depends("app:whoami")
	return {
		whoami: fetch("/bilive/whoami").then((r) => {
			if (r.status !== 200) {
				return null
			}
			return r.text()
		}),
	}
}
