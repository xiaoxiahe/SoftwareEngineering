<script lang="ts">
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { Avatar, AvatarFallback } from '$lib/components/ui/avatar';
	import {
		DropdownMenu,
		DropdownMenuContent,
		DropdownMenuItem,
		DropdownMenuLabel,
		DropdownMenuSeparator,
		DropdownMenuTrigger
	} from '$lib/components/ui/dropdown-menu';
	import { Sheet, SheetContent, SheetTrigger } from '$lib/components/ui/sheet';
	import { Button } from '$lib/components/ui/button';

	// 获取用户名首字母
	$: userInitials = $auth.user?.username ? $auth.user.username.substring(0, 2).toUpperCase() : 'A';

	// 注销处理
	const handleLogout = () => {
		auth.logout();
		console.log('用户已注销');

		goto('/login');
	};

	export let activeItem: string = '';
	import { onMount } from 'svelte';

	// 切换暗黑/明亮模式
	function toggleDarkMode() {
		const html = document.documentElement;
		const isDark = html.classList.toggle('dark');
		localStorage.setItem('theme', isDark ? 'dark' : 'light');
	}

	// 页面加载时应用本地设置
	onMount(() => {
		const savedTheme = localStorage.getItem('theme');
		if (savedTheme === 'dark') {
			document.documentElement.classList.add('dark');
		}
	});
</script>

