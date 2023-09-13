import { writable } from 'svelte/store'

type Msg = MsgInit | MsgDanmu | MsgVierfied
interface MsgInit {
	type: 'init'
	data: { code: string; room: string }
}
interface Danmu {
	uid: number
	content: string
}
interface MsgDanmu {
	type: 'danmu'
	data: Danmu
}
interface MsgVierfied {
	type: 'verified'
	data: { token: string }
}

import { save as saveToken } from './token'
import { invalidateAll } from '$app/navigation'

export const bilive = (() => {
	let ws: WebSocket
	let p = 'ws' + new URL('/bilive/pair', location.href).toString().slice('http'.length)
	const { subscribe, update } = writable(
		{
			code: '',
			room: '',
			codes: [] as Danmu[],
			closed: false,
			pending: true,
		},
		() => {
			connect()
			return () => {
				if (ws) {
					ws.close()
				}
			}
		},
	)

	async function connect(retry: boolean = false) {
		if (retry) {
			update((t) => {
				t.pending = true
				t.closed = false
				return t
			})
			await new Promise((rl) => setTimeout(rl, 5e2))
		}
		if (ws) {
			ws.onclose = () => 0
			ws.close()
		}
		ws = new WebSocket(p)
		ws.onclose = function () {
			update((t) => {
				t.closed = true
				t.pending = false
				return t
			})
		}
		ws.onerror = (e) => {
			console.error(e)
		}
		ws.onmessage = function (e) {
			let j = JSON.parse(e.data)
			msgSwitch(j)
		}
	}
	async function msgSwitch(j: Msg) {
		if (j.type === 'init') {
			let data = j.data
			update((t) => {
				t.code = data.code
				t.room = data.room
				t.pending = false
				t.closed = false
				return t
			})
		} else if (j.type == 'danmu') {
			let data = j.data
			update((t) => {
				t.codes.push(data)
				t.codes = t.codes.slice(-6)
				return t
			})
		} else if (j.type === 'verified') {
			saveToken(j.data.token)
			invalidateAll()
		}
	}
	return {
		subscribe,
		connect,
	}
})()
