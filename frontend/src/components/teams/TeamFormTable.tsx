import React, { useState } from 'react';
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
  type SortingState,
} from '@tanstack/react-table';
import type { Team, TeamFixture } from '../../types/team';
import type { Player } from '../../types/player';
import { PositionBadge } from '../common/PositionBadge';
import { FixtureChip } from '../players/FixtureChip';

const ch = createColumnHelper<Team>();

const columns = [
  ch.display({ id: 'expand',   header: '',          enableSorting: false }),
  ch.accessor('name',          { header: 'Team',     enableSorting: true }),
  ch.accessor('att_form',      { header: 'Att',      enableSorting: true }),
  ch.accessor('def_form',      { header: 'Def',      enableSorting: true }),
  ch.accessor('ovr_form',      { header: 'Form',     enableSorting: true }),
  ch.display({ id: 'fixtures', header: 'Next 5 GWs', enableSorting: false }),
];

interface Props {
  teams:   Team[];
  players: Player[];
}

function toFixture(tf: TeamFixture) {
  return tf as { gw: number; opp: string; ha: 'H' | 'A'; difficulty: 1 | 2 | 3 | 4 | 5; kickoff: string };
}

function FormBadge({ value, invert = false }: { value: number; invert?: boolean }) {
  // invert=true for defensive form: lower value = fewer goals conceded = better
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

export function TeamFormTable({ teams, players }: Props) {
  const [sorting, setSorting]   = useState<SortingState>([{ id: 'ovr_form', desc: true }]);
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  const table = useReactTable({
    data: teams,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  function toggle(teamId: string) {
    setExpanded(prev => {
      const next = new Set(prev);
      next.has(teamId) ? next.delete(teamId) : next.add(teamId);
      return next;
    });
  }

  return (
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
            const teamPlayers = players
              .filter(p => p.team.id === team.id && p.form > 1.5)
              .sort((a, b) => posOrder[a.position] - posOrder[b.position] || b.global_ownership - a.global_ownership);

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
                            {team.fixtures.slice(0, 5).map((tf, i) => {
                              const oppTeam = teams.find(t => t.short_name === tf.opp);
                              return (
                                <FixtureChip key={i} fixture={toFixture(tf)} oppOvrForm={oppTeam?.ovr_form} />
                              );
                            })}
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
  );
}
