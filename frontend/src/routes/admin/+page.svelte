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
	import { Tabs, TabsContent, TabsList, TabsTrigger } from '$lib/components/ui/tabs';
	import { Badge } from '$lib/components/ui/badge';
	import { formatCurrency } from '$lib/utils/helpers';
	import { goto } from '$app/navigation';
	import type { ChargingPile, OperationsReport } from '$lib/types';

	let isLoading = true;
	let error = '';
	let chargingPiles: {
		fastChargingPiles: ChargingPile[];
		slowChargingPiles: ChargingPile[];
	} | null = null;
	let operationsReport: OperationsReport | null = null;

	// 加载系统概览数据
	async function loadSystemOverview() {
		isLoading = true;
		error = '';

		try {
			// 获取充电桩状态
			chargingPiles = await api.chargingPiles.getAll();

			// 获取今日运营统计
			const today = new Date().toISOString().split('T')[0];
			operationsReport = await api.reports.getOperationsReport({
				period: 'day',
				date: today
			});
		} catch (err) {
			console.error('Failed to load system overview:', err);
			error = '加载系统概览数据失败，请稍后再试';
		} finally {
			isLoading = false;
		}
	}

	// 获取充电桩状态颜色
	function getPileStatusColor(status: string): string {
		switch (status) {
			case 'available':
				return 'bg-green-100 text-green-800 border-green-200';
			case 'occupied':
				return 'bg-blue-100 text-blue-800 border-blue-200';
			case 'fault':
				return 'bg-red-100 text-red-800 border-red-200';
			case 'maintenance':
				return 'bg-orange-100 text-orange-800 border-orange-200';
			case 'offline':
				return 'bg-gray-200 text-gray-800 border-gray-300';
			default:
				return 'bg-gray-100 text-gray-800 border-gray-200';
		}
	}

	// 获取充电桩状态文本
	function getPileStatusText(status: string): string {
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

	// 控制充电桩
	async function controlChargingPile(pileId: string, action: 'start' | 'stop' | 'maintenance') {
		try {
			await api.chargingPiles.control(pileId, action, `管理员操作: ${action}`);
			await loadSystemOverview(); // 重新加载数据
		} catch (err) {
			console.error(`Failed to ${action} charging pile ${pileId}:`, err);
			error = `操作充电桩失败: ${err.message}`;
		}
	}

	onMount(() => {
		loadSystemOverview();

		// 定期刷新
		const interval = setInterval(loadSystemOverview, 30000); // 每30秒刷新一次

		return () => clearInterval(interval);
	});
</script>

<svelte:head>
	<title>管理控制台 - 智能充电桩调度计费系统</title>
</svelte:head>

<div class="space-y-6">
	<div>
		<h2 class="text-3xl font-bold tracking-tight">⚡ 管理控制台</h2>
		<p class="text-muted-foreground">系统概览和充电桩状态监控</p>
	</div>

	{#if isLoading && !chargingPiles}
		<div class="flex h-40 items-center justify-center rounded-md border border-dashed">
			<p class="text-muted-foreground">正在加载系统数据...</p>
		</div>
	{:else if error}
		<div class="flex h-40 items-center justify-center rounded-md border border-dashed text-red-500">
			<p>{error}</p>
		</div>
	{:else}
		<!-- 系统状态卡片 -->
		<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
			<Card>
				<CardHeader class="flex flex-row items-center justify-between pb-2">
					<CardTitle class="text-sm font-medium">今日充电次数</CardTitle>
					
					<svg
						xmlns="http://www.w3.org/2000/svg"
						width="16"
						height="16"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						class="text-muted-foreground"
						><path d="M14 7h2a2 2 0 0 1 2 2v6a2 2 0 0 1-2 2h-2" /><path
							d="M6 7H4a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2h2"
						/><line x1="22" x2="22" y1="11" y2="13" /><path d="m14 12-4 6" /><path
							d="m10 12 4-6"
						/><line x1="2" x2="2" y1="11" y2="13" /></svg
					>
				</CardHeader>
				<CardContent>
					<div class="text-2xl font-bold">{operationsReport?.chargingSessions || 0}</div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader class="flex flex-row items-center justify-between pb-2">
					<CardTitle class="text-sm font-medium">今日收入</CardTitle>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						width="16"
						height="16"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						class="text-muted-foreground"
						><path d="M2 17a5 5 0 0 0 10 0c0-2.76-2.5-5-5-3-2.5-2-5 .24-5 3Z" /><path
							d="M12 17a5 5 0 0 0 10 0c0-2.76-2.5-5-5-3-2.5-2-5 .24-5 3Z"
						/><path d="M2 7a5 5 0 0 1 10 0c0-2.76-2.5-5-5-3-2.5-2-5 .24-5 3Z" /><path
							d="M12 7a5 5 0 0 1 10 0c0-2.76-2.5-5-5-3-2.5-2-5 .24-5 3Z"
						/></svg
					>
				</CardHeader>
				<CardContent>
					<div class="text-2xl font-bold">
						{formatCurrency(operationsReport?.totalRevenue || 0)}
					</div>
				</CardContent>
			</Card>
		</div>

		<!-- 快捷操作 -->
		<div class="grid gap-4 md:grid-cols-3">
			<Card>
				<CardHeader>
					<CardTitle>快捷操作</CardTitle>
					<CardDescription>管理系统各个功能</CardDescription>
				</CardHeader>
				<CardContent class="grid gap-2">
					<Button
						variant="outline"
						class="justify-start text-left"
						onclick={() => goto('/admin/charging-piles')}
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							width="16"
							height="16"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
							class="mr-2"
							><path d="M14 7h2a2 2 0 0 1 2 2v6a2 2 0 0 1-2 2h-2" /><path
								d="M6 7H4a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2h2"
							/><path d="m14 12-4 6" /><path d="m10 12 4-6" /></svg
						>
						🔌 充电桩详细管理
					</Button>
					<Button
						variant="outline"
						class="justify-start text-left"
						onclick={() => goto('/admin/queue')}
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							width="16"
							height="16"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
							class="mr-2"
							><line x1="10" x2="21" y1="6" y2="6" /><line x1="10" x2="21" y1="12" y2="12" /><line
								x1="10"
								x2="21"
								y1="18"
								y2="18"
							/><path d="M4 6h1v4" /><path d="M4 10h2" /><path
								d="M6 18H4c0-1 2-2 2-3s-1-1.5-2-1"
							/></svg
						>
						⏳ 排队调度管理
					</Button>
					<Button
						variant="outline"
						class="justify-start text-left"
						onclick={() => goto('/admin/reports')}
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							width="16"
							height="16"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
							class="mr-2"
							><path d="M3 3v18h18" /><path d="M18 17V9" /><path d="M13 17V5" /><path
								d="M8 17v-3"
							/></svg
						>
						📊 查看统计报表
					</Button>
				</CardContent>
			</Card>

			<Card class="col-span-2">
				<CardHeader>
					<CardTitle>今日系统状况</CardTitle>
					<CardDescription>{new Date().toLocaleDateString('zh-CN')} 系统运行情况</CardDescription>
				</CardHeader>
				<CardContent>
					<div class="relative h-[200px] w-full">
						<!-- 此处可添加图表组件，例如svelte-chartjs等 -->
						<div class="absolute inset-0 flex items-center justify-center">
							<div class="text-center">
								<p class="text-muted-foreground">
									当前时间: {new Date().toLocaleTimeString('zh-CN')}
								</p>
								<p class="mt-1 text-lg font-medium">系统正常运行中 ✅</p>
								<p class="text-muted-foreground mt-1 text-xs">
									电价类型: {(new Date().getHours() >= 10 && new Date().getHours() < 15) ||
									(new Date().getHours() >= 18 && new Date().getHours() < 21)
										? '峰时'
										: (new Date().getHours() >= 7 && new Date().getHours() < 10) ||
											  (new Date().getHours() >= 15 && new Date().getHours() < 18) ||
											  (new Date().getHours() >= 21 && new Date().getHours() < 23)
											? '平时'
											: '谷时'}
								</p>
							</div>
						</div>
					</div>

					<div class="mt-4 grid grid-cols-2 gap-2 text-center">
						<div class="space-y-1">
							<p class="text-muted-foreground text-xs">快充桩使用率</p>
							<p class="font-medium">
								{chargingPiles?.fastChargingPiles.filter((p) => p.status === 'occupied').length ||
									0} /
								{chargingPiles?.fastChargingPiles.length || 0}
							</p>
						</div>
						<div class="space-y-1">
							<p class="text-muted-foreground text-xs">慢充桩使用率</p>
							<p class="font-medium">
								{chargingPiles?.slowChargingPiles.filter((p) => p.status === 'occupied').length ||
									0} /
								{chargingPiles?.slowChargingPiles.length || 0}
							</p>
						</div>
					</div>
				</CardContent>
			</Card>
		</div>
	{/if}
</div>
