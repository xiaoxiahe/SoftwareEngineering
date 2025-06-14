import { goto } from '$app/navigation';
import { browser } from '$app/environment';
import { auth } from '$lib/stores/auth.svelte';
import { api } from '$lib/api';

// 登录逻辑
export const login = async (username: string, password: string) => {
  try {
    const response = await api.auth.login(username, password);
    
    // 临时将token保存到localStorage，以便fetchApi能够获取到token
    if (browser) {
      localStorage.setItem('token', response.token);
    }
    
    const userInfo = await api.users.getInfo(response.userId);
    
    auth.login(response, userInfo);
    // 根据用户类型跳转到不同页面
    if (response.userType === 'admin') {
      goto('/admin');
    } else {
      goto('/dashboard');
    }
    
    return { success: true };
  } catch (error) {
    console.error('Login failed:', error);
    return { success: false, error: error.message };
  }
};

// 注册逻辑
export const register = async (username: string, password: string, vehicleInfo: { licensePlate: string; batteryCapacity: number }) => {
  try {
    await api.auth.register(username, password, vehicleInfo);
    return { success: true };
  } catch (error) {
    console.error('Registration failed:', error);
    return { success: false, error: error.message };
  }
};
