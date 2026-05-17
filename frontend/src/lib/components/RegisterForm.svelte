<script lang="ts">
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	let username = '';
	let email = '';
	let password = '';

	async function handleRegister() {
		const success = await auth.register(username, email, password);
		if (success) {
			goto('/');
		}
	}
</script>

<div class="register-container">
	<div class="register-card card">
		<div class="logo">
			<span class="logo-icon">🌉</span>
			<h1>Ponte</h1>
			<p class="text-muted text-sm">Create your account</p>
		</div>

		<form on:submit|preventDefault={handleRegister} class="form">
			{#if $auth.error}
				<div class="error-banner">{$auth.error}</div>
			{/if}

			<div class="field">
				<label for="username">Username</label>
				<input
					id="username"
					type="text"
					bind:value={username}
					placeholder="johndoe"
					required
					minlength="3"
				/>
			</div>

			<div class="field">
				<label for="email">Email</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					placeholder="you@example.com"
					required
				/>
			</div>

			<div class="field">
				<label for="password">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					placeholder="••••••••"
					required
					minlength="6"
				/>
			</div>

			<button type="submit" class="btn btn-primary w-full" disabled={$auth.loading}>
				{#if $auth.loading}
					<span class="animate-pulse">Creating account...</span>
				{:else}
					Create Account
				{/if}
			</button>
		</form>

		<p class="switch-text">
			Already have an account? <a href="/login">Sign in</a>
		</p>
	</div>
</div>

<style>
	.register-container {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 100vh;
		padding: 20px;
		background: var(--bg-primary);
	}

	.register-card {
		width: 100%;
		max-width: 420px;
		padding: 40px;
		animation: fadeIn 0.5s ease;
	}

	.logo {
		text-align: center;
		margin-bottom: 32px;
	}

	.logo-icon {
		font-size: 3rem;
		display: block;
		margin-bottom: 8px;
	}

	.logo h1 {
		font-size: 1.8rem;
		font-weight: 700;
		background: linear-gradient(135deg, var(--accent), #7b9fd4);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.form {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.field label {
		font-size: 0.85rem;
		font-weight: 500;
		color: var(--text-secondary);
	}

	.error-banner {
		background: rgba(239, 68, 68, 0.1);
		border: 1px solid rgba(239, 68, 68, 0.3);
		color: var(--danger);
		padding: 10px 14px;
		border-radius: var(--radius);
		font-size: 0.85rem;
	}

	.switch-text {
		text-align: center;
		margin-top: 20px;
		font-size: 0.85rem;
		color: var(--text-muted);
	}
</style>
