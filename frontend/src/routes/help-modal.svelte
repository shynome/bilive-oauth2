<script lang="ts">
	import Help from "./help/+page.svelte"
	import CloseIcon from "./help/close-circle.svg"
	export let show: boolean = false
</script>

<div class="modal" class:show>
	<div class="main">
		<a href="#" class="close" on:click|preventDefault={() => (show = false)}>
			<img src={CloseIcon} alt="close" width="30" height="30" />
		</a>
		<div class="container">
			<Help />
		</div>
	</div>
</div>

<style>
	.modal {
		position: fixed;
		top: 0;
		left: 0;
		overflow-y: auto;
		height: 100vh;
		width: 100vw;
		padding: 5vw;
		background-color: #fff;
		z-index: 9;
		display: flex;
		align-items: center;
		justify-content: center;
		visibility: hidden;
    opacity: 0;
    transition: all 0.5s;
	}
	.modal.show {
		visibility: visible;
    opacity: 1;
	}
	.main {
		position: relative;
	}
	.close {
		position: absolute;
		right: 0;
		top: 0;
    z-index: 9;
	}
</style>
