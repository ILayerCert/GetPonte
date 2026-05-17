<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';
	import { rooms } from '$lib/stores/rooms';
	import { chat } from '$lib/stores/chat';
	import { media } from '$lib/stores/media';
	import { WebSocketClient } from '$lib/websocket';
	import { WebRTCManager } from '$lib/webrtc';
	import ChatPanel from '$lib/components/ChatPanel.svelte';
	import VideoGrid from '$lib/components/VideoGrid.svelte';
	import MediaControls from '$lib/components/MediaControls.svelte';

	let wsClient: WebSocketClient | null = null;
	let rtcManager: WebRTCManager | null = null;
	let roomId: string;
	let ready = false;

	$: roomId = $page.params.id;
	$: room = $rooms.activeRoom;

	onMount(async () => {
		if (!$auth.token) {
			goto('/login');
			return;
		}

		// 1. Join room first (ensures membership for WS)
		await rooms.joinRoom(roomId);

		// 2. Load room details
		await rooms.getRoom(roomId);

		// 3. Load chat history
		await chat.loadHistory(roomId);

		// 4. Connect WebSocket (requires membership)
		wsClient = new WebSocketClient($auth.token!, roomId);
		chat.connectWS(wsClient, roomId);
		wsClient.connect();

		ready = true;
	});

	onDestroy(() => {
		cleanup();
	});

	function cleanup() {
		if (rtcManager) {
			rtcManager.leaveCall();
			rtcManager = null;
		}
		if (wsClient) {
			wsClient.disconnect();
			wsClient = null;
		}
		media.stopAll();
		chat.clear();
	}

	async function handleJoinCall() {
		if (!wsClient || !$auth.user) return;

		const stream = await media.getUserMedia();
		if (!stream) return;

		rtcManager = new WebRTCManager(
			wsClient,
			$auth.user.id,
			(peerId, peerName, remoteStream) => {
				media.addRemotePeer(peerId, peerName, remoteStream);
			},
			(peerId) => {
				media.removeRemotePeer(peerId);
			}
		);

		rtcManager.addLocalStream(stream);
		await rtcManager.joinCall();
	}

	function handleLeaveCall() {
		if (rtcManager) {
			rtcManager.leaveCall();
			rtcManager = null;
		}
		media.stopAll();
	}
</script>

<svelte:head>
	<title>{room?.name ? `# ${room.name} — Ponte` : 'Ponte'}</title>
</svelte:head>

{#if !ready}
	<div class="room-loading">
		<span class="animate-pulse">Loading room...</span>
	</div>
{:else}
	<div class="room-view">
		<div class="room-header">
			<div class="room-info">
				<span class="room-hash">#</span>
				<h2>{room?.name || 'Room'}</h2>
				{#if room?.description}
					<span class="room-separator">|</span>
					<span class="room-desc text-muted text-sm">{room.description}</span>
				{/if}
			</div>
			{#if room?.member_count}
				<div class="room-meta">
					<span class="member-count">
						👥 {room.member_count} member{room.member_count !== 1 ? 's' : ''}
					</span>
				</div>
			{/if}
		</div>

		<div class="room-body">
			<div class="video-area">
				<VideoGrid />
				<MediaControls
					inCall={$media.inCall}
					on:join={handleJoinCall}
					on:leave={handleLeaveCall}
				/>
			</div>
			<ChatPanel {roomId} />
		</div>
	</div>
{/if}

<style>
	.room-loading {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		color: var(--text-muted);
		font-size: 1rem;
	}

	.room-view {
		display: flex;
		flex-direction: column;
		height: 100%;
	}

	.room-header {
		padding: 12px 20px;
		border-bottom: 1px solid var(--border);
		display: flex;
		align-items: center;
		justify-content: space-between;
		background: var(--bg-secondary);
		flex-shrink: 0;
	}

	.room-info {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.room-hash {
		color: var(--text-muted);
		font-size: 1.2rem;
		font-weight: 700;
	}

	.room-info h2 {
		font-size: 1rem;
		font-weight: 600;
	}

	.room-separator {
		color: var(--border-light);
	}

	.room-desc {
		max-width: 300px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.room-meta {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.member-count {
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	.room-body {
		flex: 1;
		display: flex;
		overflow: hidden;
	}

	.video-area {
		flex: 1;
		display: flex;
		flex-direction: column;
		min-width: 0;
		background: var(--bg-primary);
	}
</style>
