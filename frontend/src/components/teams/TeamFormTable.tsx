import React, { useMemo, useState } from 'react';
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
  type SortingState,
} from '@tanstack/react-table';
import type { Team } from '../../types/team';
import type { Player } from '../../types/player';
import { PositionBadge } from '../common/PositionBadge';
import { FixtureChip, type FocusMode } from '../players/FixtureChip';
import { useMediaQuery } from '../../hooks/useMediaQuery';

const ch = createColumnHelper<Team>();

function makeColumns(focusMode: FocusMode) {
  return [
    ch.display({ id: 'expand',   header: '',      enableSorting: false }),
    ch.accessor('name',          { header: 'Team', enableSorting: true }),
    ch.accessor('att_form',      { header: 'Att',  enableSorting: true }),
    ch.accessor('def_form',      { header: 'Def',  enableSorting: true }),
    ch.accessor('ovr_form',      { header: 'Form', enableSorting: true }),
    ...(focusMode !== 'overall' ? [
      focusMode === 'attack'
        ? ch.accessor('xg_sum', { header: 'xG Sum', enableSorting: true })
        : ch.accessor('cs_avg', { header: 'CS Avg', enableSorting: true }),
    ] : []),
    ch.display({ id: 'fixtures', header: 'Fixtures', enableSorting: false }),
  ];
}

interface Props {
  teams:      Team[];
  players:    Player[];
  focusMode:  FocusMode;
  window:     number;
  currentGw?: number;
  onFocusChange:  (mode: FocusMode) => void;
  onWindowChange: (w: number) => void;
}

function toFixture(tf: Team['fixtures'][number]) {
  return tf as { gw: number; opp: string; ha: 'H' | 'A'; difficulty: 1 | 2 | 3 | 4 | 5; kickoff: string; xg: number | null; cs_pct: number | null };
}

