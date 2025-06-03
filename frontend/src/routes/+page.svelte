<script lang="ts">
  import { goto } from '$app/navigation';
  import { auth } from '$lib/stores/auth';
  import { onMount } from 'svelte';

  onMount(() => {
    // 如果已经登录，则重定向到对应页面
    if ($auth.isAuthenticated) {
      if ($auth.userType === 'admin') {
        goto('/admin');
      } else {
        goto('/dashboard');
      }
    }
    else {
        // 如果未登录，重定向到登录页面
        goto('/login');
    }
  });
</script>
