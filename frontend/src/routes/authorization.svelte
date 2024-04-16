<script lang="ts" context="module">
	function getHostname(u: URL) {
		let r = u.searchParams.get('redirect_uri')
		if (!r) {
			return r
		}
		let f = new URL(r)
		return f.hostname
	}
</script>

<script lang="ts">
	import { invalidateAll } from '$app/navigation'

	import { page } from '$app/stores'
	export let whoami: string
	$: host = getHostname($page.url)
	export let token: string
	import { clear as clearToken } from './token'
</script>

<form method="post" action="/oauth/authorize">
	<div class="input-group">
		<input class="form-control text-center" readonly type="text" value="用户已验证 {whoami}" />
		<button
			type="button"
			class="btn btn-outline-danger"
			on:click={() => {
				clearToken()
				invalidateAll()
			}}
		>
			退出
		</button>
	</div>
	<div class="text-center my-3">
		{#each $page.url.searchParams as [k, v]}
			<input type="hidden" name={k} value={v} />
		{/each}
		<input type="hidden" name="bilive-token" value={token} />
		<button type="submit" disabled={!host} class="btn btn-primary">登录</button>
	</div>
</form>
