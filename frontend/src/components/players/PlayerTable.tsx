import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
  type Row,
  type SortingState,
} from '@tanstack/react-table';
import { useWindowVirtualizer } from '@tanstack/react-virtual';
import { useRef, useState, type ReactNode } from 'react';
import type { Player } from '../../types/player';
import type { Team } from '../../types/team';
import { useMediaQuery } from '../../hooks/useMediaQuery';
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
  ch.accessor('effective_ownership', { header: 'EO %',      enableSorting: true }),
  ch.accessor(r => r.top_n_ownership - r.global_ownership, {
    id: 'differential',
    header: 'Diff',
    enableSorting: true,
  }),
  ch.display({ id: 'fixtures',   header: 'Next 5 GWs',    enableSorting: false }),
];

// Estimated row/card heights for the virtualizer; measured live via measureElement.
const ROW_HEIGHT = 62;
const CARD_HEIGHT = 104;

// Default number of players shown before the user opts into the full list.
const DEFAULT_LIMIT = 50;

// Sortable fields for the mobile sort control (desktop uses column headers).
const SORT_OPTIONS: { id: string; label: string }[] = [
  { id: 'global_ownership', label: 'Global %' },
  { id: 'top_n_ownership',  label: 'Top-N %'  },
  { id: 'effective_ownership', label: 'EO %'  },
  { id: 'differential',     label: 'Diff'     },
  { id: 'form',             label: 'Form'     },
  { id: 'price',            label: 'Price'    },
  { id: 'name',             label: 'Player'   },
  { id: 'team',             label: 'Team'     },
  { id: 'position',         label: 'Pos'      },
];

// Text fields read more naturally ascending; numeric fields descending.
const ASC_BY_DEFAULT = new Set(['name', 'team', 'position']);

interface Props {
  players: Player[];
  topNSize: number;
  teams?: Team[];
  onPlayerClick?: (p: Player) => void;
}

export function PlayerTable({ players, topNSize, teams, onPlayerClick }: Props) {
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
  const isDesktop = useMediaQuery('(min-width: 768px)');

  const [expanded, setExpanded] = useState(false);
  const total = sortedRows.length;
  const capped = !expanded && total > DEFAULT_LIMIT;
  const visibleRows = capped ? sortedRows.slice(0, DEFAULT_LIMIT) : sortedRows;

  const footer = (
    <div className="flex flex-wrap items-center justify-center gap-x-2 gap-y-1 md:justify-start">
      <span>
        {capped
          ? `Showing top ${DEFAULT_LIMIT} of ${total.toLocaleString()}`
          : `${total.toLocaleString()} players`}
        {' · '}Top-N = {topNSize.toLocaleString()}
      </span>
      {total > DEFAULT_LIMIT && (
        <button
          onClick={() => setExpanded(e => !e)}
          className="text-indigo-400 hover:text-indigo-300 hover:underline"
        >
          {capped ? `Show all ${total.toLocaleString()}` : `Show top ${DEFAULT_LIMIT}`}
        </button>
      )}
    </div>
  );

  if (isDesktop) {
    return (
      <VirtualTable
        table={table}
        rows={visibleRows}
        teams={teams}
        onPlayerClick={onPlayerClick}
        footer={footer}
      />
    );
  }

  return (
    <>
      <MobileSortBar sorting={sorting} setSorting={setSorting} />
      <VirtualCards rows={visibleRows} teams={teams} onPlayerClick={onPlayerClick} footer={footer} />
    </>
  );
}

