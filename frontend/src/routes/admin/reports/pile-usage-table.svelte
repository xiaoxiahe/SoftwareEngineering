<script lang="ts">
	import { createSvelteTable, FlexRender } from '$lib/components/ui/data-table/index.js';
	import { getCoreRowModel, getSortedRowModel, type SortingState } from '@tanstack/table-core';
	import * as Table from '$lib/components/ui/table/index.js';
	import ArrowUp from '@lucide/svelte/icons/chevron-up';
	import ArrowDown from '@lucide/svelte/icons/chevron-down';
	import { writable } from 'svelte/store';
	import type { PileUsageStatistics } from './types';
	import { pileUsageColumns } from './pile-usage-columns';

	interface Props {
		data?: PileUsageStatistics[];
	}

	let { data = [] }: Props = $props();
	const sorting = writable<SortingState>([]);

	const table = $derived(
		createSvelteTable({
			get data() {
				return data;
			},
			columns: pileUsageColumns,
			state: {
				get sorting() {
					return $sorting;
				}
			},
			onSortingChange: (updater) => {
				if (typeof updater === 'function') {
					sorting.set(updater($sorting));
				} else {
					sorting.set(updater);
				}
			},
			getCoreRowModel: getCoreRowModel(),
			getSortedRowModel: getSortedRowModel()
		})
	);
</script>

<div class="rounded-md border">
	<Table.Root>
		<Table.Header>
			{#each table.getHeaderGroups() as headerGroup (headerGroup.id)}
				<Table.Row>
					{#each headerGroup.headers as header (header.id)}
						<Table.Head
							class={header.column.getCanSort() ? 'cursor-pointer select-none' : ''}
							onclick={() => {
								if (header.column.getCanSort()) {
									header.column.toggleSorting();
								}
							}}
						>
							{#if !header.isPlaceholder}
								<div class="flex items-center space-x-2">
									<FlexRender
										content={header.column.columnDef.header}
										context={header.getContext()}
									/>
									{#if header.column.getCanSort()}
										<div class="ml-1">
											{#if header.column.getIsSorted() === 'asc'}
												<ArrowUp class="h-4 w-4" />
											{:else if header.column.getIsSorted() === 'desc'}
												<ArrowDown class="h-4 w-4" />
											{:else}
												<div class="h-4 w-4 opacity-30">↕</div>
											{/if}
										</div>
									{/if}
								</div>
							{/if}
						</Table.Head>
					{/each}
				</Table.Row>
			{/each}
		</Table.Header>
		<Table.Body>
			{#each table.getRowModel().rows as row (row.id)}
				<Table.Row data-state={row.getIsSelected() && 'selected'}>
					{#each row.getVisibleCells() as cell (cell.id)}
						<Table.Cell>
							<FlexRender content={cell.column.columnDef.cell} context={cell.getContext()} />
						</Table.Cell>
					{/each}
				</Table.Row>
			{:else}
				<Table.Row>
					<Table.Cell colspan={pileUsageColumns.length} class="h-24 text-center">
						没有数据。
					</Table.Cell>
				</Table.Row>
			{/each}
		</Table.Body>
	</Table.Root>
</div>
