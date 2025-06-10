<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import type { PileQueueResponse } from '$lib/types';
	import { formatDuration } from '$lib/utils/helpers';
	// shadcn-Svelte components
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';

	let pileQueueData = $state<PileQueueResponse | null>(null);
	let isLoading = $state(true);
	let error = $state<string | null>(null); // 获取队列数据
	async function fetchQueueData() {
		try {
			isLoading = true;
			error = null;
			const response = await api.chargingPiles.getQueueVehicles();
			pileQueueData = response as PileQueueResponse;
		} catch (err: any) {
			error = err.message || '❌ 获取队列数据失败';
			console.error('获取队列数据错误:', err);
		} finally {
			isLoading = false;
		}
	}

	// 组件挂载时获取数据
	onMount(() => {
		fetchQueueData();
	}); // 处理刷新按钮点击
	function handleRefresh() {
		fetchQueueData();
	}

	// 获取充电桩状态对应的样式变体
	function getStatusVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
		switch (status) {
			case 'available':
				return 'secondary';
			case 'occupied':
				return 'default';
			case 'fault':
				return 'destructive';
			case 'maintenance':
				return 'outline';
			default:
				return 'outline';
		}
	}

	// 获取状态文本
	function getStatusText(status: string): string {
		switch (status) {
			case 'available':
				return '空闲';
			case 'occupied':
				return '使用中';
			case 'fault':
				return '故障';
			case 'maintenance':
				return '维护中';
			case 'offline':
				return '离线';
			default:
				return '未知';
		}
	}
</script>

<div class="container mx-auto p-6">
	<div class="flex flex-col gap-6">
		<!-- 标题和操作按钮 -->
		<div class="flex items-center justify-between">
			<h1 class="text-3xl font-bold tracking-tight">🚗 充电桩队列车辆信息</h1>
			<Button onclick={handleRefresh} disabled={isLoading} variant="default">
				{isLoading ? '🔄加载中...' : '🔁刷新数据'}
			</Button>
		</div>
		<!-- 错误信息显示 -->
		{#if error}
			<Alert variant="destructive">
				<AlertDescription>
					错误: {error}
				</AlertDescription>
			</Alert>
		{/if}

		<!-- 加载中 -->
		{#if isLoading}
			<div class="flex items-center justify-center p-12">
				<div
					class="border-primary h-8 w-8 animate-spin rounded-full border-4 border-t-transparent"
				></div>
			</div>
		{:else if !pileQueueData || pileQueueData.piles.length === 0}
			<Card.Root>
				<Card.Content class="p-6">
					<div class="text-muted-foreground text-center">📭 没有找到匹配的充电桩队列数据</div>
				</Card.Content>
			</Card.Root>
		{:else}
			<!-- 充电桩队列信息 -->
			<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
				{#each pileQueueData?.piles as pile (pile.pileId)}
					<Card.Root>
						<Card.Header>
							<div class="flex items-center justify-between">
								<Card.Title class="text-xl">
									充电桩: {pile.pileId}
								</Card.Title>
								<div class="flex items-center gap-2">
									<Badge variant={pile.type === 'fast' ? 'default' : 'secondary'}>
										{pile.type === 'fast' ? '⚡快充' : '🐢慢充'}
									</Badge>
									<Badge variant={getStatusVariant(pile.status)}>
										{#if pile.status === 'available'}✅ 空闲{/if}
										{#if pile.status === 'occupied'}🔄 使用中{/if}
										{#if pile.status === 'fault'}⚠️ 故障{/if}
										{#if pile.status === 'maintenance'}🛠️ 维护中{/if}
										{#if pile.status === 'offline'}🔌 离线{/if}
									</Badge>
								</div>
							</div>
							<Card.Description>
								功率: {pile.power} kW
							</Card.Description>
						</Card.Header>

						<Card.Content>
							<div class="space-y-4">
								<div class="flex items-center justify-between">
									<h4 class="text-sm font-medium">等候车辆</h4>
									<Badge variant="outline">{pile.queueVehicles?.length} 辆</Badge>
								</div>

								{#if !pile.queueVehicles}
									<div
										class="border-muted-foreground/25 rounded-lg border-2 border-dashed p-6 text-center"
									>
										<div class="text-muted-foreground">该充电桩当前没有等候车辆</div>
									</div>
								{:else}
									<div class="rounded-md border">
										<Table.Root>
											<Table.Header>
												<Table.Row>
													<Table.Head class="w-[100px]">排队号</Table.Head>
													<Table.Head>用户ID</Table.Head>
													<Table.Head>电池容量</Table.Head>
													<Table.Head>请求充电量</Table.Head>
													<Table.Head>等待时间</Table.Head>
												</Table.Row>
											</Table.Header>
											<Table.Body>
												{#each pile.queueVehicles as vehicle (vehicle.userId)}
													<Table.Row>
														<Table.Cell class="font-medium">
															{vehicle.queueNumber}
														</Table.Cell>
														<Table.Cell class="font-mono text-sm">
															{vehicle.userId.substring(0, 8)}...
														</Table.Cell>
														<Table.Cell>
															{vehicle.batteryCapacity} kWh
														</Table.Cell>
														<Table.Cell>
															<div class="space-y-1">
																<div>{vehicle.requestedCapacity} kWh</div>
																<div class="text-muted-foreground text-xs">
																	({Math.round(
																		(vehicle.requestedCapacity / vehicle.batteryCapacity) * 100
																	)}%)
																</div>
															</div>
														</Table.Cell>
														<Table.Cell>
															⏱ {formatDuration(vehicle.queueTime)}
														</Table.Cell>
													</Table.Row>
												{/each}
											</Table.Body>
										</Table.Root>
									</div>
								{/if}
							</div>
						</Card.Content>
					</Card.Root>
				{/each}
			</div>
		{/if}
	</div>
</div>
