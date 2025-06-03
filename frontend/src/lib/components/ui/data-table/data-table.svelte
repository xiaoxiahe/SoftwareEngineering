<script lang="ts">
	import { createSvelteTable } from './data-table.svelte.js';
	import FlexRender from './flex-render.svelte';
	import * as Table from '$lib/components/ui/table';
	import type { ColumnDef, TableOptions } from '@tanstack/table-core';
	import type { RowData } from '@tanstack/table-core';

	/**
	 * DataTable 组件接收 columns 和 data 属性
	 * columns: 定义表格的列
	 * data: 表格的数据源
	 */
	let {
		columns,
		data,
		...tableOptions
	}: {
		columns: ColumnDef<any, any>[];
		data: any[];
	} & Partial<TableOptions<any>> = $props();

	// 创建表格实例
	const tableOptionsWithDefaults = {
		data,
		columns,
		...tableOptions
	};
	const table = createSvelteTable(tableOptionsWithDefaults);
</script>

<div data-slot="data-table" class="w-full">
	<Table.Table>
		<Table.Header>
			{#each table.getHeaderGroups() as headerGroup}
				<Table.Row>
					{#each headerGroup.headers as header}
						<Table.Head colspan={header.colSpan}>
							{#if !header.isPlaceholder}
								<FlexRender content={header.column.columnDef.header} context={header.getContext()} />
							{/if}
						</Table.Head>
					{/each}
				</Table.Row>
			{/each}
		</Table.Header>
		<Table.Body>
			{#if table.getRowModel().rows?.length}
				{#each table.getRowModel().rows as row}
					<Table.Row data-state={row.getIsSelected() ? 'selected' : undefined}>
						{#each row.getVisibleCells() as cell}
							<Table.Cell>
								<FlexRender content={cell.column.columnDef.cell} context={cell.getContext()} />
							</Table.Cell>
						{/each}
					</Table.Row>
				{/each}
			{:else}
				<Table.Row>
					<Table.Cell colspan={columns.length} class="text-center h-24">暂无数据</Table.Cell>
				</Table.Row>
			{/if}
		</Table.Body>
	</Table.Table>
</div>
