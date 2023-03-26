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
	import Question from "./help/question-circle.svg"
	import HelpModal from "./help-modal.svelte"
	export let showHelp = false
</script>

<HelpModal bind:show={showHelp} />
<div class="card">
	<a href="#" class="help" on:click|preventDefault={() => (showHelp = true)}>
		<img src={Question} alt="help" width="35" height="35" />
	</a>
	<h4>直播间弹幕验证UID</h4>
	<hr />
	<div class="desc panel">
		{#if !roomid}
			<div class="reconnect">
				<button disabled>初始化中...</button>
			</div>
		{/if}
		{#if $v.closed}
			<div class="reconnect">
				<button on:click={() => $v.connect().catch(() => alert("重连失败"))}>重连</button>
			</div>
		{/if}
		<Clipboard let:copy text={verifyText} let:copied>
			<div class="vt clearfix">
				<input class="vtt button button-outline" value={verifyText} readonly />
				<button on:click={copy}>{copied ? "已" : ""}复制验证弹幕</button>
			</div>
		</Clipboard>
		<div class="one-click">
			<Clipboard let:copy text={verifyText}>
				<a
					class="button"
					href="https://live.bilibili.com/{roomid}"
					target="bilive_{roomid}"
					on:click={copy}
				>
					点击去直播间{roomid}发送验证弹幕
				</a>
			</Clipboard>
		</div>
		<div class="mobile-click">
			<Clipboard let:copy text={verifyText}>
				<a
					href="https://live.bilibili.com/{roomid}"
					target="bilive_{roomid}"
					class="button button-outline"
					on:click={copy}
				>
					{roomid}
				</a>
			</Clipboard>
			<Clipboard let:copy text={roomid} let:copied>
				<button on:click={copy}>{copied ? "已" : "点击"}复制直播间号</button>
			</Clipboard>
		</div>
	</div>
	<Danmu />
</div>

<style>
	.card {
		position: relative;
	}
	.help {
		position: absolute;
		right: 0;
		top: 0;
		font-size: 0;
	}
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
	.mobile-click {
		text-align: center;
	}
	.mobile-click a {
		width: 100%;
		margin-bottom: 1rem;
	}
	@media (max-width: 40rem) {
		.card {
			width: 95vw;
			padding: 1rem;
		}
		.one-click {
			display: none;
		}
	}
	@media (min-width: 40rem) {
		.mobile-click {
			display: none;
		}
	}
</style>
