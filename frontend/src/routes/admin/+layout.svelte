<script lang="ts">
	import Navigation from '$lib/components/admin/navigation.svelte';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	// 计算当前激活的导航项
	$: {
		let path = page.url.pathname;
		if (path === '/admin') {
			activeItem = 'dashboard';
		} else if (path.includes('/charging-piles')) {
			activeItem = 'charging-piles';
		} else if (path.includes('/queue')) {
			activeItem = 'queue';
		} else if (path.includes('/waiting-area')) {
			activeItem = 'waiting-area';
		} else if (path.includes('/reports')) {
			activeItem = 'reports';
		} else if (path.includes('/settings')) {
			activeItem = 'settings';
		}
	}

	let activeItem: string;
	// 初始化主题
	onMount(() => {
		const saved = localStorage.getItem('theme');
		if (saved === 'dark') {
			document.documentElement.classList.add('dark');
		} else {
			document.documentElement.classList.remove('dark');
		}
	});

	function toggleDarkMode() {
		const isDark = document.documentElement.classList.toggle('dark');
		localStorage.setItem('theme', isDark ? 'dark' : 'light');
	}
</script>

<div class="flex min-h-svh flex-col">
	<!-- 暗黑模式切换按钮 -->
	<button
		on:click={toggleDarkMode}
		class="bg-secondary text-secondary-foreground hover:bg-secondary/80 absolute top-4 right-50 rounded-md px-1 py-0 text-sm transition"
	>
		🌞/🌙
	</button>
	<Navigation {activeItem} />
	<div class="container mx-auto flex-1 px-4 py-6 md:px-6">
		<slot />
	</div>
</div>
