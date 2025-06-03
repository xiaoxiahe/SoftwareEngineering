<script lang="ts">
  import { goto } from '$app/navigation';
  import { login } from '$lib/auth';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import { Alert, AlertDescription } from '$lib/components/ui/alert';
  import { auth } from '$lib/stores/auth';
  import { onMount } from 'svelte';

  let username = '';
  let password = '';
  let isLoading = false;
  let errorMessage = '';

  onMount(() => {
    // 如果已经登录，则重定向到对应页面
    if ($auth.isAuthenticated) {
      if ($auth.userType === 'admin') {
        goto('/admin');
      } else {
        goto('/dashboard');
      }
    }
  });

  const handleSubmit = async () => {
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

<div class="flex min-h-svh bg-gradient-to-r from-blue-50 to-blue-100">
  <main class="container mx-auto flex flex-col items-center justify-center gap-6 py-16">
    <div class="w-full max-w-md rounded-lg bg-white p-8 shadow-lg">
      <div class="mb-6 text-center">
        <h1 class="text-2xl font-bold text-blue-700">登录</h1>
        <p class="text-gray-500">登录您的充电系统账户</p>
      </div>

      {#if errorMessage}
        <Alert variant="destructive" class="mb-4">
          <AlertDescription>{errorMessage}</AlertDescription>
        </Alert>
      {/if}

      <form onsubmit={handleSubmit} class="space-y-4">
        <div class="space-y-2">
          <Label for="username">用户名</Label>
          <Input id="username" type="text" bind:value={username} placeholder="请输入用户名" disabled={isLoading} />
        </div>

        <div class="space-y-2">
          <Label for="password">密码</Label>
          <Input id="password" type="password" bind:value={password} placeholder="请输入密码" disabled={isLoading} />
        </div>

        <Button type="submit" class="w-full" disabled={isLoading}>
          {isLoading ? '登录中...' : '登录'}
        </Button>

        <p class="text-center text-sm">
          没有账号？<a href="/register" class="text-blue-600 hover:underline">立即注册</a>
        </p>
      </form>
    </div>
  </main>
</div>
