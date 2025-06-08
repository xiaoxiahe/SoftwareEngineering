<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { auth } from '$lib/stores/auth';
	import { chargingRequest, queuePosition } from '$lib/stores/auth';
	import { formatDateTime } from '$lib/utils/helpers';
	import {
		Card,
		CardContent,
		CardDescription,
		CardFooter,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Progress } from '$lib/components/ui/progress';
	import { Button } from '$lib/components/ui/button';
	import {
		AlertDialog,
		AlertDialogAction,
		AlertDialogCancel,
		AlertDialogContent,
		AlertDialogDescription,
		AlertDialogFooter,
		AlertDialogHeader,
		AlertDialogTitle,
		AlertDialogTrigger
	} from '$lib/components/ui/alert-dialog';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import type { ChargingRequestStatus, UserQueuePosition } from '$lib/types';
	let isLoading = $state(true);
	let error = $state('');
	let activeRequest = $state<ChargingRequestStatus | null>(null);
	let userPosition = $state<UserQueuePosition | null>(null);
	let showCancelDialog = $state(false);
	let isRefreshing = $state(false);

	// 使用 $derived 来计算是否显示排队信息
	const showQueueInfo = $derived(
		activeRequest &&
			(activeRequest.status === 'waiting' || activeRequest.status === 'queued') &&
			userPosition
	);

	// 使用 $derived 来计算当前时段
	const currentTimePeriod = $derived(() => {
		const hour = new Date().getHours();
		if ((hour >= 10 && hour < 15) || (hour >= 18 && hour < 21)) {
			return '峰时';
		} else if ((hour >= 7 && hour < 10) || (hour >= 15 && hour < 18) || (hour >= 21 && hour < 23)) {
			return '平时';
		} else {
			return '谷时';
		}
	}); // 加载用户请求和队列位置
	async function loadUserStatus() {
		if (!$auth.user) return;

		isLoading = true;
		isRefreshing = true;
		error = '';

		try {
			// 获取用户最新的充电请求
			const latestRequest = (await api.charging.getUserLatestRequest()) as ChargingRequestStatus;

			if (latestRequest) {
				activeRequest = latestRequest;
				$chargingRequest = latestRequest;

				// 如果在等待中或排队中，获取队列位置
				if (latestRequest.status === 'waiting' || latestRequest.status === 'queued') {
					const position = (await api.queue.getUserPosition(
						$auth.user.userId
					)) as UserQueuePosition;
					userPosition = position;
					$queuePosition = position;
				} else {
					userPosition = null;
					$queuePosition = null;
				}
			} else {
				activeRequest = null;
				userPosition = null;
				$chargingRequest = null;
				$queuePosition = null;
			}
		} catch (err) {
			// 这是正常情况，用户没有活动请求
			activeRequest = null;
			userPosition = null;
			$chargingRequest = null;
			$queuePosition = null;
			error = ''; // 不显示错误信息
		} finally {
			isLoading = false;
			isRefreshing = false;
		}
	}
	// 计算充电进度（百分比）
	function calculateProgress(request: ChargingRequestStatus): number {
		if (request.status !== 'charging' || !request.actualCapacity || !request.requestedCapacity) {
			return 0;
		}

		return Math.min(100, (request.actualCapacity / request.requestedCapacity) * 100);
	}

	// 获取状态显示文本
	function getStatusText(status: string): string {
		switch (status) {
			case 'charging':
				return '充电中';
			case 'waiting':
				return '等待中';
			case 'queued':
				return '排队中';
			default:
				return '未知状态';
		}
	}

	// 获取状态颜色类
	function getStatusColor(status: string): string {
		switch (status) {
			case 'charging':
				return 'text-green-500';
			case 'waiting':
				return 'text-amber-500';
			case 'queued':
				return 'text-blue-500';
			default:
				return 'text-gray-500';
		}
	}
	// 显示取消确认对话框
	function showCancelConfirmation() {
		showCancelDialog = true;
	}
	// 取消充电请求
	async function cancelRequest() {
		if (!activeRequest) return;

		try {
			await api.charging.cancelRequest(activeRequest.requestId);
			// 关闭对话框
			showCancelDialog = false;
			// 显示成功提示
			toast.success('充电请求已成功取消');

			// 立即清空当前状态，提供即时反馈
			activeRequest = null;
			userPosition = null;
			$chargingRequest = null;
			$queuePosition = null;

			// 延迟一小段时间后重新加载最新状态，确保后端已更新
			setTimeout(() => {
				loadUserStatus();
			}, 500);
		} catch (err) {
			console.error('Failed to cancel request:', err);
			error = '取消请求失败，请稍后再试';
			toast.error('取消请求失败，请稍后再试');
		}
	}
	onMount(() => {
		loadUserStatus();

		// 定期刷新状态 - 缩短间隔以提供更好的用户体验
		const interval = setInterval(loadUserStatus, 5000);

		// 监听页面可见性变化，当用户返回页面时刷新状态
		const handleVisibilityChange = () => {
			if (!document.hidden) {
				loadUserStatus();
			}
		};

		document.addEventListener('visibilitychange', handleVisibilityChange);

		return () => {
			clearInterval(interval);
			document.removeEventListener('visibilitychange', handleVisibilityChange);
		};
	});

	// 使用 $effect 来处理响应式副作用
	$effect(() => {
		// 当用户状态变化时，可以在这里添加额外的逻辑
		if (activeRequest) {
			console.log('Active request updated:', activeRequest.status);
		}
	});
