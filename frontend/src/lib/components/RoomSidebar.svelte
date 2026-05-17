<script lang="ts">
	import { rooms } from '$lib/stores/rooms';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import UserBadge from './UserBadge.svelte';

	let showCreateModal = false;
	let newRoomName = '';
	let newRoomDesc = '';

	$: currentRoomId = $page.params?.id || '';

	async function handleCreate() {
		if (!newRoomName.trim()) return;
		const room = await rooms.createRoom(newRoomName.trim(), newRoomDesc.trim());
		if (room) {
			showCreateModal = false;
			newRoomName = '';
			newRoomDesc = '';
			goto(`/rooms/${room.id}`);
		}
	}

	function handleLogout() {
		auth.logout();
		goto('/login');
	}
</script>

<aside class="sidebar">
	<div class="sidebar-header">
		<div class="brand">
			<span class="brand-icon">🌉</span>
			<span class="brand-name">Ponte</span>
		</div>
	</div>

	<div class="sidebar-section">
		<div class="section-header">
			<span class="section-title">Rooms</span>
			<button class="btn-add" on:click={() => showCreateModal = true} title="Create room">
				+
			</button>
		</div>

		<div class="room-list">
			{#if $rooms.loading && $rooms.rooms.length === 0}
				<div class="room-empty">
					<span class="animate-pulse text-muted text-sm">Loading rooms...</span>
				</div>
			{:else if $rooms.rooms.length === 0}
				<div class="room-empty">
					<span class="text-muted text-sm">No rooms yet</span>
					<button class="btn btn-secondary text-sm" on:click={() => showCreateModal = true}>
						Create one
					</button>
				</div>
			{:else}
				{#each $rooms.rooms as room (room.id)}
					<button
						class="room-item"
						class:active={currentRoomId === room.id}
						on:click={() => goto(`/rooms/${room.id}`)}
					>
						<span class="room-hash">#</span>
						<span class="room-name">{room.name}</span>
						{#if room.member_count}
							<span class="room-count">{room.member_count}</span>
						{/if}
					</button>
				{/each}
			{/if}
		</div>
	</div>

	<div class="sidebar-footer">
		{#if $auth.user}
			<UserBadge username={$auth.user.username} size="sm" />
			<button class="btn-ghost btn-icon logout-btn" on:click={handleLogout} title="Sign out">
				⏻
			</button>
		{/if}
	</div>
</aside>

{#if showCreateModal}
	<div class="modal-overlay" on:click|self={() => showCreateModal = false}>
		<div class="modal card animate-fadeIn">
			<h3>Create Room</h3>
			<form on:submit|preventDefault={handleCreate} class="modal-form">
				<div class="field">
					<label for="room-name">Room Name</label>
					<input id="room-name" bind:value={newRoomName} placeholder="general" required />
				</div>
				<div class="field">
					<label for="room-desc">Description (optional)</label>
					<input id="room-desc" bind:value={newRoomDesc} placeholder="What's this room about?" />
				</div>
				<div class="modal-actions">
					<button type="button" class="btn btn-secondary" on:click={() => showCreateModal = false}>
						Cancel
					</button>
					<button type="submit" class="btn btn-primary" disabled={!newRoomName.trim()}>
						Create
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<style>
	.sidebar {
		width: 260px;
		height: 100vh;
		background: var(--bg-secondary);
		border-right: 1px solid var(--border);
		display: flex;
		flex-direction: column;
		flex-shrink: 0;
	}

	.sidebar-header {
		padding: 16px 16px 12px;
		border-bottom: 1px solid var(--border);
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.brand-icon {
		font-size: 1.5rem;
	}

	.brand-name {
		font-size: 1.2rem;
		font-weight: 700;
		background: linear-gradient(135deg, var(--accent), #7b9fd4);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.sidebar-section {
		flex: 1;
		overflow-y: auto;
		padding: 12px 0;
	}

	.section-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 4px 16px 8px;
	}

	.section-title {
		font-size: 0.7rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-muted);
	}

	.btn-add {
		width: 20px;
		height: 20px;
		border-radius: 4px;
		background: transparent;
		color: var(--text-muted);
		font-size: 1rem;
		display: flex;
		align-items: center;
		justify-content: center;
		line-height: 1;
	}

	.btn-add:hover {
		background: var(--bg-hover);
		color: var(--text-primary);
	}

	.room-list {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: 0 8px;
	}

	.room-empty {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		padding: 24px 16px;
	}

	.room-item {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 12px;
		border-radius: var(--radius);
		background: transparent;
		color: var(--text-secondary);
		font-size: 0.9rem;
		text-align: left;
		width: 100%;
	}

	.room-item:hover {
		background: var(--bg-hover);
		color: var(--text-primary);
	}

	.room-item.active {
		background: var(--bg-active);
		color: var(--text-primary);
	}

	.room-hash {
		color: var(--text-muted);
		font-weight: 600;
		font-size: 1rem;
	}

	.room-name {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.room-count {
		font-size: 0.7rem;
		background: var(--bg-primary);
		color: var(--text-muted);
		padding: 1px 6px;
		border-radius: 10px;
	}

	.sidebar-footer {
		padding: 12px 16px;
		border-top: 1px solid var(--border);
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.logout-btn {
		font-size: 1rem;
		width: 32px;
		height: 32px;
	}

	/* Modal */
	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.7);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 100;
		backdrop-filter: blur(4px);
	}

	.modal {
		width: 90%;
		max-width: 440px;
		padding: 28px;
	}

	.modal h3 {
		font-size: 1.1rem;
		font-weight: 600;
		margin-bottom: 20px;
	}

	.modal-form {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.field label {
		font-size: 0.8rem;
		font-weight: 500;
		color: var(--text-secondary);
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 10px;
		margin-top: 8px;
	}
</style>
