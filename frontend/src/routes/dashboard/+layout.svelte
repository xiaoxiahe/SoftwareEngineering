<script lang="ts"> 
  import Navigation from '$lib/components/dashboard/navigation.svelte';
  import { page } from '$app/state';
  import { onMount } from 'svelte';

  let activeItem: string;

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
    // 初始化主题
  onMount(() => {
    const saved = localStorage.getItem('theme');
    if (saved === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  });

  function toggleDarkMode() {
    const isDark = document.documentElement.classList.toggle('dark');
    localStorage.setItem('theme', isDark ? 'dark' : 'light');
  }

</script>

<div class="flex min-h-screen flex-col bg-background text-foreground transition-colors duration-300">

  <!-- 左侧导航栏 -->
  <Navigation activeItem={activeItem} />
  <!-- 暗黑模式切换按钮 -->
  <button
    on:click={toggleDarkMode}
    class="absolute top-4 right-20 px-1 py-0 text-sm rounded-md bg-secondary text-secondary-foreground hover:bg-secondary/80 transition"
  >
    🌞/🌙
  </button>
  <!-- 主内容区域 -->
  <main class="container mx-auto flex-1 px-4 py-6 md:px-6">
    <div class="rounded-2xl bg-card shadow-md p-6 transition-shadow hover:shadow-lg">
      <slot />
    </div>
  </main>
</div>
