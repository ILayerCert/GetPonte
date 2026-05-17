<script lang="ts">
	export let username: string;
	export let size: 'sm' | 'md' | 'lg' = 'md';
	export let showName: boolean = true;

	function getInitial(name: string): string {
		return name.charAt(0).toUpperCase();
	}

	function getColor(name: string): string {
		const colors = ['#4A6FA5', '#6366f1', '#8b5cf6', '#ec4899', '#f97316', '#14b8a6', '#84cc16'];
		let hash = 0;
		for (let i = 0; i < name.length; i++) {
			hash = name.charCodeAt(i) + ((hash << 5) - hash);
		}
		return colors[Math.abs(hash) % colors.length];
	}
</script>

<div class="user-badge {size}">
	<div class="avatar" style="background-color: {getColor(username)}">
		{getInitial(username)}
	</div>
	{#if showName}
		<span class="name">{username}</span>
	{/if}
</div>

<style>
	.user-badge {
		display: inline-flex;
		align-items: center;
		gap: 8px;
	}

	.avatar {
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		color: white;
		flex-shrink: 0;
	}

	.sm .avatar { width: 24px; height: 24px; font-size: 0.65rem; }
	.md .avatar { width: 32px; height: 32px; font-size: 0.8rem; }
	.lg .avatar { width: 40px; height: 40px; font-size: 1rem; }

	.name {
		font-size: 0.85rem;
		color: var(--text-primary);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.sm .name { font-size: 0.75rem; }
</style>
