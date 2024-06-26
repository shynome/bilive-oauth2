<script>
	import { PUBLIC_TIANJI_ID, PUBLIC_TIANJI_SITE } from '$env/static/public'
</script>

<svelte:head>
	{#if PUBLIC_TIANJI_ID}
		<script async defer src={PUBLIC_TIANJI_SITE} data-website-id={PUBLIC_TIANJI_ID}></script>
	{/if}
</svelte:head>

<slot />

<style>
	@import '@remoon.net/theme/dist/bootstrap.scss';
	@import 'bootstrap-icons/font/bootstrap-icons.min.css';
</style>
