export const ssr = true

type Info = JwtPayload & {
	nickname: string
}

import { get as getToken } from './token'
import { BROWSER } from 'esm-env'
import { jwtDecode, type JwtPayload } from 'jwt-decode'
export const load = () => {
	if (!BROWSER) {
		return {}
	}
	let token = getToken()
	if (!token) {
		return {}
	}
	try {
		const info = jwtDecode<Info>(token)
		let now = Math.floor(Date.now() / 1e3)
		if (info.exp! < now) {
			console.error('token is expired')
			return {}
		}
		return {
			whoami: info.nickname ?? info.sub,
			token: token,
		}
	} catch (err) {
		console.error(err)
		return {}
	}
}
