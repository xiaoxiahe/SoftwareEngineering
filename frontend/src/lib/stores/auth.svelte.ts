import type { User, AuthResponse, ChargingRequestStatus, UserQueuePosition } from '$lib/types';
import { browser } from '$app/environment';

// 用户认证状态接口
interface AuthState {
  token: string | null;
  user: User | null;
  userType: 'user' | 'admin' | null;
  isAuthenticated: boolean;
}

// 初始化状态
function initializeAuthState(): AuthState {
  if (!browser) {
    return {
      token: null,
      user: null,
      userType: null,
      isAuthenticated: false
    };
  }

  const storedToken = localStorage.getItem('token');
  const storedUser = localStorage.getItem('user');
  const storedUserType = localStorage.getItem('userType');

  return {
    token: storedToken,
    user: storedUser ? JSON.parse(storedUser) : null,
    userType: storedUserType as 'user' | 'admin' | null,
    isAuthenticated: !!storedToken
  };
}

// 创建全局状态
class AuthStore {
  private _state = $state<AuthState>(initializeAuthState());

  // 获取当前状态
  get current() {
    return this._state;
  }

  // 获取token
  get token() {
    return this._state.token;
  }

  // 获取用户信息
  get user() {
    return this._state.user;
  }

  // 获取用户类型
  get userType() {
    return this._state.userType;
  }

  // 获取认证状态
  get isAuthenticated() {
    return this._state.isAuthenticated;
  }

  // 是否是管理员
  get isAdmin() {
    return this._state.userType === 'admin';
  }

  // 登录
  login(response: AuthResponse, user: User) {
    const { token, userType } = response;
    
    if (browser) {
      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));
      localStorage.setItem('userType', userType);
    }
    
    this._state = {
      token,
      user,
      userType,
      isAuthenticated: true,
    };
  }

  // 登出
  logout() {
    if (browser) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      localStorage.removeItem('userType');
    }
    
    this._state = {
      token: null,
      user: null,
      userType: null,
      isAuthenticated: false,
    };
  }

  // 更新用户信息
  updateUser(user: User) {
    if (browser) {
      localStorage.setItem('user', JSON.stringify(user));
    }
    
    this._state = {
      ...this._state,
      user,
    };
  }
}

// 充电请求状态类
class ChargingRequestStore {
  private _state = $state<ChargingRequestStatus | null>(null);

  get current() {
    return this._state;
  }

  set(value: ChargingRequestStatus | null) {
    this._state = value;
  }

  clear() {
    this._state = null;
  }
}

// 队列位置状态类
class QueuePositionStore {
  private _state = $state<UserQueuePosition | null>(null);

  get current() {
    return this._state;
  }

  set(value: UserQueuePosition | null) {
    this._state = value;
  }

  clear() {
    this._state = null;
  }
}

// 导出全局store实例
export const auth = new AuthStore();
export const chargingRequest = new ChargingRequestStore();
export const queuePosition = new QueuePositionStore();
