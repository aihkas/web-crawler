import React, { useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
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
  const navigate = useNavigate();
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
        <tbody>
          {table.getRowModel().rows.map(row => (
            <tr key={row.id} onClick={() => navigate(`/analysis/${row.original.id}`)}>
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