</script>

<svelte:head>
	<title>用户主页 - 智能充电桩调度计费系统</title>
</svelte:head>

<div class="space-y-6">
	<div>
		<h2 class="text-3xl font-bold tracking-tight">欢迎, {$auth.user?.username || '用户'}</h2>
		<p class="text-muted-foreground">查看您的充电状态和系统信息</p>
	</div>

	{#if isLoading}
		<div class="flex h-40 items-center justify-center rounded-md border border-dashed">
			<p class="text-muted-foreground">加载中...</p>
		</div>
	{:else if error}
		<div class="flex h-40 items-center justify-center rounded-md border border-dashed text-red-500">
			<p>{error}</p>
		</div>
	{:else if activeRequest}
		<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-2">
			<Card>
				<CardHeader>
					<CardTitle>充电状态</CardTitle>
					<CardDescription>
						当前请求编号: {activeRequest.queueNumber}
					</CardDescription>
				</CardHeader>
				<CardContent>
					<div class="mb-4 flex items-center justify-between">
						<span class="text-sm font-medium">状态</span>
						<span class="text-sm font-bold {getStatusColor(activeRequest.status)}">
							{getStatusText(activeRequest.status)}
						</span>
					</div>

					{#if activeRequest.status === 'charging'}
						<div class="space-y-2">
							<div class="flex items-center justify-between">
								<span class="text-sm">充电进度</span>
								<span class="text-sm"
									>{activeRequest.actualCapacity?.toFixed(1) ||
										0}/{activeRequest.requestedCapacity?.toFixed(1) || 0} 度</span
								>
							</div>
							<Progress value={calculateProgress(activeRequest)} />

							<div class="mt-4 grid grid-cols-2 gap-2">
								<div>
									<p class="text-muted-foreground text-xs">充电桩</p>
									<p class="font-medium">{activeRequest.chargingPileId || '未分配'}</p>
								</div>								<div>
									<p class="text-muted-foreground text-xs">开始时间</p>
									<p class="font-medium">
										{activeRequest.startTime ? formatDateTime(activeRequest.startTime) : '-'}
									</p>
								</div>
								<div>
									<p class="text-muted-foreground text-xs">充电模式</p>
									<p class="font-medium">
										{activeRequest.queueNumber.startsWith('F') ? '快充' : '慢充'}
									</p>
								</div>
							</div>
						</div>
					{/if}
					{#if showQueueInfo && userPosition}
						<div class="space-y-4">
							<div>
								<p class="text-muted-foreground text-xs">当前排队位置</p>
								<p class="text-2xl font-bold">{userPosition.position}</p>
							</div>

							<div class="grid grid-cols-2 gap-2">
								<div>
									<p class="text-muted-foreground text-xs">前方车辆</p>
									<p class="font-medium">{userPosition.carsAhead} 辆</p>
								</div>
								<div>
									<p class="text-muted-foreground text-xs">充电模式</p>
									<p class="font-medium">
										{activeRequest.queueNumber.startsWith('F') ? '快充' : '慢充'}
									</p>
								</div>
								<div>
									<p class="text-muted-foreground text-xs">请求充电量</p>
									<p class="font-medium">{activeRequest.requestedCapacity?.toFixed(1)} 度</p>
								</div>
							</div>
						</div>
					{/if}
				</CardContent>
				<CardFooter class="flex justify-between">
					{#if activeRequest.status === 'waiting' || activeRequest.status === 'queued' || activeRequest.status === 'charging'}
						<Button variant="destructive" onclick={showCancelConfirmation}>取消充电</Button>
					{:else}
						<Button onclick={() => goto('/dashboard/details')}>查看详单</Button>
					{/if}

					<Button variant="outline" onclick={() => loadUserStatus()} disabled={isRefreshing}>
						{isRefreshing ? '刷新中...' : '刷新'}
					</Button>
				</CardFooter>
			</Card>

			<Card>
				<CardHeader>
					<CardTitle>快捷操作</CardTitle>
					<CardDescription>请求充电或查看信息</CardDescription>
				</CardHeader>
				<CardContent class="flex flex-col gap-2">
					<Button
						variant="default"
						class="w-full"
						onclick={() => goto('/dashboard/charging-request')}
					>
						{activeRequest ? '修改充电请求' : '新充电请求'}
					</Button>
					<Button variant="outline" class="w-full" onclick={() => goto('/dashboard/details')}>
						查看充电详单
					</Button>
				</CardContent>
			</Card>
		</div>
	{:else}
		<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-2">
			<Card>
				<CardHeader>
					<CardTitle>无活动充电请求</CardTitle>
					<CardDescription>您当前没有活动的充电请求</CardDescription>
				</CardHeader>
				<CardContent>
					<p class="text-muted-foreground">提交新的充电请求开始充电</p>
				</CardContent>
				<CardFooter>
					<Button onclick={() => goto('/dashboard/charging-request')}>新充电请求</Button>
				</CardFooter>
			</Card>

			<Card>
				<CardHeader>
					<CardTitle>快捷操作</CardTitle>
					<CardDescription>请求充电或查看信息</CardDescription>
				</CardHeader>
				<CardContent class="flex flex-col gap-2">
					<Button
						variant="default"
						class="w-full"
						onclick={() => goto('/dashboard/charging-request')}
					>
						新充电请求
					</Button>
					<Button variant="outline" class="w-full" onclick={() => goto('/dashboard/details')}>
						查看充电详单
					</Button>
				</CardContent>
			</Card>
		</div>
	{/if}

	<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-2">
		<Card>
			<CardHeader>
				<CardTitle>充电价格</CardTitle>
				<CardDescription>当前时段: {currentTimePeriod()}</CardDescription>
			</CardHeader>
			<CardContent>
				<div class="space-y-2">
					<div class="flex items-center justify-between">
						<span>峰时 (10:00-15:00, 18:00-21:00)</span>
						<span class="font-medium">1.0元/度</span>
					</div>
					<div class="flex items-center justify-between">
						<span>平时 (7:00-10:00, 15:00-18:00, 21:00-23:00)</span>
						<span class="font-medium">0.7元/度</span>
					</div>
					<div class="flex items-center justify-between">
						<span>谷时 (23:00-次日7:00)</span>
						<span class="font-medium">0.4元/度</span>
					</div>
					<div class="mt-2 border-t pt-2">
						<div class="flex items-center justify-between">
							<span>服务费</span>
							<span class="font-medium">0.8元/度</span>
						</div>
					</div>
				</div>
			</CardContent>
		</Card>
	</div>
</div>

<!-- 取消充电确认对话框 -->
<AlertDialog bind:open={showCancelDialog}>
	<AlertDialogContent>
		<AlertDialogHeader>
			<AlertDialogTitle>确认取消充电请求</AlertDialogTitle>
			<AlertDialogDescription>
				您确定要取消当前的充电请求吗？此操作无法撤销。
				{#if activeRequest}
					<div class="bg-muted mt-3 rounded-md p-3">
						<p class="text-sm"><strong>请求编号:</strong> {activeRequest.queueNumber}</p>
						<p class="text-sm"><strong>当前状态:</strong> {getStatusText(activeRequest.status)}</p>
						{#if activeRequest.requestedCapacity}
							<p class="text-sm">
								<strong>请求充电量:</strong>
								{activeRequest.requestedCapacity.toFixed(1)} 度
							</p>
						{/if}
					</div>
				{/if}
			</AlertDialogDescription>
		</AlertDialogHeader>
		<AlertDialogFooter>
			<AlertDialogCancel onclick={() => (showCancelDialog = false)}>保留请求</AlertDialogCancel>
			<AlertDialogAction
				onclick={cancelRequest}
				class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
			>
				确认取消
			</AlertDialogAction>
		</AlertDialogFooter>
	</AlertDialogContent>
</AlertDialog>
