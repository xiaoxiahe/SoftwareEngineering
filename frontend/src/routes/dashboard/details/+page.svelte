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
	import { Label } from '$lib/components/ui/label';
	import { Popover, PopoverContent, PopoverTrigger } from '$lib/components/ui/popover';
	import { Calendar } from '$lib/components/ui/calendar';
	import { Pagination } from '$lib/components/ui/pagination';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import type { BillingDetailList } from '$lib/types';
	import BillingDetailTable from './billing-detail-table.svelte';
	import { goto } from '$app/navigation';
	import { CalendarDate, getLocalTimeZone, now } from '@internationalized/date';
	let isLoading = true;
	let error: string | null = null;
	let billingDetails: BillingDetailList | null = null;

	// 获取当前日期和一个月前的日期
	const today = now(getLocalTimeZone());
	const oneMonthAgo = today.subtract({ months: 1 });

	// 过滤参数
	let startDate = new CalendarDate(oneMonthAgo.year, oneMonthAgo.month, oneMonthAgo.day);
	let endDate = new CalendarDate(today.year, today.month, today.day);
	let currentPage = 1;
	let pageSize = 10;
	// 加载详单列表
	async function loadBillingDetails() {
		isLoading = true;
		error = null;

		try {
			// 将 CalendarDate 格式化为 YYYY-MM-DD 字符串
			const formatCalendarDate = (date: CalendarDate) => {
				return `${date.year}-${String(date.month).padStart(2, '0')}-${String(date.day).padStart(2, '0')}`;
			};
			billingDetails = (await api.billing.getDetails({
				startDate: startDate ? formatCalendarDate(startDate) : undefined,
				endDate: endDate ? formatCalendarDate(endDate) : undefined,
				page: currentPage,
				pageSize
			})) as BillingDetailList;
		} catch (err) {
			console.error('Failed to load billing details:', err);
			error = '加载充电详单失败，请稍后再试';
		} finally {
			isLoading = false;
		}
	}

	// 处理页码变化
	function handlePageChange(page: number) {
		currentPage = page;
		loadBillingDetails();
	}
	// 处理过滤条件变化
	function handleFilterChange() {
		currentPage = 1; // 重置为第一页
		loadBillingDetails();
	}
	// 格式化日期为本地字符串
	function formatCalendarDateToLocale(date: CalendarDate) {
		const jsDate = new Date(date.year, date.month - 1, date.day);
		return jsDate.toLocaleDateString();
	}
	// 初始化
	onMount(() => {
		loadBillingDetails();
	});
</script>

<svelte:head>
	<title>充电详单 - 智能充电桩调度计费系统</title>
</svelte:head>

<div class="space-y-6">
	<div>
		<h2 class="text-3xl font-bold tracking-tight">📄 充电详单</h2>
		<p class="text-muted-foreground">🔍 查看您的充电记录和消费情况</p>
	</div>

	<!-- 过滤器 -->
	<Card>
		<CardContent class="pt-6">
			<form onsubmit={handleFilterChange} class="grid gap-4 sm:grid-cols-4">
				<div class="space-y-2">
					<Label for="startDate">开始日期</Label>
					<Popover>
						<PopoverTrigger>
							<Button variant="outline" class="w-full justify-start text-left font-normal">
								{startDate ? formatCalendarDateToLocale(startDate) : '选择开始日期'}
							</Button>
						</PopoverTrigger>
						<PopoverContent class="w-auto p-0">
							<Calendar type="single" bind:value={startDate} />
						</PopoverContent>
					</Popover>
				</div>

				<div class="space-y-2">
					<Label for="endDate">结束日期</Label>
					<Popover>
						<PopoverTrigger>
							<Button variant="outline" class="w-full justify-start text-left font-normal">
								{endDate ? formatCalendarDateToLocale(endDate) : '选择结束日期'}
							</Button>
						</PopoverTrigger>
						<PopoverContent class="w-auto p-0">
							<Calendar type="single" bind:value={endDate} />
						</PopoverContent>
					</Popover>
				</div>

				<div class="flex items-end">
					<Button type="submit" class="w-full">🪄 筛选</Button>
				</div>
				<div class="flex items-end">
					<Button
						type="button"
						variant="outline"
						class="w-full"
						onclick={() => {
							const today = now(getLocalTimeZone());
							const oneMonthAgo = today.subtract({ months: 1 });

							startDate = new CalendarDate(oneMonthAgo.year, oneMonthAgo.month, oneMonthAgo.day);
							endDate = new CalendarDate(today.year, today.month, today.day);
							handleFilterChange();
						}}
					>
						🔁重置
					</Button>
				</div>
			</form>
		</CardContent>
	</Card>

	<!-- 详单列表 -->
	<Card>
		<CardHeader>
			<CardTitle>📋 充电详单记录</CardTitle>
			<CardDescription>
				{startDate && endDate
					? `${formatCalendarDateToLocale(startDate)} - ${formatCalendarDateToLocale(endDate)}`
					: '所有时间'}
				的充电记录
			</CardDescription>
		</CardHeader>
		<CardContent>
			{#if isLoading}
				<div class="space-y-2">
					<Skeleton class="h-8 w-full" />
					<Skeleton class="h-8 w-full" />
					<Skeleton class="h-8 w-full" />
					<Skeleton class="h-8 w-full" />
					<Skeleton class="h-8 w-full" />
				</div>
			{:else if error}
				<div class="flex h-40 items-center justify-center">
					<p class="text-red-500">{error}</p>
				</div>
			{:else if billingDetails && billingDetails.details.length > 0}
				<BillingDetailTable data={billingDetails.details} />

				<!-- 分页 -->
				{#if billingDetails.total > pageSize}
					<div class="mt-4 flex justify-center">
						<Pagination
							count={billingDetails.total}
							perPage={pageSize}
							page={currentPage}
							onPageChange={handlePageChange}
						/>
					</div>
				{/if}
			{:else}
				<div class="flex h-40 items-center justify-center">
					<div class="text-center">
						<p class="text-muted-foreground">📭 没有找到充电详单</p>
						<Button
							variant="outline"
							class="mt-4"
							onclick={() => goto('/dashboard/charging-request')}
						>
							➕ 创建充电请求
						</Button>
					</div>
				</div>
			{/if}
		</CardContent>
	</Card>
</div>
