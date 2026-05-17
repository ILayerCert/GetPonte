<script lang="ts">
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	onMount(() => {
		if (!$auth.token) {
			goto('/login');
		}
	});
</script>

<div class="welcome">
	<div class="welcome-content">
		<span class="welcome-icon">🌉</span>
		<h2>Welcome to Ponte</h2>
		<p class="text-muted">Select a room from the sidebar to start collaborating</p>
		<div class="features">
			<div class="feature">
				<span class="feature-icon">💬</span>
				<span>Real-time chat</span>
			</div>
			<div class="feature">
				<span class="feature-icon">📹</span>
				<span>Video calls</span>
			</div>
			<div class="feature">
				<span class="feature-icon">🖥️</span>
				<span>Screen sharing</span>
			</div>
		</div>
	</div>
</div>

<style>
	.welcome {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.welcome-content {
		text-align: center;
		animation: fadeIn 0.6s ease;
	}

	.welcome-icon {
		font-size: 4rem;
		display: block;
		margin-bottom: 16px;
	}

	.welcome-content h2 {
		font-size: 1.6rem;
		font-weight: 600;
		margin-bottom: 8px;
	}

	.features {
		display: flex;
		gap: 24px;
		margin-top: 32px;
		justify-content: center;
	}

	.feature {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		padding: 16px 24px;
		background: var(--bg-secondary);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		font-size: 0.85rem;
		color: var(--text-secondary);
	}

	.feature-icon {
		font-size: 1.5rem;
	}
</style>