function FormBadge({ value, invert = false }: { value: number; invert?: boolean }) {
  const good = invert ? value < 1.0 : value >= 2.5;
  const mid  = invert ? value < 1.8 : value >= 1.5;
  let color: string;
  let label: string;
  if (good)      { color = 'bg-emerald-400/20 text-emerald-400 ring-emerald-400/30'; label = '▲'; }
  else if (mid)  { color = 'bg-amber-400/20 text-amber-400 ring-amber-400/30';       label = '●'; }
  else           { color = 'bg-red-400/20 text-red-400 ring-red-400/30';              label = '▼'; }
  return (
    <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-xs font-semibold ring-1 tabular-nums ${color}`}>
      {label} {value.toFixed(1)}
    </span>
  );
}

const posOrder: Record<string, number> = { GK: 0, DEF: 1, MID: 2, FWD: 3 };

function inFormPlayers(players: Player[], teamId: string): Player[] {
  return players
    .filter(p => p.team.id === teamId && p.form > 1.5)
    .sort((a, b) => posOrder[a.position] - posOrder[b.position] || b.global_ownership - a.global_ownership);
}

const FOCUS_OPTIONS: { label: string; value: FocusMode }[] = [
  { label: '⚔ Attack',  value: 'attack'  },
  { label: '🛡 Defense', value: 'defense' },
  { label: 'Overall',   value: 'overall' },
];

const WINDOW_OPTIONS = [3, 5, 8];

// Sortable fields for the mobile sort control (desktop uses column headers).
function sortOptions(focusMode: FocusMode): { id: string; label: string }[] {
  return [
    { id: 'ovr_form', label: 'Form' },
    { id: 'att_form', label: 'Att'  },
    { id: 'def_form', label: 'Def'  },
    ...(focusMode === 'attack'  ? [{ id: 'xg_sum', label: 'xG Sum' }] : []),
    ...(focusMode === 'defense' ? [{ id: 'cs_avg', label: 'CS Avg' }] : []),
    { id: 'name', label: 'Team' },
  ];
}

function defaultSort(focusMode: FocusMode): SortingState {
  if (focusMode === 'attack')  return [{ id: 'xg_sum',   desc: true }];
  if (focusMode === 'defense') return [{ id: 'cs_avg',   desc: true }];
  return [{ id: 'ovr_form', desc: true }];
}

export function TeamFormTable({ teams, players, focusMode, window, currentGw, onFocusChange, onWindowChange }: Props) {
  // Opponent form lookup by short name, so no-odds chips can colour by the
  // opponent's ovr_form (after GW 5) — matching the Players tab.
  const formByShortName = useMemo(
    () => new Map(teams.map(t => [t.short_name, t.ovr_form])),
    [teams],
  );
  const [sorting, setSorting]   = useState<SortingState>(defaultSort(focusMode));
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const isDesktop = useMediaQuery('(min-width: 768px)');

  const columns = makeColumns(focusMode);

  const table = useReactTable({
    data: teams,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  function handleFocusChange(mode: FocusMode) {
    onFocusChange(mode);
    setSorting(defaultSort(mode));
  }

  function toggle(teamId: string) {
    setExpanded(prev => {
      const next = new Set(prev);
      next.has(teamId) ? next.delete(teamId) : next.add(teamId);
      return next;
    });
  }

  return (
    <div>
      {/* Controls */}
      <div className="flex flex-wrap items-center gap-3 mb-3">
        <div className="flex rounded-md overflow-hidden border border-slate-600">
          {FOCUS_OPTIONS.map(opt => (
            <button
              key={opt.value}
              onClick={() => handleFocusChange(opt.value)}
              className={`px-3 py-1 text-xs font-medium transition-colors ${
                focusMode === opt.value
                  ? 'bg-slate-500 text-slate-100'
                  : 'bg-slate-800 text-slate-400 hover:bg-slate-700'
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>
        <div className="flex items-center gap-1.5">
          <span className="text-xs text-slate-500">GWs:</span>
          <div className="flex rounded-md overflow-hidden border border-slate-600">
            {WINDOW_OPTIONS.map(w => (
              <button
                key={w}
                onClick={() => onWindowChange(w)}
                className={`px-2.5 py-1 text-xs font-medium transition-colors ${
                  window === w
                    ? 'bg-slate-500 text-slate-100'
                    : 'bg-slate-800 text-slate-400 hover:bg-slate-700'
                }`}
              >
                {w}
              </button>
            ))}
          </div>
        </div>
      </div>

      {!isDesktop && (
        <MobileTeamCards
          rows={table.getRowModel().rows.map(r => r.original)}
          players={players}
          focusMode={focusMode}
          formByShortName={formByShortName}
          currentGw={currentGw}
          expanded={expanded}
          onToggle={toggle}
          sorting={sorting}
          setSorting={setSorting}
        />
      )}

      {isDesktop && (
      <div className="overflow-x-auto rounded-lg border border-slate-700/50 bg-slate-800/50">
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
            {table.getRowModel().rows.map(row => {
              const team = row.original;
              const isExpanded = expanded.has(team.id);
              const teamPlayers = inFormPlayers(players, team.id);

              return (
                <React.Fragment key={row.id}>
                  <tr className="border-b border-slate-700/50 hover:bg-slate-700/30 transition-colors">
                    {row.getVisibleCells().map(cell => {
                      if (cell.column.id === 'expand') {
                        return (
                          <td key={cell.id} className="px-3 py-2 w-8">
                            <button
                              onClick={() => toggle(team.id)}
                              className="text-slate-500 hover:text-slate-200 text-sm leading-none"
                              aria-label={isExpanded ? 'Collapse' : 'Expand'}
                            >
                              {isExpanded ? '▼' : '▶'}
                            </button>
                          </td>
                        );
                      }
                      if (cell.column.id === 'name') {
                        return (
                          <td key={cell.id} className="px-3 py-2 text-sm font-medium text-slate-100">
                            {team.name}
                          </td>
                        );
                      }
                      if (cell.column.id === 'fixtures') {
                        return (
                          <td key={cell.id} className="px-3 py-2">
                            <div className="flex gap-1 flex-wrap">
                              {team.fixtures.map((tf, i) => (
                                <FixtureChip
                                  key={i}
                                  fixture={toFixture(tf)}
                                  xg={tf.xg}
                                  csPct={tf.cs_pct}
                                  focusMode={focusMode}
                                  oppOvrForm={formByShortName.get(tf.opp)}
                                  currentGw={currentGw}
                                />
                              ))}
                            </div>
                          </td>
                        );
                      }
                      if (cell.column.id === 'att_form' || cell.column.id === 'def_form' || cell.column.id === 'ovr_form') {
                        return (
                          <td key={cell.id} className="px-3 py-2">
                            <FormBadge value={cell.getValue() as number} invert={cell.column.id === 'def_form'} />
                          </td>
                        );
                      }
                      if (cell.column.id === 'xg_sum') {
                        const val = team.xg_sum;
                        return (
                          <td key={cell.id} className="px-3 py-2 tabular-nums text-xs text-orange-300">
                            {val !== null ? val.toFixed(2) : '—'}
                          </td>
                        );
                      }
                      if (cell.column.id === 'cs_avg') {
                        const val = team.cs_avg;
                        return (
                          <td key={cell.id} className="px-3 py-2 tabular-nums text-xs text-sky-300">
                            {val !== null ? `${val.toFixed(0)}%` : '—'}
                          </td>
                        );
                      }
                      return (
                        <td key={cell.id} className="px-3 py-2 text-sm text-slate-300">
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </td>
                      );
                    })}
                  </tr>

                  {isExpanded && (
                    <tr className="bg-slate-700/20 border-b border-slate-700/50">
                      <td colSpan={columns.length} className="px-6 py-3">
                        {teamPlayers.length === 0 ? (
                          <span className="text-xs text-slate-500">No player data available.</span>
                        ) : (
                          <table className="w-full text-xs">
                            <thead>
                              <tr className="text-slate-500 uppercase tracking-wide">
                                <th className="text-left pb-1 pr-4">Player</th>
                                <th className="text-left pb-1 pr-4">Pos</th>
                                <th className="text-right pb-1 pr-4">Price</th>
                                <th className="text-right pb-1 pr-4">Form</th>
                                <th className="text-right pb-1">Global %</th>
                              </tr>
                            </thead>
                            <tbody>
                              {teamPlayers.map(p => (
                                <tr key={p.id} className="text-slate-300">
                                  <td className="pr-4 py-0.5">{p.name}</td>
                                  <td className="pr-4 py-0.5">
                                    <PositionBadge position={p.position} />
                                  </td>
                                  <td className="pr-4 py-0.5 text-right tabular-nums">£{p.price.toFixed(1)}m</td>
                                  <td className="pr-4 py-0.5 text-right tabular-nums">{p.form.toFixed(1)}</td>
                                  <td className="py-0.5 text-right tabular-nums">{p.global_ownership.toFixed(1)}%</td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        )}
                      </td>
                    </tr>
                  )}
                </React.Fragment>
              );
            })}
          </tbody>
        </table>
        <div className="px-4 py-2 text-xs text-slate-500 border-t border-slate-700/50">
          {teams.length} teams
        </div>
      </div>
      )}
    </div>
  );
}

function MobileTeamSortBar({
  focusMode,
  sorting,
  setSorting,
}: {
  focusMode: FocusMode;
  sorting: SortingState;
  setSorting: (s: SortingState) => void;
}) {
  const options = sortOptions(focusMode);
  const current = sorting[0] ?? { id: 'ovr_form', desc: true };

  return (
    <div className="mb-3 flex items-center gap-2">
      <label htmlFor="team-sort" className="text-xs text-slate-400">Sort</label>
      <select
        id="team-sort"
        value={current.id}
        onChange={e => setSorting([{ id: e.target.value, desc: e.target.value !== 'name' }])}
        className="flex-1 rounded-md border border-slate-600 bg-slate-700/50 px-2 py-1.5 text-sm text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500"
      >
        {options.map(o => (
          <option key={o.id} value={o.id}>{o.label}</option>
        ))}
      </select>
      <button
        type="button"
        onClick={() => setSorting([{ id: current.id, desc: !current.desc }])}
        aria-label={current.desc ? 'Sorted descending, tap for ascending' : 'Sorted ascending, tap for descending'}
        className="rounded-md border border-slate-600 bg-slate-700/50 px-3 py-1.5 text-sm text-slate-200 hover:bg-slate-600 focus:outline-none focus:ring-1 focus:ring-indigo-500"
      >
        {current.desc ? '↓' : '↑'}
      </button>
    </div>
  );
}

function MobileTeamCards({
  rows,
  players,
  focusMode,
  formByShortName,
  currentGw,
  expanded,
  onToggle,
  sorting,
  setSorting,
}: {
  rows: Team[];
  players: Player[];
  focusMode: FocusMode;
  formByShortName: Map<string, number>;
  currentGw?: number;
  expanded: Set<string>;
  onToggle: (id: string) => void;
  sorting: SortingState;
  setSorting: (s: SortingState) => void;
}) {
  return (
    <div>
      <MobileTeamSortBar focusMode={focusMode} sorting={sorting} setSorting={setSorting} />
      <div className="space-y-2">
        {rows.map(team => {
          const isExpanded = expanded.has(team.id);
          const teamPlayers = inFormPlayers(players, team.id);
          return (
            <div key={team.id} className="rounded-lg border border-slate-700/50 bg-slate-800/50">
              <button
                type="button"
                onClick={() => onToggle(team.id)}
                aria-expanded={isExpanded}
                className="w-full px-3 py-2 text-left focus:outline-none focus:ring-1 focus:ring-inset focus:ring-indigo-500 rounded-lg"
              >
                {/* Line 1 — name + overall form */}
                <div className="flex items-center justify-between gap-2">
                  <div className="flex items-center gap-1.5 min-w-0">
                    <span className="text-slate-500 text-xs leading-none">{isExpanded ? '▼' : '▶'}</span>
                    <span className="text-sm font-semibold text-slate-100 truncate">{team.name}</span>
                  </div>
                  <FormBadge value={team.ovr_form} />
                </div>

                {/* Line 2 — att / def (+ focus metric) */}
                <div className="mt-1.5 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs">
                  <span className="text-slate-500">Att</span>
                  <FormBadge value={team.att_form} />
                  <span className="text-slate-500">Def</span>
                  <FormBadge value={team.def_form} invert />
                  {focusMode === 'attack' && (
                    <span className="ml-auto tabular-nums text-orange-300">
                      xG {team.xg_sum !== null ? team.xg_sum.toFixed(2) : '—'}
                    </span>
                  )}
                  {focusMode === 'defense' && (
                    <span className="ml-auto tabular-nums text-sky-300">
                      CS {team.cs_avg !== null ? `${team.cs_avg.toFixed(0)}%` : '—'}
                    </span>
                  )}
                </div>

                {/* Line 3 — fixtures */}
                <div className="mt-1.5 flex flex-wrap gap-1">
                  {team.fixtures.map((tf, i) => (
                    <FixtureChip
                      key={i}
                      fixture={toFixture(tf)}
                      xg={tf.xg}
                      csPct={tf.cs_pct}
                      focusMode={focusMode}
                      oppOvrForm={formByShortName.get(tf.opp)}
                      currentGw={currentGw}
                      compact
                    />
                  ))}
                </div>
              </button>

              {isExpanded && (
                <div className="border-t border-slate-700/50 px-3 py-2">
                  {teamPlayers.length === 0 ? (
                    <span className="text-xs text-slate-500">No player data available.</span>
                  ) : (
                    <ul className="divide-y divide-slate-700/40">
                      {teamPlayers.map(p => (
                        <li key={p.id} className="flex items-center gap-2 py-1.5 text-xs">
                          <PositionBadge position={p.position} />
                          <span className="text-slate-200 truncate flex-1">{p.name}</span>
                          <span className="tabular-nums text-slate-400">£{p.price.toFixed(1)}m</span>
                          <span className="tabular-nums text-slate-300 w-10 text-right">F {p.form.toFixed(1)}</span>
                          <span className="tabular-nums text-slate-400 w-12 text-right">{p.global_ownership.toFixed(1)}%</span>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              )}
            </div>
          );
        })}
      </div>
      <p className="px-1 py-2 text-xs text-slate-500">{rows.length} teams</p>
    </div>
  );
}
