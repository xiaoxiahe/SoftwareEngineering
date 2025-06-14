// 用户相关类型
export interface User {
  userId: string;
  username: string;
  vehicleInfo?: {
    licensePlate: string;
    batteryCapacity: number;
  };
  createdAt: string;
}

export interface AuthResponse {
  token: string;
  userId: string;
  userType: 'user' | 'admin';
  expiresIn: number;
}

// 充电请求相关类型
export interface ChargingRequest {
  chargingMode: 'fast' | 'slow';
  requestedCapacity: number;
  urgency?: 'normal' | 'urgent';
}

export interface ChargingRequestResponse {
  requestId: string;
  queueNumber: string;
  waitingPosition: number;
}

export interface ChargingRequestStatus {
  requestId: string;
  status: 'waiting' | 'queued' | 'charging' | 'completed' | 'cancelled';
  queueNumber: string;
  chargingPileId?: string;
  waitingPosition?: number;
  createdAt?: string;
  endTime?: string;
  actualCapacity?: number;
  requestedCapacity?: number;
}

// 队列状态相关类型
export interface QueueStatus {
  fastChargingQueue: {
    waiting: QueueItem[];
    availableSlots: number;
  };
  slowChargingQueue: {
    waiting: QueueItem[];
    availableSlots: number;
  };
}

export interface QueueItem {
  queueNumber: string;
  userId: string;
  waitTime: number;
  requestedCapacity: number;
}

export interface UserQueuePosition {
  queueNumber: string;
  position: number;
  carsAhead: number;
}

// 充电桩相关类型
export interface ChargingPile {
  pileId: string;
  type: 'fast' | 'slow';
  status: 'available' | 'occupied' | 'fault' | 'maintenance' | 'offline';
  power: number;
  currentUser?: {
    userId: string;
    queueNumber: string;
    startTime: string;
    requestedCapacity: number;
    currentCapacity: number;
    estimatedEndTime: string;
  };
  queue: ChargingPileQueueItem[];
  statistics?: {
    totalChargingSessions: number;
    totalChargingTime: number;
    totalEnergyDelivered: number;
  };
}

export interface ChargingPileQueueItem {
  position: number;
  userId: string;
  queueNumber: string;
  status?: 'charging' | 'waiting';
  startTime?: string;
  estimatedEndTime?: string;
  estimatedStartTime?: string;
  requestedCapacity: number;
}

// 充电详单相关类型
export interface BillingDetail {
  detailId: string;
  userId?: string;
  generatedAt: string;
  pileId: string;
  queueNumber?: string;
  chargingCapacity: number;
  chargingDuration: number;
  startTime: string;
  endTime: string;
  unitPrice?: number;
  serviceFeeRate?: number;
  chargingFee: number;
  serviceFee: number;
  totalFee: number;
  priceType: 'peak' | 'normal' | 'valley';
}

export interface BillingDetailList {
  total: number;
  page: number;
  pageSize: number;
  details: BillingDetail[];
}

// 报表相关类型
export interface PileReport {
  pileId: string;
  pileType: 'fast' | 'slow';
  totalSessions: number;
  totalDuration: number;
  totalCapacity: number;
  totalChargingFee: number;
  totalServiceFee: number;
  totalRevenue: number;
}

export interface OperationsReport {
  totalUsers: number;
  activeUsers: number;
  totalSessions: number;
  averageWaitTime: number;
  averageChargingTime: number;
  totalRevenue: number;
  peakHourUtilization: number;
  customerSatisfaction: number;
}

// API响应通用格式
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
  timestamp: string;
}

// 管理员查看充电桩队列车辆相关类型
export interface QueueVehicleInfo {
  userId: string;
  batteryCapacity: number;
  requestedCapacity: number;
  currentChargedCapacity: number; // 当前充电量
  currentFee: number; // 当前费用
  queueTime: number;
  queuePosition: number;
  queueNumber: string;
}

export interface PileQueueInfo {
  pileId: string;
  type: 'fast' | 'slow';
  status: 'available' | 'occupied' | 'fault' | 'maintenance' | 'offline';
  power: number;
  queueVehicles: QueueVehicleInfo[];
}

export interface PileQueueResponse {
  piles: PileQueueInfo[];
}
