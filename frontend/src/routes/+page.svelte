<script lang="ts">
	import type { Data } from "./+page"
	export let data: Data
	import Authentication from "./authentication.svelte"
	import Authorization from "./authorization.svelte"
</script>

<div class="wh">
	<div class="x">
		{#if data.whoami}
			<Authorization whoami={data.whoami} />
		{:else}
			<Authentication />
		{/if}
	</div>
</div>

<style>
	.wh {
		display: flex;
		height: 100vh;
		width: 100vw;
		justify-content: center;
		align-items: center;
	}
</style>
