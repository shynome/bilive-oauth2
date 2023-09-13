<script lang="ts">
	import { dev } from '$app/environment'
	import bootstrap from '@remoon.net/theme/dist/bootstrap'
	import { onMount } from 'svelte'

	let devClass = ''
	if (dev) {
		// devClass = 'd-block show'
	}
	onMount(() => {
		return () => {
			bootstrap.Modal.getInstance('#help')?.dispose()
		}
	})
</script>

<div class="modal fade {devClass}" tabindex="-1" id="help">
	<div class="modal-dialog modal-dialog-centered">
		<div class="modal-content">
			<div class="modal-header">
				<h3 class="modal-title">介绍</h3>
				<button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close" />
			</div>
			<div class="modal-body">
				<p>通过直播间弹幕验证你的B站身份(UID), 用在"礼物查询"这些需要UID的应用中</p>
				<h5>登录流程</h5>
				<ol>
					<li>打开页面中指定的直播间</li>
					<li>发送页面中的验证弹幕</li>
					<li>验证成功后将会显示 "登录按钮"</li>
					<li>点击 "登录按钮" 登录完成</li>
				</ol>
				<h5>视频展示</h5>
				<p>
					<a
						href="https://www.bilibili.com/video/BV1EX4y1R7hY/"
						class="fs-5"
						target="bilibili-login-example"
					>
						点击打开B站 - 弹幕验证示例
					</a>
				</p>
				<h5>注意事项</h5>
				<ul>
					<li>刷新页面会更新验证弹幕</li>
				</ul>
			</div>
			<div class="modal-footer">
				<button type="button" class="btn btn-secondary" data-bs-dismiss="modal">了解</button>
			</div>
		</div>
	</div>
</div>
