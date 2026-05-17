<script lang="ts">
	import { chat, type ChatMessage } from '$lib/stores/chat';
	import { auth } from '$lib/stores/auth';
	import { onMount, afterUpdate, tick } from 'svelte';
	import UserBadge from './UserBadge.svelte';

	export let roomId: string;

	let messageInput = '';
	let messagesContainer: HTMLDivElement;
	let shouldAutoScroll = true;
	let typingTimeout: ReturnType<typeof setTimeout>;

	$: messages = $chat.messages.get(roomId) || [];
	$: typingUsers = Array.from($chat.typingUsers.get(roomId) || []).filter(
		u => u !== $auth.user?.username
	);

	function handleScroll() {
		if (!messagesContainer) return;
		const { scrollTop, scrollHeight, clientHeight } = messagesContainer;
		shouldAutoScroll = scrollHeight - scrollTop - clientHeight < 100;
	}

	afterUpdate(() => {
		if (shouldAutoScroll && messagesContainer) {
			messagesContainer.scrollTop = messagesContainer.scrollHeight;
		}
	});

	function sendMessage() {
		if (!messageInput.trim()) return;
		chat.sendMessage(roomId, messageInput);
		messageInput = '';
	}

	function handleKeyDown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			sendMessage();
		} else {
			// Send typing indicator (debounced)
			clearTimeout(typingTimeout);
			typingTimeout = setTimeout(() => {
				chat.sendTyping(roomId);
			}, 300);
		}
	}

	function formatTime(timestamp: string): string {
		const date = new Date(timestamp);
		return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
	}

	function formatDate(timestamp: string): string {
		const date = new Date(timestamp);
		const today = new Date();
		if (date.toDateString() === today.toDateString()) return 'Today';
		const yesterday = new Date(today);
		yesterday.setDate(yesterday.getDate() - 1);
		if (date.toDateString() === yesterday.toDateString()) return 'Yesterday';
		return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
	}

	function shouldShowDateDivider(messages: ChatMessage[], index: number): boolean {
		if (index === 0) return true;
		const curr = new Date(messages[index].timestamp).toDateString();
		const prev = new Date(messages[index - 1].timestamp).toDateString();
		return curr !== prev;
	}

	function isConsecutive(messages: ChatMessage[], index: number): boolean {
		if (index === 0) return false;
		const prev = messages[index - 1];
		const curr = messages[index];
		if (prev.sender_id !== curr.sender_id) return false;
		const diff = new Date(curr.timestamp).getTime() - new Date(prev.timestamp).getTime();
		return diff < 120000; // 2 minutes
	}
</script>

<div class="chat-panel">
	<div class="messages" bind:this={messagesContainer} on:scroll={handleScroll}>
		{#if $chat.loading}
			<div class="loading-messages">
				<span class="animate-pulse text-muted">Loading messages...</span>
			</div>
		{:else if messages.length === 0}
			<div class="empty-messages">
				<span class="empty-icon">💬</span>
				<p>No messages yet</p>
				<p class="text-sm text-muted">Be the first to say something!</p>
			</div>
		{:else}
			{#each messages as msg, i (msg.id)}
				{#if shouldShowDateDivider(messages, i)}
					<div class="date-divider">
						<span>{formatDate(msg.timestamp)}</span>
					</div>
				{/if}
				<div class="message" class:consecutive={isConsecutive(messages, i)} class:own={msg.sender_id === $auth.user?.id}>
					{#if !isConsecutive(messages, i)}
						<div class="message-header">
							<UserBadge username={msg.sender_name} size="sm" showName={false} />
							<span class="sender-name">{msg.sender_name}</span>
							<span class="message-time">{formatTime(msg.timestamp)}</span>
						</div>
					{/if}
					<div class="message-body" class:has-avatar={!isConsecutive(messages, i)}>
						{msg.content}
					</div>
				</div>
			{/each}
		{/if}

		{#if typingUsers.length > 0}
			<div class="typing-indicator animate-fadeIn">
				<div class="typing-dots">
					<span></span><span></span><span></span>
				</div>
				<span class="typing-text">
					{typingUsers.join(', ')} {typingUsers.length === 1 ? 'is' : 'are'} typing
				</span>
			</div>
		{/if}
	</div>

	<div class="input-area">
		<div class="input-wrapper">
			<input
				type="text"
				bind:value={messageInput}
				on:keydown={handleKeyDown}
				placeholder="Type a message..."
			/>
			<button class="send-btn" on:click={sendMessage} disabled={!messageInput.trim()}>
				↵
			</button>
		</div>
	</div>
</div>

<style>
	.chat-panel {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: var(--bg-primary);
		border-left: 1px solid var(--border);
		width: 380px;
		flex-shrink: 0;
	}

	.messages {
		flex: 1;
		overflow-y: auto;
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.loading-messages, .empty-messages {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		flex: 1;
		gap: 8px;
	}

	.empty-icon {
		font-size: 2.5rem;
		opacity: 0.5;
	}

	.empty-messages p {
		color: var(--text-secondary);
	}

	.date-divider {
		display: flex;
		align-items: center;
		gap: 12px;
		margin: 16px 0 8px;
	}

	.date-divider::before, .date-divider::after {
		content: '';
		flex: 1;
		height: 1px;
		background: var(--border);
	}

	.date-divider span {
		font-size: 0.7rem;
		font-weight: 600;
		color: var(--text-muted);
		text-transform: uppercase;
	}

	.message {
		padding: 4px 0;
	}

	.message:not(.consecutive) {
		margin-top: 8px;
	}

	.message-header {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 2px;
	}

	.sender-name {
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--text-primary);
	}

	.message-time {
		font-size: 0.7rem;
		color: var(--text-muted);
	}

	.message-body {
		font-size: 0.9rem;
		line-height: 1.45;
		color: var(--text-primary);
		word-wrap: break-word;
	}

	.message-body.has-avatar {
		padding-left: 40px;
	}

	.consecutive .message-body {
		padding-left: 40px;
	}

	.typing-indicator {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 0;
	}

	.typing-text {
		font-size: 0.75rem;
		color: var(--text-muted);
	}

	.input-area {
		padding: 12px 16px;
		border-top: 1px solid var(--border);
	}

	.input-wrapper {
		display: flex;
		align-items: center;
		gap: 8px;
		background: var(--bg-secondary);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: 4px 4px 4px 14px;
	}

	.input-wrapper input {
		flex: 1;
		border: none;
		background: transparent;
		padding: 8px 0;
		font-size: 0.9rem;
	}

	.input-wrapper input:focus {
		box-shadow: none;
	}

	.send-btn {
		width: 36px;
		height: 36px;
		border-radius: var(--radius);
		background: var(--accent);
		color: white;
		font-size: 1.1rem;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.send-btn:hover:not(:disabled) {
		background: var(--accent-hover);
	}

	.send-btn:disabled {
		opacity: 0.3;
		cursor: default;
	}
</style>
