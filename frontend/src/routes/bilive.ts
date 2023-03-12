import { readable } from "svelte/store"

type Msg = {
	info: [[], string, [number, string]]
}

export const verify = (() => {
	let t = {
		code: "",
		room: "",
		codes: [] as Msg[],
		closed: false,
		connect: () => Promise.resolve(),
	}
	let x = readable(t, (set) => {
		let p = "ws" + new URL("/bilive/pair", location.href).toString().slice("http".length)
		let ws: WebSocket
		function connect() {
			return new Promise<void>((rl, rj) => {
				if (typeof ws !== "undefined") {
					ws.onclose = () => 0
					ws.close()
				}
				ws = new WebSocket(p)
				ws.onclose = function () {
					if (this !== ws) {
						return
					}
					t.closed = true
					set(t)
				}
				ws.onopen = function () {
					if (this !== ws) {
						return
					}
					t.closed = false
					set(t)
					rl()
				}
				ws.onerror = (e) => {
					rj(e)
				}
				ws.onmessage = function (e) {
					let ini = JSON.parse(e.data)
					t.code = ini.code
					t.room = ini.room
					set(t)
					this.onmessage = (e) => {
						t.codes.push(JSON.parse(e.data))
						t.codes = t.codes.slice(-6)
						set(t)
					}
				}
			})
		}
		connect()
		t.connect = connect
		set(t)
		return () => {
			ws.close()
		}
	})
	return x
})()
