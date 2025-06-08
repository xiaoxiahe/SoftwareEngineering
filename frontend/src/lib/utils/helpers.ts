import type { BillingDetail, ChargingPile } from '$lib/types';

// 根据时间获取电价类型
export function getPriceType(time: string): 'peak' | 'normal' | 'valley' {
  const hour = new Date(time).getHours();
  
  if ((hour >= 10 && hour < 15) || (hour >= 18 && hour < 21)) {
    return 'peak';
  } else if ((hour >= 7 && hour < 10) || (hour >= 15 && hour < 18) || (hour >= 21 && hour < 23)) {
    return 'normal';
  } else {
    return 'valley';
  }
}

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

// 根据充电模式获取对应的功率
export function getPowerByMode(mode: 'fast' | 'slow'): number {
  return mode === 'fast' ? 30 : 7;
}

// 计算充电时间（小时）
export function calculateChargingTime(capacity: number, power: number): number {
  return capacity / power;
}

// 计算充电费用
export function calculateChargingFee(
  capacity: number,
  startTime: string
): { chargingFee: number; serviceFee: number; totalFee: number; priceType: 'peak' | 'normal' | 'valley' } {
  const priceType = getPriceType(startTime);
  const electricityPrice = getElectricityPrice(priceType);
  const serviceFeeRate = 0.8;
  
  const chargingFee = capacity * electricityPrice;
  const serviceFee = capacity * serviceFeeRate;
  
  return {
    chargingFee,
    serviceFee,
    totalFee: chargingFee + serviceFee,
    priceType
  };
}

// 转换充电桩ID为字符串数组（如快充桩A、B，慢充桩C、D、E）
export function getPileIds(fastCount: number, slowCount: number): string[] {
  const ids: string[] = [];
  
  // 快充桩：A, B, ...
  for (let i = 0; i < fastCount; i++) {
    ids.push(String.fromCharCode(65 + i)); // A, B, C, ...
  }
  
  // 慢充桩：从快充桩之后的字母开始
  for (let i = 0; i < slowCount; i++) {
    ids.push(String.fromCharCode(65 + fastCount + i));
  }
  
  return ids;
}
