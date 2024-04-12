<script lang="ts">
	import Authentication from './authentication.svelte'
	import Authorization from './authorization.svelte'
	import Danmu from './danmu.svelte'
	import type { PageData } from './$types'
	export let data: PageData
	import Help from './help.svelte'
</script>

<Help />
<div class="modal position-static d-block" tabindex="-1">
	<div class="modal-dialog modal-dialog-centered">
		<div class="modal-content">
			<div class="modal-header">
				<h5 class="modal-title">直播间弹幕验证OpenID</h5>
				<button
					type="button"
					class="btn btn-sm"
					data-bs-toggle="modal"
					data-bs-target="#help"
					aria-label="help"
				>
					<i class="bi bi-question-circle fs-5" />
				</button>
			</div>
			<div class="modal-body">
				{#if !!data.whoami}
					<Authorization whoami={data.whoami} token={data.token} />
				{:else}
					<Authentication />
					<Danmu />
				{/if}
			</div>
		</div>
	</div>
</div>

<style>
	.modal {
		height: 100vh;
	}
</style>