function MobileSortBar({
  sorting,
  setSorting,
}: {
  sorting: SortingState;
  setSorting: (s: SortingState) => void;
}) {
  const current = sorting[0] ?? { id: 'global_ownership', desc: true };

  function changeField(id: string) {
    setSorting([{ id, desc: !ASC_BY_DEFAULT.has(id) }]);
  }

  function toggleDir() {
    setSorting([{ id: current.id, desc: !current.desc }]);
  }

  return (
    <div className="mb-3 flex items-center gap-2">
      <label htmlFor="mobile-sort" className="text-xs text-slate-400">Sort</label>
      <select
        id="mobile-sort"
        value={current.id}
        onChange={e => changeField(e.target.value)}
        className="flex-1 rounded-md border border-slate-600 bg-slate-700/50 px-2 py-1.5 text-sm text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500"
      >
        {SORT_OPTIONS.map(o => (
          <option key={o.id} value={o.id}>{o.label}</option>
        ))}
      </select>
      <button
        type="button"
        onClick={toggleDir}
        aria-label={current.desc ? 'Sorted descending, tap for ascending' : 'Sorted ascending, tap for descending'}
        className="rounded-md border border-slate-600 bg-slate-700/50 px-3 py-1.5 text-sm text-slate-200 tabular-nums hover:bg-slate-600 focus:outline-none focus:ring-1 focus:ring-indigo-500"
      >
        {current.desc ? '↓' : '↑'}
      </button>
    </div>
  );
}

interface VirtualTableProps {
  table: ReturnType<typeof useReactTable<Player>>;
  rows: Row<Player>[];
  teams?: Team[];
  onPlayerClick?: (p: Player) => void;
  footer: ReactNode;
}

function VirtualTable({ table, rows, teams, onPlayerClick, footer }: VirtualTableProps) {
  const scrollRef = useRef<HTMLDivElement>(null);

  const virtualizer = useWindowVirtualizer({
    count: rows.length,
    estimateSize: () => ROW_HEIGHT,
    overscan: 8,
    scrollMargin: scrollRef.current?.offsetTop ?? 0,
  });

  const virtualRows = virtualizer.getVirtualItems();
  const paddingTop = virtualRows.length ? virtualRows[0].start - virtualizer.options.scrollMargin : 0;
  const paddingBottom = virtualRows.length
    ? virtualizer.getTotalSize() - virtualRows[virtualRows.length - 1].end
    : 0;

  return (
    <div ref={scrollRef} className="overflow-x-auto rounded-lg border border-slate-700/50 bg-slate-800/50">
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
          {paddingTop > 0 && <tr aria-hidden="true"><td style={{ height: paddingTop }} /></tr>}
          {virtualRows.map(vr => {
            const row = rows[vr.index];
            return (
              <PlayerRow
                key={row.id}
                ref={virtualizer.measureElement}
                data-index={vr.index}
                row={row}
                teams={teams}
                onPlayerClick={onPlayerClick}
              />
            );
          })}
          {paddingBottom > 0 && <tr aria-hidden="true"><td style={{ height: paddingBottom }} /></tr>}
        </tbody>
      </table>
      <div className="px-4 py-2 text-xs text-slate-500 border-t border-slate-700/50">{footer}</div>
    </div>
  );
}

function VirtualCards({
  rows,
  teams,
  onPlayerClick,
  footer,
}: {
  rows: Row<Player>[];
  teams?: Team[];
  onPlayerClick?: (p: Player) => void;
  footer: ReactNode;
}) {
  const scrollRef = useRef<HTMLDivElement>(null);

  const virtualizer = useWindowVirtualizer({
    count: rows.length,
    estimateSize: () => CARD_HEIGHT,
    overscan: 6,
    scrollMargin: scrollRef.current?.offsetTop ?? 0,
  });

  return (
    <div>
      <div
        ref={scrollRef}
        className="relative w-full"
        style={{ height: virtualizer.getTotalSize() }}
      >
        {virtualizer.getVirtualItems().map(vr => {
          const row = rows[vr.index];
          return (
            <div
              key={row.id}
              ref={virtualizer.measureElement}
              data-index={vr.index}
              className="absolute left-0 w-full pb-2"
              style={{ transform: `translateY(${vr.start - virtualizer.options.scrollMargin}px)` }}
            >
              <PlayerCard player={row.original} teams={teams} onClick={onPlayerClick} />
            </div>
          );
        })}
      </div>
      <div className="text-xs text-slate-500 text-center pt-1">{footer}</div>
    </div>
  );
}
