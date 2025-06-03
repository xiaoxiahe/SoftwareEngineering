<script lang="ts">
	import Navigation from '$lib/components/admin/navigation.svelte';
	import { page } from '$app/state';

	// 计算当前激活的导航项
	$: {
		let path = page.url.pathname;
		if (path === '/admin') {
			activeItem = 'dashboard';
		} else if (path.includes('/charging-piles')) {
			activeItem = 'charging-piles';
		} else if (path.includes('/queue')) {
			activeItem = 'queue';
		} else if (path.includes('/reports')) {
			activeItem = 'reports';
		} else if (path.includes('/settings')) {
			activeItem = 'settings';
		}
	}

	let activeItem: string;
</script>

<div class="flex min-h-svh flex-col">
	<Navigation {activeItem} />
	<div class="container mx-auto flex-1 px-4 py-6 md:px-6">
		<slot />
	</div>
</div>
