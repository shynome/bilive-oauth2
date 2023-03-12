<script lang="ts">
	import { invalidate } from "$app/navigation"
	import Clipboard from "$lib/clipboard.svelte"
	import { verify as v } from "./bilive"
	$: verifyText = $v.code
	$: roomid = $v.room
	$: {
		if ($v.closed) {
			invalidate("app:whoami")
		}
	}
	import Danmu from "./danmu.svelte"
</script>

<div class="card">
	<h4>直播间弹幕验证UID</h4>
	<hr />
	<div class="desc panel">
		{#if $v.closed}
			<div class="reconnect">
				<button on:click={() => $v.connect().catch(() => alert("重连失败"))}>重连</button>
			</div>
		{/if}
		<Clipboard let:copy text={verifyText} let:copied>
			<div class="vt clearfix">
				<input class="vtt button button-outline" value={verifyText} readonly />
				<button on:click={copy}>{copied ? "已" : "点击"}复制</button>
			</div>
		</Clipboard>
		<div class="one-click">
			<Clipboard let:copy text={verifyText}>
				<a
					class="button"
					href="https://live.bilibili.com/{roomid}"
					on:click={copy}
					target="bilive_{roomid}"
				>
					点击复制, 去直播间{roomid}发送验证弹幕
				</a>
			</Clipboard>
		</div>
	</div>
	<Danmu />
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
		text-transform: unset;
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
	.panel {
		position: relative;
	}
	.reconnect {
		position: absolute;
		background-color: #000000a8;
		left: 0;
		right: 0;
		width: 100%;
		height: 100%;
		display: flex;
		justify-content: center;
		align-items: center;
	}
	@media (max-width: 40rem) {
		.card {
			width: 95vw;
			padding: 1rem;
		}
	}
</style>
