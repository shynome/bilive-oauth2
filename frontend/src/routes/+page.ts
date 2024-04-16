export const ssr = true

type Info = {
	/**unix timestamp*/
	exp: number
	/**uid */
	sub: string
	nickname: string
}

import { get as getToken } from './token'
import { BROWSER } from 'esm-env'
export const load = () => {
	if (!BROWSER) {
		return {}
	}
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
		whoami: info.nickname ?? info.sub,
		token: token,
	}
}
