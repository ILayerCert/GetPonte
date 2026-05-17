<script lang="ts">
	import { media } from '$lib/stores/media';
	import { auth } from '$lib/stores/auth';
	import { onMount } from 'svelte';

	let localVideoEl: HTMLVideoElement;

	$: remotePeers = $media.remotePeers;
	$: totalParticipants = remotePeers.length + ($media.localStream ? 1 : 0);
	$: gridCols = totalParticipants <= 1 ? 1 : totalParticipants <= 4 ? 2 : 3;

	onMount(() => {
		// Reactive update of local video
		const unsub = media.subscribe(state => {
			if (localVideoEl && state.localStream) {
				localVideoEl.srcObject = state.localStream;
			}
		});
		return unsub;
	});

	function attachStream(el: HTMLVideoElement, stream: MediaStream) {
		el.srcObject = stream;
	}
</script>

<div class="video-grid" style="--cols: {gridCols}">
	{#if !$media.inCall}
		<div class="no-call">
			<div class="no-call-content">
				<span class="no-call-icon">📹</span>
				<h3>No active call</h3>
				<p class="text-muted text-sm">Start a video call to connect with others</p>
			</div>
		</div>
	{:else}
		{#if remotePeers.length === 0 && $media.localStream}
			<!-- Only local user in call -->
			<div class="video-tile main-tile">
				<video
					bind:this={localVideoEl}
					autoplay
					muted
					playsinline
				></video>
				<div class="video-overlay">
					<span class="peer-name">{$auth.user?.username || 'You'}</span>
					<div class="status-icons">
						{#if !$media.audioEnabled}
							<span class="status-icon muted" title="Muted">🔇</span>
						{/if}
						{#if !$media.videoEnabled}
							<span class="status-icon camera-off" title="Camera off">📷</span>
						{/if}
					</div>
				</div>
				{#if !$media.videoEnabled}
					<div class="camera-off-overlay">
						<div class="avatar-placeholder">
							{($auth.user?.username || 'U').charAt(0).toUpperCase()}
						</div>
					</div>
				{/if}
			</div>
		{:else}
			<!-- Remote peers as main tiles -->
			{#each remotePeers as peer (peer.peerId)}
				<div class="video-tile">
					<video
						autoplay
						playsinline
						use:attachStream={peer.stream}
					></video>
					<div class="video-overlay">
						<span class="peer-name">{peer.peerName}</span>
					</div>
				</div>
			{/each}

			<!-- Local user as small PiP -->
			{#if $media.localStream}
				<div class="pip-video">
					<video
						bind:this={localVideoEl}
						autoplay
						muted
						playsinline
					></video>
					{#if !$media.videoEnabled}
						<div class="camera-off-overlay pip-overlay">
							<div class="avatar-placeholder small">
								{($auth.user?.username || 'U').charAt(0).toUpperCase()}
							</div>
						</div>
					{/if}
					<div class="pip-label">You</div>
				</div>
			{/if}
		{/if}
	{/if}
</div>

<style>
	.video-grid {
		flex: 1;
		display: grid;
		grid-template-columns: repeat(var(--cols), 1fr);
		gap: 8px;
		padding: 12px;
		position: relative;
		min-height: 0;
	}

	.no-call {
		grid-column: 1 / -1;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.no-call-content {
		text-align: center;
	}

	.no-call-icon {
		font-size: 3rem;
		display: block;
		margin-bottom: 12px;
		opacity: 0.4;
	}

	.no-call-content h3 {
		color: var(--text-secondary);
		font-weight: 500;
		margin-bottom: 4px;
	}

	.video-tile {
		position: relative;
		background: var(--bg-tertiary);
		border-radius: var(--radius-lg);
		overflow: hidden;
		aspect-ratio: 16 / 9;
		min-height: 0;
	}

	.main-tile {
		grid-column: 1 / -1;
		max-height: 70vh;
	}

	.video-tile video {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.video-overlay {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		padding: 8px 12px;
		background: linear-gradient(transparent, rgba(0, 0, 0, 0.7));
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.peer-name {
		font-size: 0.8rem;
		font-weight: 500;
		color: white;
		text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
	}

	.status-icons {
		display: flex;
		gap: 6px;
	}

	.status-icon {
		font-size: 0.9rem;
	}

	.camera-off-overlay {
		position: absolute;
		inset: 0;
		background: var(--bg-tertiary);
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.avatar-placeholder {
		width: 80px;
		height: 80px;
		border-radius: 50%;
		background: var(--accent);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 2rem;
		font-weight: 600;
		color: white;
	}

	.avatar-placeholder.small {
		width: 40px;
		height: 40px;
		font-size: 1rem;
	}

	.pip-video {
		position: absolute;
		bottom: 20px;
		right: 20px;
		width: 180px;
		aspect-ratio: 16 / 9;
		border-radius: var(--radius);
		overflow: hidden;
		box-shadow: var(--shadow-lg);
		border: 2px solid var(--border);
		z-index: 10;
	}

	.pip-video video {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transform: scaleX(-1);
	}

	.pip-overlay {
		border-radius: 0;
	}

	.pip-label {
		position: absolute;
		bottom: 4px;
		left: 8px;
		font-size: 0.65rem;
		color: white;
		text-shadow: 0 1px 3px rgba(0, 0, 0, 0.7);
		font-weight: 500;
	}
</style>
