import React, { useMemo } from 'react';
import { Analysis } from '../types';
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  flexRender,
  ColumnDef,
  SortingState,
} from '@tanstack/react-table';
import './AnalysisTable.css';

export const AnalysisTable: React.FC<{ data: Analysis[] }> = ({ data }) => {
  const [sorting, setSorting] = React.useState<SortingState>([]);

  const columns = useMemo<ColumnDef<Analysis>[]>(() => [
    { accessorKey: 'status', header: 'Status' },
    { accessorKey: 'page_title', header: 'Page Title' },
    { accessorKey: 'url', header: 'URL', cell: info => <div className="url-cell">{info.getValue<string>()}</div> },
    { accessorKey: 'html_version', header: 'HTML Version' },
    { accessorKey: 'internal_link_count', header: '# Internal Links' },
    { accessorKey: 'external_link_count', header: '# External Links' },
    { accessorKey: 'created_at', header: 'Analyzed At', cell: info => new Date(info.getValue<string>()).toLocaleString() },
  ], []);

  const table = useReactTable({
    data,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  return (
    <div className="table-container">
      <table>
        <thead>
          {table.getHeaderGroups().map(headerGroup => (
            <tr key={headerGroup.id}>
              {headerGroup.headers.map(header => (
                <th key={header.id} onClick={header.column.getToggleSortingHandler()}>
                  {flexRender(header.column.columnDef.header, header.getContext())}
                  {{ asc: ' ▲', desc: ' ▼' }[header.column.getIsSorted() as string] ?? null}
                </th>
              ))}
            </tr>
          ))}
        </thead>
        <tbody>
          {table.getRowModel().rows.map(row => (
            <tr key={row.id}>
              {row.getVisibleCells().map(cell => (
                <td key={cell.id}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};
