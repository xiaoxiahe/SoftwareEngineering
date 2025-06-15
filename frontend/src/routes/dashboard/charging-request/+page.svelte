<script lang="ts">
	import { api } from '$lib/api';
	import { auth, chargingRequest } from '$lib/stores/auth.svelte';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { RadioGroup, RadioGroupItem } from '$lib/components/ui/radio-group';
	import { Alert, AlertDescription, AlertTitle } from '$lib/components/ui/alert';
	import { Separator } from '$lib/components/ui/separator';
	import {
		Dialog,
		DialogContent,
		DialogDescription,
		DialogFooter,
		DialogHeader,
		DialogTitle
	} from '$lib/components/ui/dialog';
	import { getElectricityPrice } from '$lib/utils/helpers';

	// 表单状态
	let chargingMode = $state<'fast' | 'slow'>('fast');
	let requestedCapacity = $state(20);
	let isLoading = $state(false);
	let isModifying = $state(false);
	let error = $state('');
	let success = $state('');

	// 计费预览
	let estimatedChargingTime = $state(0);
	let estimatedFee = $state({
		chargingFee: 0,
		serviceFee: 0,
		totalFee: 0,
		priceType: 'normal' as 'peak' | 'normal' | 'valley',
		unitPrice: 0
	});
	// 弹窗状态
	let showErrorDialog = $state(false);
	let errorDialogMessage = $state(''); // 检查是否有活动的请求
	$effect(() => {
		(async () => {
			if (chargingRequest.current) {
				// 如果有存储的请求，使用它
				isModifying = chargingRequest.current.status === 'waiting';
				// if (isModifying) {
				// 	// 预填表单
				// 	chargingMode = chargingRequest.current.queueNumber.startsWith('F') ? 'fast' : 'slow';
				// 	requestedCapacity = chargingRequest.current.requestedCapacity || 20;
				// }
			} else {
				// 如果没有存储的请求，尝试从API获取
				try {
					const latestRequest = await api.charging.getUserLatestRequest();
					if (latestRequest && (latestRequest as any).status === 'waiting') {
						chargingRequest.set(latestRequest as any);
						isModifying = true;
						chargingMode = (latestRequest as any).queueNumber?.startsWith('F') ? 'fast' : 'slow';
						requestedCapacity = (latestRequest as any).requestedCapacity || 20;
					}
				} catch (err: any) {
					console.error('Failed to get latest request:', err);
				}
			}
			await updateEstimates();
		})();
	});

	// 使用API计算充电时间和费用
	async function updateEstimates() {
		try {
			const result = await api.billing.calculateFee({
				capacity: requestedCapacity,
				chargingMode,
				startTime: new Date().toISOString()
			});

			estimatedChargingTime = (result as any).chargingDuration;
			estimatedFee = {
				chargingFee: (result as any).chargingFee,
				serviceFee: (result as any).serviceFee,
				totalFee: (result as any).totalFee,
				priceType: (result as any).priceType,
				unitPrice: (result as any).unitPrice
			};
		} catch (err) {
			console.error('Failed to calculate estimates:', err);
			// 如果API调用失败，设置默认值
			estimatedChargingTime = 0;
			estimatedFee = {
				chargingFee: 0,
				serviceFee: 0,
				totalFee: 0,
				priceType: 'normal',
				unitPrice: 0
			};
		}
	}
	// 监听表单值变化并更新充电量限制
	$effect(() => {
		const maxCapacity = Number(auth.user?.vehicleInfo?.batteryCapacity || 60);
		requestedCapacity = Math.max(0.1, Math.min(maxCapacity, requestedCapacity));
	});

	// 当充电模式或充电量变化时更新估算
	$effect(() => {
		// 确保依赖chargingMode和requestedCapacity
		chargingMode;
		requestedCapacity;
		updateEstimates();
	});

	// 提交充电请求
	async function submitRequest() {
		isLoading = true;
		error = '';
		success = '';

		try {
			let result;
			const requestData = {
				chargingMode,
				requestedCapacity
			};
			if (isModifying && chargingRequest.current) {
				// 修改请求
				result = await api.charging.updateRequest(chargingRequest.current.requestId, requestData);
				success = '充电请求修改成功！';
			} else {
				// 创建新请求
				result = await api.charging.createRequest(requestData);
				success = '充电请求提交成功！';
			} // 更新存储的请求
			chargingRequest.set((await api.charging.getRequest((result as any).requestId)) as any);

			// 重置表单
			isModifying = true;
		} catch (err: any) {
			errorDialogMessage =
				'您的充电请求当前不在等候区状态，无法修改。只有在等候区的请求才能进行修改。如需更改，请先取消当前请求，然后重新提交。';
			showErrorDialog = true;
		} finally {
			isLoading = false;
		}
	}
	// 取消修改
	function cancelModification() {
		if (!chargingRequest.current) return;

		isModifying = true;
		chargingMode = chargingRequest.current.queueNumber.startsWith('F') ? 'fast' : 'slow';
		requestedCapacity = chargingRequest.current.requestedCapacity || 20;
		error = '';
		success = '';
	}

	// 开始一个新请求
	function startNewRequest() {
		isModifying = false;
		chargingMode = 'fast';
		requestedCapacity = 20;
		error = '';
		success = '';
	}
