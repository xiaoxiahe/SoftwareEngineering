<script lang="ts">
	import '../app.css';
	import { auth } from '$lib/stores/auth';
	import { goto, afterNavigate } from '$app/navigation';
	import { Toaster } from '$lib/components/ui/sonner';

	let { children } = $props();

	// 路由权限检查
	afterNavigate(({ to }) => {
		const path = to?.url.pathname || '';

		// 如果是管理员页面，但用户不是管理员
		if (path.startsWith('/admin') && $auth.userType !== 'admin') {
			goto('/dashboard');
		}

		// 如果是需要登录的页面，但用户未登录
		if ((path.startsWith('/dashboard') || path.startsWith('/admin')) && !$auth.isAuthenticated) {
			goto('/login');
		}
	});
</script>

{@render children()}
<Toaster />
