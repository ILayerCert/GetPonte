<script lang="ts">
	import RegisterForm from '$lib/components/RegisterForm.svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	onMount(() => {
		if ($auth.token) goto('/');
	});
</script>

<RegisterForm />
