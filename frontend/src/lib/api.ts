import { browser } from '$app/environment';
import type { ApiResponse } from '$lib/types';
import { goto } from '$app/navigation';

const API_BASE_URL = 'http://localhost:8080/api/v1';

export class ApiError extends Error {
  code: number;
  
  constructor(message: string, code: number) {
    super(message);
    this.code = code;
    this.name = 'ApiError';
  }
}

export async function fetchApi<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const token = browser ? localStorage.getItem('token') : null;
  const headers = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers,
  };

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  const data: ApiResponse<T> = await response.json();

  if (data.code !== 200) {
    if (data.code === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      goto('/login');
    }
    throw new ApiError(data.message, data.code);
  }

  return data.data;
}

export const api = {
  // 认证相关
  auth: {
    register: (username: string, password: string, vehicleInfo: { licensePlate: string; batteryCapacity: number }) =>
      fetchApi('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ username, password, vehicleInfo }),
      }),
    
    login: (username: string, password: string) =>
      fetchApi('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ username, password }),
      }),
  },
  
  // 用户相关
  users: {
    getInfo: (userId: string) => fetchApi(`/users/${userId}`),
  },
    // 充电请求相关
  charging: {
    createRequest: (request: { chargingMode: 'fast' | 'slow'; requestedCapacity: number; urgency?: 'normal' | 'urgent' }) =>
      fetchApi('/charging/requests', {
        method: 'POST',
        body: JSON.stringify(request),
      }),
    
    updateRequest: (requestId: string, updates: { chargingMode?: 'fast' | 'slow'; requestedCapacity?: number }) =>
      fetchApi(`/charging/requests/${requestId}`, {
        method: 'PUT',
        body: JSON.stringify(updates),
      }),
    
    cancelRequest: (requestId: string) =>
      fetchApi(`/charging/requests/${requestId}`, {
        method: 'DELETE',
      }),
    
    getRequest: (requestId: string) =>
      fetchApi(`/charging/requests/${requestId}`),
    
    // 获取用户最新的充电请求
    getUserLatestRequest: () =>
      fetchApi(`/charging/requests/latest`),
      
    // 获取用户的所有充电请求
    getUserRequests: (userId: string, params: { status?: 'waiting' | 'charging' | 'completed' | 'cancelled'; page?: number; pageSize?: number } = {}) => {
      const queryParams = new URLSearchParams();
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          queryParams.append(key, value.toString());
        }
      });
      return fetchApi(`/charging/requests/user/${userId}?${queryParams.toString()}`);
    },
  },
  
  // 队列相关
  queue: {
    getStatus: (mode?: 'fast' | 'slow' | 'all') =>
      fetchApi(`/queue/status${mode ? `?mode=${mode}` : ''}`),
    
    getUserPosition: (userId: string) =>
      fetchApi(`/queue/position/${userId}`),
  },
  
  // 充电桩相关
  chargingPiles: {
    getAll: () => fetchApi('/charging-piles'),
    
    getOne: (pileId: string) =>
      fetchApi(`/charging-piles/${pileId}`),
    
    // 管理员功能
    control: (pileId: string, action: 'start' | 'stop' | 'maintenance', reason: string) =>
      fetchApi(`/admin/charging-piles/${pileId}/control`, {
        method: 'POST',
        body: JSON.stringify({ action, reason }),
      }),
    
    getQueueVehicles: (pileId?: string) =>
      fetchApi(`/admin/charging-piles/queue-vehicles${pileId ? `?pileId=${pileId}` : ''}`),
  },
  
  // 计费相关
  billing: {
    getDetails: (params: { startDate?: string; endDate?: string; page?: number; pageSize?: number } = {}) => {
      const queryParams = new URLSearchParams();
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          queryParams.append(key, value.toString());
        }
      });
      return fetchApi(`/billing/details?${queryParams.toString()}`);
    },
    
    getDetail: (detailId: string) =>
      fetchApi(`/billing/details/${detailId}`),
    
    calculateFee: (params: { capacity: number; chargingMode: 'fast' | 'slow'; startTime: string }) =>
      fetchApi('/billing/calculate', {
        method: 'POST',
        body: JSON.stringify(params),
      }),
  },
  
  // 报表相关（管理员）
  reports: {
    getPileReports: (params: { period: 'day' | 'week' | 'month'; startDate: string; endDate: string }) => {
      const queryParams = new URLSearchParams();
      Object.entries(params).forEach(([key, value]) => {
        queryParams.append(key, value.toString());
      });
      return fetchApi(`/admin/reports/charging-piles?${queryParams.toString()}`);
    },
    
    getOperationsReport: (params: { period: 'day' | 'week' | 'month'; date: string }) => {
      const queryParams = new URLSearchParams();
      Object.entries(params).forEach(([key, value]) => {
        queryParams.append(key, value.toString());
      });
      return fetchApi(`/admin/reports/operations?${queryParams.toString()}`);
    },
  },
};
