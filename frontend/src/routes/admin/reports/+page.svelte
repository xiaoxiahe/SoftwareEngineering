<script lang="ts">
	import { api } from '$lib/api';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Calendar } from '$lib/components/ui/calendar/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import PileUsageTable from './pile-usage-table.svelte';
	import {
		type DateValue,
		today,
		getLocalTimeZone,
		parseDate,
		DateFormatter
	} from '@internationalized/date';
	import { cn } from '$lib/utils.js';
	import type { PileUsageStatistics } from './types';

	const dateFormatter = new DateFormatter('zh-CN', {
		dateStyle: 'medium'
	});

	// 使用 Svelte 5 的 runes
	let startDate = $state<DateValue>(
		parseDate(new Date(new Date().setDate(new Date().getDate() - 7)).toISOString().split('T')[0])
	);
	let endDate = $state<DateValue>(today(getLocalTimeZone()));
	let period = $state<'day' | 'week' | 'month'>('day');
	let pileStatistics = $state<PileUsageStatistics[]>([]);
	let isLoading = $state(false);
	let error = $state<string | null>(null);

	// 格式化日期为 YYYY-MM-DD 格式
	function formatDate(date: DateValue): string {
		return date.toString();
	}

	// 获取充电桩报表数据
	async function fetchPileReports() {
		isLoading = true;
		error = null;

		try {
			const params = {
				period,
				startDate: formatDate(startDate),
				endDate: formatDate(endDate)
			};

			pileStatistics = await api.reports.getPileReports(params);
		} catch (err) {
			console.error('获取报表数据失败', err);
			error = err instanceof Error ? err.message : '获取报表数据失败';
		} finally {
			isLoading = false;
		}
	}

	// 页面加载时获取数据
	$effect(() => {
		fetchPileReports();
	});

	// 当日期或周期改变时重新获取数据
	$effect(() => {
		if (startDate && endDate && period) {
			fetchPileReports();
		}
	});
</script>

<div class="container mx-auto space-y-8 py-6">
	<h1 class="text-3xl font-bold">📊 充电桩使用报表</h1>

	<Card.Root class="mb-6">
		<Card.Header>
			<Card.Title>报表筛选条件</Card.Title>
			<Card.Description>选择时间范围和报表周期</Card.Description>
		</Card.Header>
		<Card.Content>
			<div class="flex flex-wrap items-end gap-4">
				<!-- 开始日期选择器 -->
				<div class="space-y-2">
					<label for="startDate" class="text-sm font-medium">开始日期</label>
					<Popover.Root>
						<Popover.Trigger>
							{#snippet child({ props })}
								<Button
									variant="outline"
									class={cn(
										'w-[240px] justify-start text-left font-normal',
										!startDate && 'text-muted-foreground'
									)}
									{...props}
								>
									<CalendarIcon class="mr-2 size-4" />
									{startDate
										? dateFormatter.format(startDate.toDate(getLocalTimeZone()))
										: '选择开始日期'}
								</Button>
							{/snippet}
						</Popover.Trigger>
						<Popover.Content class="w-auto p-0">
							<Calendar bind:value={startDate} type="single" initialFocus />
						</Popover.Content>
					</Popover.Root>
				</div>

				<!-- 结束日期选择器 -->
				<div class="space-y-2">
					<label for="endDate" class="text-sm font-medium">结束日期</label>
					<Popover.Root>
						<Popover.Trigger>
							{#snippet child({ props })}
								<Button
									variant="outline"
									class={cn(
										'w-[240px] justify-start text-left font-normal',
										!endDate && 'text-muted-foreground'
									)}
									{...props}
								>
									<CalendarIcon class="mr-2 size-4" />
									{endDate
										? dateFormatter.format(endDate.toDate(getLocalTimeZone()))
										: '选择结束日期'}
								</Button>
							{/snippet}
						</Popover.Trigger>
						<Popover.Content class="w-auto p-0">
							<Calendar bind:value={endDate} type="single" initialFocus />
						</Popover.Content>
					</Popover.Root>
				</div>

				<!-- 报表周期选择 -->
				<div class="space-y-2">
					<label class="text-sm font-medium">报表周期</label>
					<Tabs.Root
						value={period}
						onValueChange={(value) => (period = value as 'day' | 'week' | 'month')}
					>
						<Tabs.List>
							<Tabs.Trigger value="day">日报</Tabs.Trigger>
							<Tabs.Trigger value="week">周报</Tabs.Trigger>
							<Tabs.Trigger value="month">月报</Tabs.Trigger>
						</Tabs.List>
					</Tabs.Root>
				</div>

				<!-- 刷新按钮 -->
				<Button variant="default" onclick={fetchPileReports} disabled={isLoading}>
					{#if isLoading}
						<span class="animate-spin mr-2">🔄</span> 加载中...
					{:else}
						🔁 刷新数据
					{/if}
				</Button>
			</div>
		</Card.Content>
	</Card.Root>

	<!-- 错误提示 -->
	{#if error}
		<div class="rounded-md border border-red-200 bg-red-50 p-4 text-red-700">
			{error}
		</div>
	{/if}

	<!-- 报表数据表格 -->
	<Card.Root>
		<Card.Header>
			<Card.Title>充电桩使用统计</Card.Title>
			<Card.Description>
				{startDate ? dateFormatter.format(startDate.toDate(getLocalTimeZone())) : ''} 至
				{endDate ? dateFormatter.format(endDate.toDate(getLocalTimeZone())) : ''} 的
				{{ day: '每日', week: '每周', month: '每月' }[period]}报表
			</Card.Description>
		</Card.Header>
		<Card.Content>
		{#if isLoading}
			<div class="flex h-24 items-center justify-center gap-2 text-muted-foreground">
				<span class="animate-spin text-xl">🔄</span>
				正在加载数据...
			</div>
			{:else if pileStatistics.length === 0}
			<div class="text-muted-foreground py-6 text-center text-lg">
				📭 此时间段内没有统计数据
			</div>
			{:else}
				<PileUsageTable data={pileStatistics} />
			{/if}
		</Card.Content>
	</Card.Root>
</div>
