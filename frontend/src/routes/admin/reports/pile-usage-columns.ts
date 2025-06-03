import type { ColumnDef } from "@tanstack/table-core";
import { formatCurrency, formatDuration } from '$lib/utils/helpers';
import type { PileUsageStatistics } from "./types";

export const pileUsageColumns: ColumnDef<PileUsageStatistics>[] = [
  {
    accessorKey: "pileID",
    header: "充电桩编号",
    cell: ({ row }) => row.getValue("pileID"),
  },
  {
    accessorKey: "count",
    header: "充电次数",
    cell: ({ row }) => row.getValue("count"),
  },
  {
    accessorKey: "totalDuration",
    header: "充电时长",
    cell: ({ row }) => formatDuration(row.getValue("totalDuration")),
  },
  {
    accessorKey: "totalCapacity",
    header: "充电电量",
    cell: ({ row }) => `${(row.getValue("totalCapacity") as number).toFixed(1)}度`,
  },
  {
    accessorKey: "totalChargingFee",
    header: "充电费用",
    cell: ({ row }) => formatCurrency(row.getValue("totalChargingFee") as number),
  },
  {
    accessorKey: "totalServiceFee",
    header: "服务费用",
    cell: ({ row }) => formatCurrency(row.getValue("totalServiceFee") as number),
  },
  {
    accessorKey: "totalFee",
    header: "总费用",
    cell: ({ row }) => formatCurrency(row.getValue("totalFee") as number),
  },
];
