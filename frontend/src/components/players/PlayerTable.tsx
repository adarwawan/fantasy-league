import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
  type SortingState,
} from '@tanstack/react-table';
import { useState } from 'react';
import type { Player } from '../../types/player';
import { PlayerRow } from './PlayerRow';
import { PlayerCard } from './PlayerCard';

const ch = createColumnHelper<Player>();

const columns = [
  ch.accessor('name',             { header: 'Player',       enableSorting: true }),
  ch.accessor(r => r.team.short_name, { id: 'team', header: 'Team', enableSorting: true }),
  ch.accessor('position',         { header: 'Pos',          enableSorting: true }),
  ch.accessor('price',            { header: 'Price',        enableSorting: true }),
  ch.accessor('form',             { header: 'Form',         enableSorting: true }),
  ch.accessor('global_ownership', { header: 'Global %',     enableSorting: true }),
  ch.accessor('top_n_ownership',  { header: 'Top-N %',      enableSorting: true }),
  ch.accessor(r => r.top_n_ownership - r.global_ownership, {
    id: 'differential',
    header: 'Diff',
    enableSorting: true,
  }),
  ch.display({ id: 'fixtures',   header: 'Next 5 GWs',    enableSorting: false }),
];

interface Props {
  players: Player[];
  topNSize: number;
}

export function PlayerTable({ players, topNSize }: Props) {
  const [sorting, setSorting] = useState<SortingState>([
    { id: 'global_ownership', desc: true },
  ]);

  const table = useReactTable({
    data: players,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  const sortedRows = table.getRowModel().rows;

  return (
    <>
      {/* Desktop table */}
      <div className="hidden md:block overflow-x-auto rounded-lg border border-slate-700/50 bg-slate-800/50">
        <table className="min-w-full text-left">
          <thead className="bg-slate-700/40 border-b border-slate-700/50">
            {table.getHeaderGroups().map(hg => (
              <tr key={hg.id}>
                {hg.headers.map(header => {
                  const canSort = header.column.getCanSort();
                  const sorted  = header.column.getIsSorted();
                  return (
                    <th
                      key={header.id}
                      className={`px-3 py-2 text-xs font-semibold text-slate-400 uppercase tracking-wide select-none ${canSort ? 'cursor-pointer hover:text-slate-200' : ''}`}
                      onClick={canSort ? header.column.getToggleSortingHandler() : undefined}
                    >
                      {flexRender(header.column.columnDef.header, header.getContext())}
                      {sorted === 'asc' ? ' ↑' : sorted === 'desc' ? ' ↓' : ''}
                    </th>
                  );
                })}
              </tr>
            ))}
          </thead>
          <tbody>
            {sortedRows.map(row => (
              <PlayerRow key={row.id} row={row} />
            ))}
          </tbody>
        </table>
        <div className="px-4 py-2 text-xs text-slate-500 border-t border-slate-700/50">
          {players.length} players · Top-N = {topNSize.toLocaleString()}
        </div>
      </div>

      {/* Mobile card stack */}
      <div className="md:hidden space-y-2">
        {sortedRows.map(row => (
          <PlayerCard key={row.id} player={row.original} />
        ))}
        <p className="text-xs text-slate-500 text-center pt-1">
          {players.length} players · Top-N = {topNSize.toLocaleString()}
        </p>
      </div>
    </>
  );
}
