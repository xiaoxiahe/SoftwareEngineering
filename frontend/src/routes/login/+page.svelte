<script lang="ts">
	import { goto } from '$app/navigation';
	import { login } from '$lib/auth';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Alert, AlertDescription } from '$lib/components/ui/alert';
	import { auth } from '$lib/stores/auth.svelte';
	import { onMount } from 'svelte';

	let username = '';
	let password = '';
	let isLoading = false;
	let errorMessage = '';

	onMount(() => {
		if (auth.isAuthenticated) {
			goto(auth.userType === 'admin' ? '/admin' : '/dashboard');
		}
	});

	const handleSubmit = async (event: Event) => {
		event.preventDefault();
		if (!username || !password) {
			errorMessage = '请填写用户名和密码';
			return;
		}

		isLoading = true;
		errorMessage = '';

		try {
			const result = await login(username, password);
			if (!result.success) {
				errorMessage = result.error || '登录失败，请检查用户名和密码';
			}
		} catch (error) {
			errorMessage = '登录过程中发生错误，请稍后再试';
			console.error('Login error:', error);
		} finally {
			isLoading = false;
		}
	};
</script>

<!-- 登录背景 -->
<div
	class="flex min-h-svh items-center justify-center bg-cover bg-center px-4 py-16"
	style="background-image: url('/background.png');"
>
	<main class="w-full max-w-md rounded-2xl bg-white/60 p-8 shadow-xl backdrop-blur-md">
		<div class="mb-6 space-y-1 text-center">
			<div class="text-4xl">🔐</div>
			<h1 class="text-2xl font-extrabold tracking-tight text-blue-700">欢迎登录</h1>
			<p class="text-sm text-gray-600">登录您的充电系统账户</p>
		</div>

		{#if errorMessage}
			<Alert variant="destructive" class="mb-4 border-l-4 border-red-500">
				<AlertDescription class="text-red-600">{errorMessage}</AlertDescription>
			</Alert>
		{/if}

		<form on:submit={handleSubmit} class="space-y-5">
			<div class="space-y-1">
				<Label for="username" class="text-sm font-medium text-gray-700">用户名</Label>
				<Input
					id="username"
					type="text"
					bind:value={username}
					placeholder="请输入用户名"
					disabled={isLoading}
					class="focus:border-blue-500 focus:ring-2 focus:ring-blue-500"
				/>
			</div>

			<div class="space-y-1">
				<Label for="password" class="text-sm font-medium text-gray-700">密码</Label>
				<Input
					id="password"
					type="password"
					bind:value={password}
					placeholder="请输入密码"
					disabled={isLoading}
					class="focus:border-blue-500 focus:ring-2 focus:ring-blue-500"
				/>
			</div>

			<Button
				type="submit"
				class="w-full bg-blue-600 text-white transition hover:bg-blue-700"
				disabled={isLoading}
			>
				{isLoading ? '登录中...' : '登录'}
			</Button>

			<p class="text-center text-sm text-gray-600">
				没有账号？
				<a href="/register" class="font-medium text-blue-600 hover:underline">立即注册</a>
			</p>
		</form>
	</main>
</div>
