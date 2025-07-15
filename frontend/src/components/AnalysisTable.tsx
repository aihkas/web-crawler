import React, { useMemo, HTMLProps } from 'react';
import { useNavigate } from 'react-router-dom';
import { Analysis } from '../types';
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  flexRender,
  ColumnDef,
  SortingState,
  RowSelectionState,
} from '@tanstack/react-table';
import './AnalysisTable.css';

// Helper component for the header/row checkboxes
function IndeterminateCheckbox({
  indeterminate,
  className = '',
  ...rest
}: { indeterminate?: boolean } & HTMLProps<HTMLInputElement>) {
  const ref = React.useRef<HTMLInputElement>(null!);
  React.useEffect(() => {
    if (typeof indeterminate === 'boolean') {
      ref.current.indeterminate = !rest.checked && indeterminate;
    }
  }, [ref, indeterminate, rest.checked]);
  return <input type="checkbox" ref={ref} className={className + ' cursor-pointer'} {...rest} />;
}

interface AnalysisTableProps {
  data: Analysis[];
  rowSelection: RowSelectionState;
  setRowSelection: React.Dispatch<React.SetStateAction<RowSelectionState>>;
}


export const AnalysisTable: React.FC<AnalysisTableProps> = ({ data, rowSelection, setRowSelection }) => {
  const navigate = useNavigate();
  const [sorting, setSorting] = React.useState<SortingState>([]);

  const columns = useMemo<ColumnDef<Analysis>[]>(() => [
     {
      id: 'select',
      header: ({ table }) => (
        <IndeterminateCheckbox
          {...{
            checked: table.getIsAllRowsSelected(),
            indeterminate: table.getIsSomeRowsSelected(),
            onChange: table.getToggleAllRowsSelectedHandler(),
          }}
        />
      ),
      cell: ({ row }) => (
        <div onClick={(e) => e.stopPropagation()}>
            <IndeterminateCheckbox
            {...{
                checked: row.getIsSelected(),
                disabled: !row.getCanSelect(),
                indeterminate: row.getIsSomeSelected(),
                onChange: row.getToggleSelectedHandler(),
            }}
            />
        </div>
      ),
    },
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
    state: { sorting, rowSelection },
    onSortingChange: setSorting,
    onRowSelectionChange: setRowSelection,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  return (
    <div className="table-container">
      <table>
        <thead>
        </thead>
        <tbody>
          {table.getRowModel().rows.map(row => (
            <tr key={row.id} onClick={() => navigate(`/analysis/${row.original.id}`)}
                className={row.getIsSelected() ? 'selected-row' : ''}>
              {row.getVisibleCells().map(cell => (
                <td key={cell.id} onClick={cell.column.id === 'select' ? (e) => e.stopPropagation() : undefined}>
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
