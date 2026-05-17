<script lang="ts">
	import LoginForm from '$lib/components/LoginForm.svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	onMount(() => {
		if ($auth.token) goto('/');
	});
</script>

<LoginForm />
