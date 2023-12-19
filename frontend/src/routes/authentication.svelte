<script context="module" lang="ts">
	import { writable } from 'svelte/store'

	function createCopiedTimer(timeout = 2e3) {
		let timer: number
		let { subscribe, set } = writable(false, () => {
			return () => {
				clearTimeout(timer)
			}
		})
		return {
			subscribe,
			set: (v: boolean) => {
				set(v)
				clearTimeout(timer)
				timer = setTimeout(() => {
					set(false)
				}, timeout)
			},
		}
	}
</script>

<script lang="ts">
	import { bilive } from './bilive'
	$: roomid = $bilive.room
	import { copy, copyText } from 'svelte-copy'
	let danmuCopied = createCopiedTimer()
	let roomCopied = createCopiedTimer()
	// @ts-ignore
	import Clipboard from 'svelte-clipboard'
</script>

<div class="root position-relative">
	<div
		class="cover bg-body d-flex flex-column justify-content-center"
		class:invisible={!$bilive.pending}
	>
		<div class="text-center">
			<button class="btn btn-outline-primary" disabled>初始化中...</button>
		</div>
	</div>
	<div
		class="cover bg-body d-flex flex-column justify-content-center"
		class:invisible={!$bilive.closed}
	>
		<div class="text-center">
			<button
				type="button"
				class="btn btn-lg btn-primary"
				disabled={$bilive.pending}
				on:click={() => bilive.connect(true)}
			>
				重连
			</button>
		</div>
	</div>
	<div class="input-group">
		<input class="form-control text-center" type="text" readonly value={$bilive.code} />
		<button
			class="btn btn-primary"
			use:copy={$bilive.code}
			on:svelte-copy={() => {
				$danmuCopied = true
			}}
		>
			{$danmuCopied ? '已' : ''}复制验证弹幕
		</button>
	</div>
	<div class="d-block d-sm-none text-center my-3">
		<div class="input-group">
			<input class="form-control text-center" type="text" readonly value={roomid} />
			<button
				class="btn btn-primary"
				use:copy={roomid}
				on:svelte-copy={() => {
					$roomCopied = true
				}}
			>
				{$roomCopied ? '已' : ''}复制直播间号
			</button>
		</div>
		<Clipboard let:copy text={$bilive.code}>
			<a
				class="btn btn-primary my-2"
				href="https://live.bilibili.com/{roomid}"
				target="bilive_{roomid}"
				on:click={() => {
					copy()
				}}
			>
				打开上方直播间, 发送验证弹幕后<br />
				再回到此页面即可看到登录成功
			</a>
		</Clipboard>
	</div>
	<div class="d-none d-sm-block text-center my-3">
		<Clipboard let:copy text={$bilive.code}>
			<a
				href="https://live.bilibili.com/{roomid}"
				class="btn btn-primary"
				target="bilive_{roomid}"
				on:click={() => {
					copy()
				}}
			>
				点击去直播间{roomid}发送验证弹幕
			</a>
		</Clipboard>
	</div>
</div>

<style>
	.cover {
		--bs-bg-opacity: 0.8;
		position: absolute;
		z-index: 3;
		left: 0;
		top: 0;
		height: 100%;
		width: 100%;
	}
</style>
