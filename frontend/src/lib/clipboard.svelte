<script lang="ts">
	// @ts-ignore
	import Clipboard from "svelte-clipboard"
	export let text: string
	let copied = false
	const handleCopy = () => {
		copied = true
		setTimeout(() => {
			copied = false
		}, 2e3)
	}
</script>

<Clipboard let:copy {text} on:copy={handleCopy}>
	<slot {copy} {copied} />
</Clipboard>