</script>

<svelte:head>
	<title>{isModifying ? '修改' : '新建'}⚡ 充电请求 - 智能充电桩调度计费系统</title>
</svelte:head>

<div class="space-y-6">
	<div>
		<h2 class="text-3xl font-bold tracking-tight">{isModifying ? '修改' : '新建'}充电请求</h2>
		<p class="text-muted-foreground">
			{isModifying ? '您可以修改当前在等候区的充电请求' : '提交新的充电请求并加入排队队列'}
		</p>
	</div>

	<div class="grid gap-6 md:grid-cols-2">
		<Card>
			<CardHeader>
				<CardTitle>🔋 充电请求信息</CardTitle>
				<CardDescription>请填写您的充电需求</CardDescription>
			</CardHeader>
			<CardContent>
				{#if error}
					<Alert variant="destructive" class="mb-6">
						<AlertTitle>❌ 错误</AlertTitle>
						<AlertDescription>{error}</AlertDescription>
					</Alert>
				{/if}

				{#if success}
					<Alert variant="default" class="mb-6 border-green-200 bg-green-50 text-green-700">
						<AlertTitle>✅ 成功</AlertTitle>
						<AlertDescription>{success}</AlertDescription>
					</Alert>
				{/if}
				{#if isModifying && chargingRequest.current}
					<div class="mb-6 rounded-md bg-blue-50 p-4 text-sm text-blue-700">
						<p>
							✏️ 您正在修改排队号为 <strong>{chargingRequest.current.queueNumber}</strong> 的充电请求
						</p>
						<p class="mt-2">
							⚠️ 注意：修改充电模式会重新生成排队号，您将排到对应模式队列的最后一位
						</p>
					</div>
				{/if}

				<form onsubmit={submitRequest} class="space-y-4">
					<div class="space-y-2">
						<Label>充电模式</Label>
						<RadioGroup
							value={chargingMode}
							onValueChange={(value) => (chargingMode = value as 'fast' | 'slow')}
						>
							<div class="flex items-center space-x-2">
								<RadioGroupItem value="fast" id="fast" />
								<Label for="fast" class="cursor-pointer">🚀 快充 (30度/小时)</Label>
							</div>
							<div class="flex items-center space-x-2">
								<RadioGroupItem value="slow" id="slow" />
								<Label for="slow" class="cursor-pointer">🐢 慢充 (7度/小时)</Label>
							</div>
						</RadioGroup>
					</div>

					<div class="space-y-2">
						<Label for="capacity">请求充电量 (度)</Label>
						<Input
							id="capacity"
							type="number"
							min="0.1"
							max={auth.user?.vehicleInfo?.batteryCapacity || 60}
							step="0.01"
							bind:value={requestedCapacity}
							disabled={isLoading}
						/>
						{#if auth.user?.vehicleInfo?.batteryCapacity}
							<p class="text-muted-foreground text-xs">
								⚠️ 最大可充: {auth.user.vehicleInfo.batteryCapacity} 度
							</p>
						{/if}
					</div>
					<div class="mt-2 flex justify-between">
						{#if isModifying && chargingRequest.current}
							<Button
								type="button"
								variant="outline"
								onclick={startNewRequest}
								disabled={isLoading}
							>
								新建请求
							</Button>
						{:else}
							<Button
								type="button"
								variant="outline"
								onclick={cancelModification}
								disabled={isLoading || !chargingRequest.current}
							>
								修改请求
							</Button>
						{/if}
						<Button type="submit" disabled={isLoading}>
							{isLoading ? '提交中...' : isModifying ? '保存修改' : '提交请求'}
						</Button>
					</div>
				</form>
			</CardContent>
		</Card>

		<Card>
			<CardHeader>
				<CardTitle>🧾 费用预估</CardTitle>
				<CardDescription>根据您的请求计算预估费用</CardDescription>
			</CardHeader>
			<CardContent>
				<div class="space-y-4">
					<div>
						<p class="text-sm font-medium">🔌 充电信息</p>
						<div class="mt-2 grid grid-cols-2 gap-2">
							<div>
								<p class="text-muted-foreground text-xs">充电模式</p>
								<p class="font-medium">
									{chargingMode === 'fast' ? '快充 (30度/小时)' : '慢充 (7度/小时)'}
								</p>
							</div>
							<div>
								<p class="text-muted-foreground text-xs">⚡ 充电电量</p>
								<p class="font-medium">{requestedCapacity.toFixed(1)} 度</p>
							</div>
							<div>
								<p class="text-muted-foreground text-xs">🕒 估计充电时长</p>
								<p class="font-medium">{estimatedChargingTime.toFixed(1)} 小时</p>
							</div>
							<div>
								<p class="text-muted-foreground text-xs">电价类型</p>
								<p class="font-medium">
									{estimatedFee.priceType === 'peak'
										? '🌞 峰时'
										: estimatedFee.priceType === 'normal'
											? '🌤️ 平时'
											: '🌙 谷时'}
									({getElectricityPrice(estimatedFee.priceType).toFixed(1)}元/度)
								</p>
							</div>
						</div>
					</div>

					<Separator />

					<div>
						<p class="text-sm font-medium">费用明细</p>
						<div class="mt-2 space-y-2">
							<div class="flex items-center justify-between">
								<span>充电费</span>
								<span class="font-medium">{estimatedFee.chargingFee.toFixed(2)}元</span>
							</div>
							<div class="flex items-center justify-between">
								<span>服务费</span>
								<span class="font-medium">{estimatedFee.serviceFee.toFixed(2)}元</span>
							</div>
							<Separator />
							<div class="flex items-center justify-between font-bold">
								<span>💡 总计</span>
								<span>{estimatedFee.totalFee.toFixed(2)}元</span>
							</div>
						</div>
					</div>

					<div class="bg-muted rounded-md p-4 text-sm">
						<p>ℹ️ 注意: 此为预估费用，实际费用将根据充电开始时的电价计算。</p>
						<p class="mt-1">📊 峰谷时段电价不同，请参考系统价格信息。</p>
					</div>
				</div>
			</CardContent>
		</Card>
	</div>
</div>

<!-- 错误提示弹窗 -->
<Dialog bind:open={showErrorDialog}>
	<DialogContent>
		<DialogHeader>
			<DialogTitle>无法修改充电请求</DialogTitle>
			<DialogDescription>请求状态限制</DialogDescription>
		</DialogHeader>

		<div class="py-4">
			<p class="text-sm text-gray-700">
				{errorDialogMessage}
			</p>
		</div>

		<DialogFooter>
			<Button onclick={() => (showErrorDialog = false)}>确定</Button>
		</DialogFooter>
	</DialogContent>
</Dialog>
