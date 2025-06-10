<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Alert, AlertDescription, AlertTitle } from '$lib/components/ui/alert';
	import { Tabs, TabsContent, TabsList, TabsTrigger } from '$lib/components/ui/tabs';
	import { getPileStatusInfo } from '$lib/utils/helpers';
	import type { ChargingPile } from '$lib/types';

	let isLoading = true;
	let error = '';
	let chargingPiles: {
		fastChargingPiles: ChargingPile[];
		slowChargingPiles: ChargingPile[];
	} | null = null;

	// 加载充电桩数据
	async function loadChargingPiles() {
		isLoading = true;
		error = '';

		try {
			chargingPiles = await api.chargingPiles.getAll();
		} catch (err) {
			console.error('Failed to load charging piles:', err);
			error = '加载充电桩数据失败，请稍后再试';
		} finally {
			isLoading = false;
		}
	}

	// 控制充电桩
	async function controlChargingPile(pileId: string, action: 'start' | 'stop' | 'maintenance') {
		try {
			await api.chargingPiles.control(pileId, action, `管理员操作: ${action}`);
			await loadChargingPiles(); // 重新加载数据
		} catch (err) {
			console.error(`Failed to ${action} charging pile ${pileId}:`, err);
			error = `操作充电桩失败: ${err.message}`;
		}
	}

	onMount(() => {
		loadChargingPiles();

		// 定期刷新
		const interval = setInterval(loadChargingPiles, 10000); // 每10秒刷新一次

		return () => clearInterval(interval);
	});
</script>

<svelte:head>
	<title>充电桩管理 - 智能充电桩调度计费系统</title>
</svelte:head>


<div class="space-y-6">
	<div>
		<h2 class="text-3xl font-bold tracking-tight">充电桩管理</h2>
		<p class="text-muted-foreground">管理和监控充电桩状态、队列和故障处理</p>
	</div>

	{#if error}
		<Alert variant="destructive">
			<AlertTitle>错误</AlertTitle>
			<AlertDescription>{error}</AlertDescription>
		</Alert>
	{/if}

	<div class="flex items-center justify-between">
		<h3 class="text-xl font-semibold">充电桩状态监控</h3>
		<Button variant="outline" onclick={loadChargingPiles}>刷新</Button>
	</div>

	<Tabs value="fast">
		<TabsList>
			<TabsTrigger value="fast">快充桩</TabsTrigger>
			<TabsTrigger value="slow">慢充桩</TabsTrigger>
		</TabsList>

		<TabsContent value="fast">
			{#if isLoading}
				<div class="flex h-40 items-center justify-center">
					<p class="text-muted-foreground">加载中...</p>
				</div>
			{:else if chargingPiles && chargingPiles.fastChargingPiles.length > 0}
				<div class="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
					{#each chargingPiles.fastChargingPiles as pile}
						<Card>
							<CardHeader class="pb-2">
								<div class="flex items-center justify-between">
									<CardTitle>{pile.pileId}号充电桩</CardTitle>
									{#if true}
										{@const status = getPileStatusInfo(pile.status)}
										<Badge
											class={pile.status === 'available'
												? 'border-green-200 bg-green-100 text-green-800'
												: pile.status === 'occupied'
													? 'border-blue-200 bg-blue-100 text-blue-800'
													: pile.status === 'fault'
														? 'border-red-200 bg-red-100 text-red-800'
														: 'border-orange-200 bg-orange-100 text-orange-800'}
										>
											{status.text}
										</Badge>
									{/if}
								</div>
								<CardDescription>快充 - {pile.power}度/小时</CardDescription>
							</CardHeader>
							<CardContent>
								<div class="mt-3 flex justify-between">
									{#if pile.status === 'offline'}
										<Button
											variant="default"
											size="sm"
											class="w-full"
											onclick={() => controlChargingPile(pile.pileId, 'start')}
										>
											启动充电桩
										</Button>
									{:else}
										<Button
											variant="destructive"
											size="sm"
											class="w-full"
											disabled={pile.status !== 'available'}
											onclick={() => controlChargingPile(pile.pileId, 'stop')}
										>
											关闭
										</Button>
									{/if}
								</div>
							</CardContent>
						</Card>
					{/each}
				</div>
			{:else}
				<div class="rounded-lg border border-dashed p-8 text-center">
					<p class="text-muted-foreground">没有找到快充电桩信息</p>
				</div>
			{/if}
		</TabsContent>

		<TabsContent value="slow">
			{#if isLoading}
				<div class="flex h-40 items-center justify-center">
					<p class="text-muted-foreground">加载中...</p>
				</div>
			{:else if chargingPiles && chargingPiles.slowChargingPiles.length > 0}
				<div class="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
					{#each chargingPiles.slowChargingPiles as pile}
						<Card>
							<CardHeader class="pb-2">
								<div class="flex items-center justify-between">
									<CardTitle>{pile.pileId}号充电桩</CardTitle>
									{#if true}
										{@const status = getPileStatusInfo(pile.status)}
										<Badge
											class={pile.status === 'available'
												? 'border-green-200 bg-green-100 text-green-800'
												: pile.status === 'occupied'
													? 'border-blue-200 bg-blue-100 text-blue-800'
													: pile.status === 'fault'
														? 'border-red-200 bg-red-100 text-red-800'
														: 'border-orange-200 bg-orange-100 text-orange-800'}
										>
											{status.text}
										</Badge>
									{/if}
								</div>
								<CardDescription>慢充 - {pile.power}度/小时</CardDescription>
							</CardHeader>
							<CardContent>
								<div class="mt-3 flex justify-between">
									{#if pile.status === 'offline'}
										<Button
											variant="default"
											size="sm"
											class="w-full"
											onclick={() => controlChargingPile(pile.pileId, 'start')}
										>
											启动充电桩
										</Button>
									{:else}
										<Button
											variant="destructive"
											size="sm"
											class="w-full"
											disabled={pile.status !== 'available'}
											onclick={() => controlChargingPile(pile.pileId, 'stop')}
										>
											关闭
										</Button>
									{/if}
								</div>
							</CardContent>
						</Card>
					{/each}
				</div>
			{:else}
				<div class="rounded-lg border border-dashed p-8 text-center">
					<p class="text-muted-foreground">没有找到慢充电桩信息</p>
				</div>
			{/if}
		</TabsContent>
	</Tabs>
</div>
