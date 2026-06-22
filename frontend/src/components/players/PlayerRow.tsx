import type { Row } from '@tanstack/react-table';
import type { Player } from '../../types/player';
import { PositionBadge } from '../common/PositionBadge';
import { FixtureChip } from './FixtureChip';

const STATUS_DOT: Record<Player['status'], string> = {
  available: 'bg-green-500',
  doubt:     'bg-amber-400',
  injured:   'bg-red-500',
};

const POS_BAR: Record<Player['position'], string> = {
  GK:  'bg-emerald-400',
  DEF: 'bg-blue-400',
  MID: 'bg-purple-400',
  FWD: 'bg-red-400',
};

function OwnershipBar({ value, position, max = 80 }: { value: number; position: Player['position']; max?: number }) {
  const pct = Math.min((value / max) * 100, 100);
  return (
    <div className="flex items-center gap-2 min-w-[80px]">
      <div className="flex-1 h-1.5 rounded-full bg-slate-700">
        <div
          className={`h-full rounded-full ${POS_BAR[position]}`}
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="text-xs text-slate-300 tabular-nums w-10 text-right">{value.toFixed(1)}%</span>
    </div>
  );
}

export function PlayerRow({ row, onPlayerClick }: { row: Row<Player>; onPlayerClick?: (p: Player) => void }) {
  const player = row.original;
  const isInjured = player.status === 'injured';
  const diff = player.top_n_ownership - player.global_ownership;
  const diffColor = diff > 0 ? 'text-emerald-400' : diff < 0 ? 'text-red-400' : 'text-slate-400';
  const diffSign  = diff > 0 ? '+' : '';

  return (
    <tr
      className={`border-b border-slate-700/50 hover:bg-slate-700/30 transition-colors ${isInjured ? 'opacity-40' : ''} ${onPlayerClick ? 'cursor-pointer' : ''}`}
      onClick={onPlayerClick ? () => onPlayerClick(player) : undefined}
    >
      {/* Name + status */}
      <td className="px-3 py-2 text-sm font-medium text-slate-100 whitespace-nowrap">
        <div className="flex items-center gap-1.5">
          <span
            className={`shrink-0 w-2 h-2 rounded-full ${STATUS_DOT[player.status]}`}
            title={player.news || player.status}
          />
          <span className="cursor-default" title={player.news || undefined}>
            {player.name}
          </span>
        </div>
      </td>

      {/* Team */}
      <td className="px-3 py-2 text-sm text-slate-400">{player.team.short_name}</td>

      {/* Position */}
      <td className="px-3 py-2">
        <PositionBadge position={player.position} />
      </td>

      {/* Price */}
      <td className="px-3 py-2 text-sm text-slate-300 text-right tabular-nums">
        £{player.price.toFixed(1)}m
      </td>

      {/* Form */}
      <td className="px-3 py-2 text-sm text-slate-300 text-right tabular-nums">
        {player.form.toFixed(1)}
      </td>

      {/* Global ownership bar */}
      <td className="px-3 py-2">
        <OwnershipBar value={player.global_ownership} position={player.position} />
      </td>

      {/* Top-N ownership bar */}
      <td className="px-3 py-2">
        <OwnershipBar value={player.top_n_ownership} position={player.position} />
      </td>

      {/* Differential */}
      <td className={`px-3 py-2 text-sm font-semibold text-right tabular-nums ${diffColor}`}>
        {diffSign}{diff.toFixed(1)}%
      </td>

      {/* Next 5 fixture chips */}
      <td className="px-3 py-2">
        <div className="flex gap-1 flex-wrap">
          {player.fixtures.slice(0, 5).map((f, i) => (
            <FixtureChip key={i} fixture={f} />
          ))}
        </div>
      </td>
    </tr>
  );
}
