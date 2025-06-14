<script lang="ts">
	import { auth } from '$lib/stores/auth.svelte';
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
	$: userInitials = auth.user?.username ? auth.user.username.substring(0, 2).toUpperCase() : 'U';

	// 注销处理
	const handleLogout = () => {
		auth.logout();
		console.log('用户已注销');

		goto('/login');
	};

	export let activeItem: string = '';
</script>

<nav
	class="border-b bg-gradient-to-r from-blue-50 via-white to-blue-50 px-4 py-2 shadow-sm md:px-6"
>
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
								href="/dashboard"
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
									class="lucide lucide-home"
									><path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" /><polyline
										points="9 22 9 12 15 12 15 22"
									/></svg
								>
								主页
							</a>

							<a
								href="/dashboard/charging-request"
								class="text-muted-foreground hover:text-primary flex items-center gap-2 rounded-md px-3 py-2 transition-all {activeItem ===
								'charging-request'
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
									class="lucide lucide-zap"
									><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" /></svg
								>
								充电请求
							</a>

							<a
								href="/dashboard/details"
								class="text-muted-foreground hover:text-primary flex items-center gap-2 rounded-md px-3 py-2 transition-all {activeItem ===
								'details'
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
									class="lucide lucide-receipt"
									><path
										d="M4 2v20l2-1 2 1 2-1 2 1 2-1 2 1 2-1 2 1V2l-2 1-2-1-2 1-2-1-2 1-2-1-2 1Z"
									/><path d="M16 8h-6a2 2 0 1 0 0 4h4a2 2 0 1 1 0 4H8" /><path
										d="M12 17.5v-11"
									/></svg
								>
								充电详单
							</a>
						</div>
					</SheetContent>
				</Sheet>
			</div>

			<a href="/dashboard" class="flex items-center gap-2">
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
					class="text-primary h-6 w-6"><path d="M6 7h12M6 12h12M6 17h12" /></svg
				>
				<span class="text-xl font-semibold">充电系统</span>
			</a>

			<div class="ml-10 hidden md:flex md:gap-4">
				<a
					href="/dashboard"
					class="hover:text-primary text-sm font-medium transition-colors {activeItem ===
					'dashboard'
						? 'text-primary'
						: 'text-muted-foreground'}"
				>
					主页
				</a>
				<a
					href="/dashboard/charging-request"
					class="hover:text-primary text-sm font-medium transition-colors {activeItem ===
					'charging-request'
						? 'text-primary'
						: 'text-muted-foreground'}"
				>
					充电请求
				</a>
				<a
					href="/dashboard/details"
					class="hover:text-primary text-sm font-medium transition-colors {activeItem === 'details'
						? 'text-primary'
						: 'text-muted-foreground'}"
				>
					充电详单
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
					{#if auth.user}
						<DropdownMenuLabel>
							{auth.user.username}
						</DropdownMenuLabel>
						<DropdownMenuSeparator />
					{/if}
					<DropdownMenuItem onclick={handleLogout}>退出登录</DropdownMenuItem>
				</DropdownMenuContent>
			</DropdownMenu>
		</div>
	</div>
</nav>
