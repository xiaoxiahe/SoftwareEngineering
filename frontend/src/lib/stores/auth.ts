import { writable, derived } from 'svelte/store';
import type { User, AuthResponse, ChargingRequestStatus, UserQueuePosition } from '$lib/types';
import { browser } from '$app/environment';

// 用户认证状态
function createAuthStore() {
  // 从localStorage获取初始状态
  const storedToken = browser ? localStorage.getItem('token') : null;
  const storedUser = browser ? localStorage.getItem('user') : null;
  const storedUserType = browser ? localStorage.getItem('userType') : null;
  
  // 创建store
  const { subscribe, set, update } = writable<{
    token: string | null;
    user: User | null;
    userType: 'user' | 'admin' | null;
    isAuthenticated: boolean;
  }>({
    token: storedToken,
    user: storedUser ? JSON.parse(storedUser) : null,
    userType: (storedUserType as 'user' | 'admin' | null),
    isAuthenticated: !!storedToken,
  });

  return {
    subscribe,
    login: (response: AuthResponse, user: User) => {
      const { token, userType } = response;
      
      if (browser) {
        localStorage.setItem('token', token);
        localStorage.setItem('user', JSON.stringify(user));
        localStorage.setItem('userType', userType);
      }
      
      set({
        token,
        user,
        userType,
        isAuthenticated: true,
      });
    },
    logout: () => {
      if (browser) {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        localStorage.removeItem('userType');
      }
      
      set({
        token: null,
        user: null,
        userType: null,
        isAuthenticated: false,
      });
    },
    updateUser: (user: User) => {
      if (browser) {
        localStorage.setItem('user', JSON.stringify(user));
      }
      
      update((state) => ({
        ...state,
        user,
      }));
    },
  };
}

export const auth = createAuthStore();

// 是否是管理员
export const isAdmin = derived(auth, ($auth) => $auth.userType === 'admin');

// 充电请求状态
function createChargingRequestStore() {
  const { subscribe, set, update } = writable<ChargingRequestStatus | null>(null);

  return {
    subscribe,
    set,
    update,
    clear: () => set(null)
  };
}

export const chargingRequest = createChargingRequestStore();

// 队列位置
function createQueuePositionStore() {
  const { subscribe, set, update } = writable<UserQueuePosition | null>(null);

  return {
    subscribe,
    set,
    update,
    clear: () => set(null)
  };
}

export const queuePosition = createQueuePositionStore();
