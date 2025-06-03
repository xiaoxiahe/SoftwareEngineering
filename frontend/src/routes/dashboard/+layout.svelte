<script lang="ts">
  import Navigation from '$lib/components/dashboard/navigation.svelte';
  import { page } from '$app/state';

  // 计算当前激活的导航项
  $: {
    let path = page.url.pathname;
    if (path === '/dashboard') {
      activeItem = 'dashboard';
    } else if (path.includes('/charging-request')) {
      activeItem = 'charging-request';
    } else if (path.includes('/queue')) {
      activeItem = 'queue';
    } else if (path.includes('/details')) {
      activeItem = 'details';
    }
  }

  let activeItem: string;
</script>

<div class="flex min-h-svh flex-col">
  <Navigation activeItem={activeItem} />
  <div class="container mx-auto flex-1 px-4 py-6 md:px-6">
    <slot />
  </div>
</div>
