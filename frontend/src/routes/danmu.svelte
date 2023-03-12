<script lang="ts">
	import { verify as v } from "./bilive"
</script>

<div class="danmu">
	<h5>直播间弹幕</h5>
	<ul>
		<li class="tip">
			弹幕延迟3s左右, 如被屏蔽<a href="#" on:click|preventDefault={() => $v.connect()}>
				点此更换验证弹幕
			</a>
		</li>
		{#if $v.codes.length === 0}
			<li>暂未收到任何验证弹幕</li>
		{:else}
			{#each $v.codes as { info }}
				<li>{info[2][1]}: {info[1]}</li>
			{/each}
		{/if}
	</ul>
</div>

<style>
	.danmu {
		margin-top: 3rem;
	}
	h5 {
		text-align: center;
		margin-bottom: 1rem;
	}
	ul {
		margin-bottom: 0;
	}
</style>
