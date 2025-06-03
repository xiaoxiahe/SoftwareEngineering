<script lang="ts">
  import { goto } from '$app/navigation';
  import { register } from '$lib/auth';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { Label } from '$lib/components/ui/label';
  import { Alert, AlertDescription } from '$lib/components/ui/alert';
  import { auth } from '$lib/stores/auth';
  import { onMount } from 'svelte';

  let username = '';
  let password = '';
  let confirmPassword = '';
  let licensePlate = '';
  let batteryCapacity = 60;
  let isLoading = false;
  let errorMessage = '';
  let successMessage = '';

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
    // 验证输入
    if (!username || !password || !confirmPassword || !licensePlate || !batteryCapacity) {
      errorMessage = '请填写所有必填字段';
      return;
    }

    if (password !== confirmPassword) {
      errorMessage = '两次输入的密码不一致';
      return;
    }

    if (batteryCapacity <= 0) {
      errorMessage = '电池容量必须大于零';
      return;
    }

    isLoading = true;
    errorMessage = '';
    successMessage = '';

    try {
      const result = await register(username, password, {
        licensePlate,
        batteryCapacity
      });
      
      if (result.success) {
        successMessage = '注册成功！即将跳转到登录页面...';
        setTimeout(() => {
          goto('/login');
        }, 2000);
      } else {
        errorMessage = result.error || '注册失败，请稍后再试';
      }
    } catch (error) {
      errorMessage = '注册过程中发生错误，请稍后再试';
      console.error('Registration error:', error);
    } finally {
      isLoading = false;
    }
  };
</script>

<div class="flex min-h-svh bg-gradient-to-r from-blue-50 to-blue-100">
  <main class="container mx-auto flex flex-col items-center justify-center gap-6 py-12">
    <div class="w-full max-w-md rounded-lg bg-white p-8 shadow-lg">
      <div class="mb-6 text-center">
        <h1 class="text-2xl font-bold text-blue-700">注册</h1>
        <p class="text-gray-500">创建您的充电系统账户</p>
      </div>

      {#if errorMessage}
        <Alert variant="destructive" class="mb-4">
          <AlertDescription>{errorMessage}</AlertDescription>
        </Alert>
      {/if}

      {#if successMessage}
        <Alert variant="default" class="mb-4 bg-green-50 text-green-700 border-green-200">
          <AlertDescription>{successMessage}</AlertDescription>
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

        <div class="space-y-2">
          <Label for="confirmPassword">确认密码</Label>
          <Input 
            id="confirmPassword" 
            type="password" 
            bind:value={confirmPassword} 
            placeholder="请再次输入密码" 
            disabled={isLoading} 
          />
        </div>

        <div class="space-y-1 pt-2">
          <h3 class="text-sm font-medium">车辆信息</h3>
        </div>

        <div class="space-y-2">
          <Label for="licensePlate">车牌号</Label>
          <Input 
            id="licensePlate" 
            type="text" 
            bind:value={licensePlate} 
            placeholder="请输入车牌号" 
            disabled={isLoading} 
          />
        </div>

        <div class="space-y-2">
          <Label for="batteryCapacity">电池容量（度）</Label>
          <Input 
            id="batteryCapacity" 
            type="number" 
            bind:value={batteryCapacity} 
            min="1"
            step="0.1"
            placeholder="请输入电池容量" 
            disabled={isLoading} 
          />
        </div>

        <Button type="submit" class="w-full" disabled={isLoading}>
          {isLoading ? '注册中...' : '注册'}
        </Button>

        <p class="text-center text-sm">
          已有账号？<a href="/login" class="text-blue-600 hover:underline">立即登录</a>
        </p>
      </form>
    </div>
  </main>
</div>
