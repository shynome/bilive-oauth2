<script lang="ts">
	import { invalidate } from "$app/navigation"
	import { page } from "$app/stores"

	export let whoami: string
	let logout_pending = false
	function logout() {
		logout_pending = true
		Promise.resolve()
			.then(async () => {
				await fetch("/bilive/logout")
				invalidate("app:whoami")
			})
			.finally(() => {
				logout_pending = false
			})
	}
</script>

<div class="card">
	<h4>直播间弹幕验证UID</h4>
	<hr />
	<div class="desc">
		<div class="vt">
			<div class="vtt button button-outline">
				UID已验证 {whoami}
			</div>
			<button disabled={logout_pending} on:click={logout}>退出</button>
		</div>
		<div class="one-click">
			<a class="button" href="/oauth/allow">授权</a>
		</div>
	</div>
</div>

<style>
	h4 {
		text-align: center;
	}
	.card {
		border: 1px solid #0000003d;
		padding: 3rem;
		border-radius: 1rem;
		width: 60rem;
		box-shadow: 1px 1px 2px 0px #0000003d;
	}
	.vt {
		display: flex;
		margin-bottom: 2rem;
	}
	.vt .vtt {
		flex: 1;
		text-align: center;
		border: 1px solid;
		margin-right: 1rem;
		/* text-overflow: ellipsis; */
	}
	.tip {
		font-size: 1.5rem;
	}
	button,
	.button {
		margin-bottom: 0;
	}
	.one-click {
		text-align: center;
	}
	@media (max-width: 40rem) {
		.card {
			width: 95vw;
			padding: 1rem;
		}
	}
</style>
