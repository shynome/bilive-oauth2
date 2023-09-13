<script>
	import { bilive } from './bilive'
	$: codes = $bilive.codes
</script>

<div class="danmu mt-4">
	<h5 class="text-center">直播间弹幕</h5>
	<ul>
		<li class="tip">
			弹幕延迟3s左右, 如被屏蔽
			<a
				href="#"
				class="text-decoration-none"
				class:link-secondary={$bilive.pending}
				on:click|preventDefault={() => {
					if ($bilive.pending) {
						return
					}
					bilive.connect(true)
				}}
			>
				点此更换验证弹幕
			</a>
		</li>
		{#if codes.length === 0}
			<li>暂未收到任何验证弹幕</li>
		{:else}
			{#each codes as d}
				<li>danmu: {d.content}</li>
			{/each}
		{/if}
	</ul>
</div>
