<script lang="ts">
	import { media } from '$lib/stores/media';
	import { createEventDispatcher } from 'svelte';

	const dispatch = createEventDispatcher();

	export let inCall: boolean = false;

	function handleToggleMic() {
		media.toggleAudio();
	}

	function handleToggleCamera() {
		media.toggleVideo();
	}

	async function handleToggleScreen() {
		await media.toggleScreenShare();
	}

	function handleJoinCall() {
		dispatch('join');
	}

	function handleLeaveCall() {
		dispatch('leave');
	}
</script>

<div class="media-controls">
	{#if inCall}
		<div class="controls-group">
			<button
				class="control-btn"
				class:off={!$media.audioEnabled}
				on:click={handleToggleMic}
				title={$media.audioEnabled ? 'Mute microphone' : 'Unmute microphone'}
			>
				<span class="control-icon">{$media.audioEnabled ? '🎤' : '🔇'}</span>
				<span class="control-label">{$media.audioEnabled ? 'Mic' : 'Muted'}</span>
			</button>

			<button
				class="control-btn"
				class:off={!$media.videoEnabled}
				on:click={handleToggleCamera}
				title={$media.videoEnabled ? 'Turn off camera' : 'Turn on camera'}
			>
				<span class="control-icon">{$media.videoEnabled ? '📷' : '📷'}</span>
				<span class="control-label">{$media.videoEnabled ? 'Camera' : 'Off'}</span>
			</button>

			<button
				class="control-btn"
				class:active={$media.screenSharing}
				on:click={handleToggleScreen}
				title={$media.screenSharing ? 'Stop sharing' : 'Share screen'}
			>
				<span class="control-icon">🖥️</span>
				<span class="control-label">{$media.screenSharing ? 'Stop' : 'Screen'}</span>
			</button>

			<div class="divider"></div>

			<button
				class="control-btn leave"
				on:click={handleLeaveCall}
				title="Leave call"
			>
				<span class="control-icon">📞</span>
				<span class="control-label">Leave</span>
			</button>
		</div>
	{:else}
		<button class="join-call-btn" on:click={handleJoinCall}>
			<span>📹</span> Join Call
		</button>
	{/if}
</div>

<style>
	.media-controls {
		padding: 12px 20px;
		background: var(--bg-secondary);
		border-top: 1px solid var(--border);
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.controls-group {
		display: flex;
		align-items: center;
		gap: 8px;
		background: var(--bg-primary);
		padding: 6px 12px;
		border-radius: 40px;
		border: 1px solid var(--border);
	}

	.control-btn {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2px;
		padding: 10px 16px;
		border-radius: var(--radius-lg);
		background: transparent;
		color: var(--text-primary);
	}

	.control-btn:hover {
		background: var(--bg-hover);
	}

	.control-btn.off {
		color: var(--danger);
		background: rgba(239, 68, 68, 0.1);
	}

	.control-btn.off:hover {
		background: rgba(239, 68, 68, 0.2);
	}

	.control-btn.active {
		color: var(--success);
		background: rgba(74, 222, 128, 0.1);
	}

	.control-btn.leave {
		color: var(--danger);
	}

	.control-btn.leave:hover {
		background: rgba(239, 68, 68, 0.15);
	}

	.control-icon {
		font-size: 1.3rem;
		line-height: 1;
	}

	.control-label {
		font-size: 0.65rem;
		font-weight: 500;
	}

	.divider {
		width: 1px;
		height: 32px;
		background: var(--border);
		margin: 0 4px;
	}

	.join-call-btn {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 28px;
		border-radius: 40px;
		background: var(--success);
		color: #0f0f23;
		font-weight: 600;
		font-size: 0.95rem;
	}

	.join-call-btn:hover {
		background: #22c55e;
		transform: scale(1.02);
		box-shadow: 0 4px 20px rgba(74, 222, 128, 0.3);
	}
</style>
