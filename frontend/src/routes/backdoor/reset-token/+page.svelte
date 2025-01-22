<svelte:head>
	<title>重置Token</title>
</svelte:head>
<div class="container">
	<form method="get">
		<div class="my-3">
			<label for="token" class="form-label">Token</label>
			<input type="text" name="token" id="token" class="form-control" />
			<div class="form-text">输入要替换的Token</div>
		</div>
		<div class="my-3 text-center">
			<button class="btn btn-primary w-100">重置Token</button>
		</div>
	</form>
</div>