<nav class="bg-background border-b px-4 py-2 md:px-6">
	<div class="container flex items-center justify-between">
		<div class="flex items-center gap-2">
			<div class="block md:hidden">
				<Sheet>
					<SheetTrigger>
						<Button variant="outline" size="icon" class="md:hidden">
							<svg
								xmlns="http://www.w3.org/2000/svg"
								width="24"
								height="24"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								stroke-linecap="round"
								stroke-linejoin="round"
								class="lucide lucide-menu"
								><line x1="4" x2="20" y1="12" y2="12" /><line x1="4" x2="20" y1="6" y2="6" /><line
									x1="4"
									x2="20"
									y1="18"
									y2="18"
								/></svg
							>
							<span class="sr-only">切换菜单</span>
						</Button>
					</SheetTrigger>
					<SheetContent side="left" class="w-[240px] sm:w-[300px]">
						<div class="grid gap-2 py-6">
							<a
								href="/admin"
								class="text-muted-foreground hover:text-primary flex items-center gap-2 rounded-md px-3 py-2 transition-all {activeItem ===
								'dashboard'
									? 'bg-muted text-primary font-medium'
									: ''}"
							>
								<svg
									xmlns="http://www.w3.org/2000/svg"
									width="20"
									height="20"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									stroke-linecap="round"
									stroke-linejoin="round"
									class="lucide lucide-layout-dashboard"
									><rect width="7" height="9" x="3" y="3" rx="1" /><rect
										width="7"
										height="5"
										x="14"
										y="3"
										rx="1"
									/><rect width="7" height="9" x="14" y="12" rx="1" /><rect
										width="7"
										height="5"
										x="3"
										y="16"
										rx="1"
									/></svg
								>
								系统控制台
							</a>

							<a
								href="/admin/charging-piles"
								class="text-muted-foreground hover:text-primary flex items-center gap-2 rounded-md px-3 py-2 transition-all {activeItem ===
								'charging-piles'
									? 'bg-muted text-primary font-medium'
									: ''}"
							>
								<svg
									xmlns="http://www.w3.org/2000/svg"
									width="20"
									height="20"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									stroke-linecap="round"
									stroke-linejoin="round"
									class="lucide lucide-battery-charging"
									><path d="M14 7h2a2 2 0 0 1 2 2v6a2 2 0 0 1-2 2h-2" /><path
										d="M6 7H4a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2h2"
									/><path d="m14 12-4 6" /><path d="m10 12 4-6" /></svg
								>
								充电桩管理
							</a>
							<a
								href="/admin/queue"
								class="text-muted-foreground hover:text-primary flex items-center gap-2 rounded-md px-3 py-2 transition-all {activeItem ===
								'queue'
									? 'bg-muted text-primary font-medium'
									: ''}"
							>
								<svg
									xmlns="http://www.w3.org/2000/svg"
									width="20"
									height="20"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									stroke-linecap="round"
									stroke-linejoin="round"
									class="lucide lucide-list-ordered"
									><line x1="10" x2="21" y1="6" y2="6" /><line
										x1="10"
										x2="21"
										y1="12"
										y2="12"
									/><line x1="10" x2="21" y1="18" y2="18" /><path d="M4 6h1v4" /><path
										d="M4 10h2"
									/><path d="M6 18H4c0-1 2-2 2-3s-1-1.5-2-1" /></svg
								>
								排队调度
							</a>

							<a
								href="/admin/waiting-area"
								class="text-muted-foreground hover:text-primary flex items-center gap-2 rounded-md px-3 py-2 transition-all {activeItem ===
								'waiting-area'
									? 'bg-muted text-primary font-medium'
									: ''}"
							>
								<svg
									xmlns="http://www.w3.org/2000/svg"
									width="20"
									height="20"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									stroke-linecap="round"
									stroke-linejoin="round"
									class="lucide lucide-car"
									><path d="M8 19h8" /><path d="M6 19V8a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v11" /><path
										d="M10 6V2"
									/><path d="M14 6V2" /></svg
								>
								等待区管理
							</a>

							<a
								href="/admin/reports"
								class="text-muted-foreground hover:text-primary flex items-center gap-2 rounded-md px-3 py-2 transition-all {activeItem ===
								'reports'
									? 'bg-muted text-primary font-medium'
									: ''}"
							>
								<svg
									xmlns="http://www.w3.org/2000/svg"
									width="20"
									height="20"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="2"
									stroke-linecap="round"
									stroke-linejoin="round"
									class="lucide lucide-bar-chart-3"
									><path d="M3 3v18h18" /><path d="M18 17V9" /><path d="M13 17V5" /><path
										d="M8 17v-3"
									/></svg
								>
								统计报表
							</a>
						</div>
					</SheetContent>
				</Sheet>
			</div>

			<a href="/admin" class="flex items-center gap-2">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					width="24"
					height="24"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					class="text-primary h-6 w-6"
					><path d="M17 8h1a4 4 0 1 1 0 8h-1" /><path
						d="M3 8h14v9a4 4 0 0 1-4 4H7a4 4 0 0 1-4-4Z"
					/><line x1="3" x2="21" y1="12" y2="12" /></svg
				>
				<span class="text-xl font-semibold">充电系统管理</span>
			</a>

			<div class="ml-10 hidden md:flex md:gap-4">
				<a
					href="/admin"
					class="hover:text-primary text-sm font-medium transition-colors {activeItem ===
					'dashboard'
						? 'text-primary'
						: 'text-muted-foreground'}"
				>
					控制台
				</a>
				<a
					href="/admin/charging-piles"
					class="hover:text-primary text-sm font-medium transition-colors {activeItem ===
					'charging-piles'
						? 'text-primary'
						: 'text-muted-foreground'}"
				>
					充电桩管理
				</a>
				<a
					href="/admin/queue"
					class="hover:text-primary text-sm font-medium transition-colors {activeItem === 'queue'
						? 'text-primary'
						: 'text-muted-foreground'}"
				>
					排队调度
				</a>
				<a
					href="/admin/waiting-area"
					class="hover:text-primary text-sm font-medium transition-colors {activeItem ===
					'waiting-area'
						? 'text-primary'
						: 'text-muted-foreground'}"
				>
					等待区管理
				</a>
				<a
					href="/admin/reports"
					class="hover:text-primary text-sm font-medium transition-colors {activeItem === 'reports'
						? 'text-primary'
						: 'text-muted-foreground'}"
				>
					统计报表
				</a>
			</div>
		</div>

		<div class="flex items-center gap-2">
			<DropdownMenu>
				<DropdownMenuTrigger>
					<Button variant="ghost" size="icon" class="rounded-full">
						<Avatar>
							<AvatarFallback>{userInitials}</AvatarFallback>
						</Avatar>
						<span class="sr-only">打开用户菜单</span>
					</Button>
				</DropdownMenuTrigger>
				<DropdownMenuContent align="end">
					<DropdownMenuLabel>管理员</DropdownMenuLabel>
					<DropdownMenuLabel>
						{$auth.user?.username}
					</DropdownMenuLabel>
					<DropdownMenuSeparator />
					<DropdownMenuItem onclick={handleLogout}>退出登录</DropdownMenuItem>
				</DropdownMenuContent>
			</DropdownMenu>
		</div>
	</div>
</nav>
