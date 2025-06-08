import type {  ChargingPile } from '$lib/types';

// 获取电价
export function getElectricityPrice(priceType: 'peak' | 'normal' | 'valley'): number {
  const prices = {
    peak: 1.0,   // 峰时
    normal: 0.7, // 平时
    valley: 0.4  // 谷时
  };
  return prices[priceType];
}

// 格式化日期时间
export function formatDateTime(dateTime: string): string {
  const date = new Date(dateTime);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
}

// 格式化持续时间为小时、分钟和秒
export function formatDuration(duration: number, unit: 'seconds' | 'hours' = 'seconds'): string {
    // 根据单位转换为秒
    const totalSeconds = unit === 'hours' ? duration * 3600 : duration;
    
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = Math.floor(totalSeconds % 60);
    
    if (hours === 0) {
      if (minutes === 0) {
        return `${seconds}秒`;
      }
      return `${minutes}分钟${seconds > 0 ? ` ${seconds}秒` : ''}`;
    }
    
    return `${hours}小时${minutes > 0 ? ` ${minutes}分钟` : ''}`;
}

// 格式化金额
export function formatCurrency(amount: number): string {
  return amount.toFixed(2) + '元';
}

// 根据充电桩状态获取状态文本和颜色
export function getPileStatusInfo(status: ChargingPile['status']): { text: string; color: string } {
  switch (status) {
    case 'available':
      return { text: '空闲中', color: 'text-green-500' };
    case 'occupied':
      return { text: '使用中', color: 'text-blue-500' };
    case 'fault':
      return { text: '故障', color: 'text-red-500' };
    case 'maintenance':
      return { text: '维护中', color: 'text-orange-500' };
    case 'offline':
      return { text: '离线', color: 'text-gray-500' };
    default:
      return { text: '未知', color: 'text-gray-500' };
  }
}
