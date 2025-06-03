// 充电桩使用统计类型
export interface PileUsageStatistics {
  pileID: string;          // 充电桩ID
  count: number;           // 充电次数
  totalDuration: number;   // 总充电时长（小时）
  totalCapacity: number;   // 总充电电量（度）
  totalChargingFee: number;// 总充电费用（元）
  totalServiceFee: number; // 总服务费用（元）
  totalFee: number;        // 总费用（元）
}
