export const ssr = false

type Info = {
	/**unix timestamp*/
	exp: number
	/**uid */
	sub: string
}

import { get as getToken } from './token'
export const load = () => {
	let token = getToken()
	if (!token) {
		return {}
	}
	let [_, b] = token.split('.')
	if (!b) {
		return {}
	}
	let info: Info = JSON.parse(atob(b))
	let now = Math.floor(Date.now() / 1e3)
	if (info.exp < now) {
		console.error('token is expired')
		return {}
	}
	return {
		whoami: info.sub,
		token: token,
	}
}
