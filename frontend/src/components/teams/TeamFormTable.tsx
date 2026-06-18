import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
  type SortingState,
} from '@tanstack/react-table';
import { useState } from 'react';
import type { Team, TeamFixture } from '../../types/team';
import type { Player } from '../../types/player';
import { FixtureChip } from '../players/FixtureChip';

const ch = createColumnHelper<Team>();

const columns = [
  ch.display({ id: 'expand', header: '', enableSorting: false }),
  ch.accessor('name',      { header: 'Team',     enableSorting: true }),
  ch.accessor('att_form',  { header: 'Att Form', enableSorting: true }),
  ch.accessor('def_form',  { header: 'Def Form', enableSorting: true }),
  ch.accessor('ovr_form',  { header: 'Ovr Form', enableSorting: true }),
  ch.display({ id: 'fixtures', header: 'Next 5 GWs', enableSorting: false }),
];

interface Props {
  teams: Team[];
  players: Player[];
}

function toFixture(tf: TeamFixture) {
  return tf as { gw: number; opp: string; ha: 'H' | 'A'; difficulty: 1 | 2 | 3 | 4 | 5; kickoff: string };
}

const posOrder: Record<string, number> = { GK: 0, DEF: 1, MID: 2, FWD: 3 };

export function TeamFormTable({ teams, players }: Props) {
  const [sorting, setSorting] = useState<SortingState>([
    { id: 'ovr_form', desc: true },
  ]);
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
    <div className="overflow-x-auto rounded-lg border border-gray-200 bg-white">
      <table className="min-w-full text-left">
        <thead className="bg-gray-50 border-b border-gray-200">
          {table.getHeaderGroups().map(hg => (
            <tr key={hg.id}>
              {hg.headers.map(header => {
                const canSort = header.column.getCanSort();
                const sorted  = header.column.getIsSorted();
                return (
                  <th
                    key={header.id}
                    className={`px-3 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wide select-none ${canSort ? 'cursor-pointer hover:text-gray-700' : ''}`}
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
              .filter(p => p.team.id === team.id && p.form > 0)
              .sort((a, b) => posOrder[a.position] - posOrder[b.position] || b.global_ownership - a.global_ownership);

            return (
              <>
                <tr key={row.id} className="border-b border-gray-100 hover:bg-gray-50">
                  {row.getVisibleCells().map(cell => {
                    if (cell.column.id === 'expand') {
                      return (
                        <td key={cell.id} className="px-3 py-2 w-8">
                          <button
                            onClick={() => toggle(team.id)}
                            className="text-gray-400 hover:text-gray-700 text-sm leading-none"
                            aria-label={isExpanded ? 'Collapse' : 'Expand'}
                          >
                            {isExpanded ? '▼' : '▶'}
                          </button>
                        </td>
                      );
                    }
                    if (cell.column.id === 'fixtures') {
                      return (
                        <td key={cell.id} className="px-3 py-2">
                          <div className="flex gap-1 flex-wrap">
                            {team.fixtures.slice(0, 5).map((tf, i) => (
                              <FixtureChip key={i} fixture={toFixture(tf)} />
                            ))}
                          </div>
                        </td>
                      );
                    }
                    if (cell.column.id === 'att_form' || cell.column.id === 'def_form' || cell.column.id === 'ovr_form') {
                      const val = cell.getValue() as number;
                      return (
                        <td key={cell.id} className="px-3 py-2 text-sm tabular-nums font-medium text-gray-800">
                          {val.toFixed(1)}
                        </td>
                      );
                    }
                    return (
                      <td key={cell.id} className="px-3 py-2 text-sm text-gray-800">
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </td>
                    );
                  })}
                </tr>
                {isExpanded && (
                  <tr key={`${row.id}-expand`} className="bg-gray-50 border-b border-gray-200">
                    <td colSpan={columns.length} className="px-6 py-3">
                      {teamPlayers.length === 0 ? (
                        <span className="text-xs text-gray-400">No player data available.</span>
                      ) : (
                        <table className="w-full text-xs">
                          <thead>
                            <tr className="text-gray-400 uppercase tracking-wide">
                              <th className="text-left pb-1 pr-4">Player</th>
                              <th className="text-left pb-1 pr-4">Pos</th>
                              <th className="text-right pb-1 pr-4">Price</th>
                              <th className="text-right pb-1 pr-4">Form</th>
                              <th className="text-right pb-1">Global %</th>
                            </tr>
                          </thead>
                          <tbody>
                            {teamPlayers.map(p => (
                              <tr key={p.id} className="text-gray-700">
                                <td className="pr-4 py-0.5">{p.name}</td>
                                <td className="pr-4 py-0.5 text-gray-500">{p.position}</td>
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
              </>
            );
          })}
        </tbody>
      </table>
      <div className="px-4 py-2 text-xs text-gray-400 border-t border-gray-100">
        {teams.length} teams
      </div>
    </div>
  );
}
