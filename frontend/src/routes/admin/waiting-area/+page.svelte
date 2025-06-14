<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import type { WaitingVehiclesResponse } from '$lib/types';
	import { formatDateTime, formatDuration } from '$lib/utils/helpers';

	// shadcn-Svelte components
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { RefreshCw, Car, Clock, Zap, AlertCircle } from 'lucide-svelte';

	// Svelte 5 runes
	let waitingData = $state<WaitingVehiclesResponse | null>(null);
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	let lastRefreshTime = $state<Date | null>(null);

	// 获取等待区数据
	async function fetchWaitingVehicles() {
		try {
			isLoading = true;
			error = null;
			const response = await api.queue.getWaitingVehicles();
			waitingData = response as WaitingVehiclesResponse;
			lastRefreshTime = new Date();
		} catch (err: any) {
			error = err.message || '❌ 获取等待区数据失败';
			console.error('获取等待区数据错误:', err);
		} finally {
			isLoading = false;
		}
	}

	// 组件挂载时获取数据
	onMount(() => {
		fetchWaitingVehicles();
	});

	// 处理手动刷新
	function handleRefresh() {
		fetchWaitingVehicles();
	}

	// 获取充电类型对应的样式变体
	function getRequestTypeVariant(
		requestType: string
	): 'default' | 'secondary' | 'destructive' | 'outline' {
		switch (requestType) {
			case '快充':
				return 'default';
			case '慢充':
				return 'secondary';
			default:
				return 'outline';
		}
	}

	// 计算等待时间
	function calculateWaitTime(createdAt: string): string {
		const created = new Date(createdAt);
		const now = new Date();
		const diffMs = now.getTime() - created.getTime();
		const diffMinutes = Math.floor(diffMs / (1000 * 60));

		if (diffMinutes < 60) {
			return `${diffMinutes} 分钟`;
		} else {
			const hours = Math.floor(diffMinutes / 60);
			const minutes = diffMinutes % 60;
			return `${hours} 小时 ${minutes} 分钟`;
		}
	}

	// 格式化队列号
	function formatQueueNumber(queueNumber: number): string {
		return queueNumber.toString().padStart(3, '0');
	}
</script>

<svelte:head>
	<title>等待区管理 - 充电桩管理系统</title>
</svelte:head>

<div class="space-y-6">
	<!-- 页面标题和操作栏 -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-3xl font-bold tracking-tight">等待区管理</h1>
			<p class="text-muted-foreground mt-2">实时查看和管理充电站等待区中的车辆信息</p>
		</div>

		<div class="flex items-center gap-4">
			{#if lastRefreshTime}
				<span class="text-muted-foreground text-sm">
					最后更新: {formatDateTime(lastRefreshTime.toISOString())}
				</span>
			{/if}

			<Button onclick={handleRefresh} disabled={isLoading} size="sm" variant="outline">
				<RefreshCw class="h-4 w-4 {isLoading ? 'animate-spin' : ''}" />
				刷新
			</Button>
		</div>
	</div>

	<!-- 错误提示 -->
	{#if error}
		<Alert variant="destructive">
			<AlertCircle class="h-4 w-4" />
			<AlertDescription>{error}</AlertDescription>
		</Alert>
	{/if}

	<!-- 统计卡片 -->
	{#if waitingData}
		<div class="grid gap-4 md:grid-cols-3">
			<Card.Root>
				<Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
					<Card.Title class="text-sm font-medium">总等待车辆</Card.Title>
					<Car class="text-muted-foreground h-4 w-4" />
				</Card.Header>
				<Card.Content>
					<div class="text-2xl font-bold">{waitingData.totalCount}</div>
					<p class="text-muted-foreground text-xs">当前在等待区的车辆总数</p>
				</Card.Content>
			</Card.Root>

			<Card.Root>
				<Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
					<Card.Title class="text-sm font-medium">快充等待</Card.Title>
					<Zap class="text-muted-foreground h-4 w-4" />
				</Card.Header>
				<Card.Content>
					<div class="text-2xl font-bold">{waitingData.fastCount}</div>
					<p class="text-muted-foreground text-xs">等待快充的车辆数量</p>
				</Card.Content>
			</Card.Root>

			<Card.Root>
				<Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
					<Card.Title class="text-sm font-medium">慢充等待</Card.Title>
					<Clock class="text-muted-foreground h-4 w-4" />
				</Card.Header>
				<Card.Content>
					<div class="text-2xl font-bold">{waitingData.slowCount}</div>
					<p class="text-muted-foreground text-xs">等待慢充的车辆数量</p>
				</Card.Content>
			</Card.Root>
		</div>
	{:else if isLoading}
		<div class="grid gap-4 md:grid-cols-3">
			{#each Array(3) as _}
				<Card.Root>
					<Card.Header>
						<Skeleton class="h-4 w-24" />
					</Card.Header>
					<Card.Content>
						<Skeleton class="mb-2 h-8 w-16" />
						<Skeleton class="h-3 w-32" />
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	{/if}

	<!-- 等待车辆列表 -->
	<Card.Root>
		<Card.Header>
			<Card.Title class="flex items-center gap-2">
				<Car class="h-5 w-5" />
				等待车辆列表
			</Card.Title>
			<Card.Description>显示当前在等待区的所有车辆，按到达时间排序</Card.Description>
		</Card.Header>
		<Card.Content>
			{#if isLoading}
				<div class="space-y-4">
					{#each Array(5) as _}
						<div class="flex items-center space-x-4">
							<Skeleton class="h-4 w-16" />
							<Skeleton class="h-4 w-24" />
							<Skeleton class="h-4 w-20" />
							<Skeleton class="h-4 w-16" />
							<Skeleton class="h-4 w-20" />
						</div>
					{/each}
				</div>
			{:else if waitingData && waitingData.waitingVehicles.length > 0}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>队列号</Table.Head>
							<Table.Head>车牌号</Table.Head>
							<Table.Head>充电类型</Table.Head>
							<Table.Head>请求容量</Table.Head>
							<Table.Head>等待时间</Table.Head>
							<Table.Head>到达时间</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each waitingData.waitingVehicles as vehicle}
							<Table.Row>
								<Table.Cell class="font-mono">
									#{formatQueueNumber(vehicle.queueNumber)}
								</Table.Cell>
								<Table.Cell class="font-semibold">
									{vehicle.licensePlate}
								</Table.Cell>
								<Table.Cell>
									<Badge variant={getRequestTypeVariant(vehicle.requestType)}>
										{vehicle.requestType}
									</Badge>
								</Table.Cell>
								<Table.Cell>
									{vehicle.requestedCapacity} kWh
								</Table.Cell>
								<Table.Cell class="text-muted-foreground">
									{calculateWaitTime(vehicle.createdAt)}
								</Table.Cell>
								<Table.Cell class="text-muted-foreground">
									{formatDateTime(vehicle.createdAt)}
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{:else}
				<div class="py-8 text-center">
					<Car class="text-muted-foreground mx-auto mb-4 h-12 w-12" />
					<h3 class="text-lg font-semibold">暂无等待车辆</h3>
					<p class="text-muted-foreground">当前等待区没有车辆等待充电</p>
				</div>
			{/if}
		</Card.Content>
	</Card.Root>
</div>
