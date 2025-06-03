import type { ColumnDef } from "@tanstack/table-core";
import { formatCurrency, formatDuration, formatDateTime } from '$lib/utils/helpers';
import type { BillingDetail } from "$lib/types";

export const billingDetailColumns: ColumnDef<BillingDetail>[] = [
  {
    accessorKey: "detailId",
    header: "详单编号",
    cell: ({ row }) => {
      const detailId = row.getValue("detailId") as string;
      return detailId.slice(0, 8) + "..."; // 显示前8位
    },
  },
  {
    accessorKey: "startTime",
    header: "开始时间",
    cell: ({ row }) => formatDateTime(row.getValue("startTime") as string),
  },
  {
    accessorKey: "pileId",
    header: "充电桩",
    cell: ({ row }) => row.getValue("pileId"),
  },
  {
    accessorKey: "chargingCapacity",
    header: "充电电量",
    cell: ({ row }) => `${(row.getValue("chargingCapacity") as number).toFixed(2)}度`,
  },
  {
    accessorKey: "chargingDuration",
    header: "充电时长",
    cell: ({ row }) => formatDuration(row.getValue("chargingDuration") as number),
  },
  {
    accessorKey: "chargingFee",
    header: "充电费",
    cell: ({ row }) => formatCurrency(row.getValue("chargingFee") as number),
  },
  {
    accessorKey: "serviceFee",
    header: "服务费",
    cell: ({ row }) => formatCurrency(row.getValue("serviceFee") as number),
  },
  {
    accessorKey: "totalFee",
    header: "总费用",
    cell: ({ row }) => formatCurrency(row.getValue("totalFee") as number),
  },
];
