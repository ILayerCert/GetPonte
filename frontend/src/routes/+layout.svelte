<script lang="ts">
	import '../app.css';
	import { auth } from '$lib/stores/auth';
	import { rooms } from '$lib/stores/rooms';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import RoomSidebar from '$lib/components/RoomSidebar.svelte';

	$: isAuthPage = $page.url.pathname === '/login' || $page.url.pathname === '/register';
	$: isLoggedIn = !!$auth.token;

	onMount(() => {
		if (isLoggedIn && !isAuthPage) {
			rooms.loadRooms();
		}
	});

	$: if (isLoggedIn && !isAuthPage) {
		rooms.loadRooms();
	}
</script>

{#if isAuthPage || !isLoggedIn}
	<slot />
{:else}
	<div class="app-layout">
		<RoomSidebar />
		<main class="main-content">
			<slot />
		</main>
	</div>
{/if}

<style>
	.app-layout {
		display: flex;
		height: 100vh;
		overflow: hidden;
	}

	.main-content {
		flex: 1;
		display: flex;
		flex-direction: column;
		min-width: 0;
		overflow: hidden;
	}
</style>
