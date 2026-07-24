import { forwardRef } from 'react';
import type { Row } from '@tanstack/react-table';
import type { Player } from '../../types/player';
import type { Team } from '../../types/team';
import { PositionBadge } from '../common/PositionBadge';
import { SetPieceBadges } from '../common/SetPieceBadge';
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

interface PlayerRowProps {
  row: Row<Player>;
  teams?: Team[];
  currentGw?: number;
  onPlayerClick?: (p: Player) => void;
  'data-index'?: number;
}

export const PlayerRow = forwardRef<HTMLTableRowElement, PlayerRowProps>(function PlayerRow(
  { row, teams, currentGw, onPlayerClick, 'data-index': dataIndex },
  ref,
) {
  const player = row.original;
  const isInjured = player.status === 'injured';
  const diff = player.top_n_ownership - player.global_ownership;
  const diffColor = diff > 0 ? 'text-emerald-400' : diff < 0 ? 'text-red-400' : 'text-slate-400';
  const diffSign  = diff > 0 ? '+' : '';

  function handleKeyDown(e: React.KeyboardEvent<HTMLTableRowElement>) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onPlayerClick?.(player);
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      const next = e.currentTarget.nextElementSibling as HTMLElement | null;
      next?.focus();
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      const prev = e.currentTarget.previousElementSibling as HTMLElement | null;
      prev?.focus();
    }
  }

  return (
    <tr
      ref={ref}
      data-index={dataIndex}
      tabIndex={onPlayerClick ? 0 : undefined}
      role={onPlayerClick ? 'button' : undefined}
      aria-label={onPlayerClick ? `View ${player.name}` : undefined}
      className={`border-b border-slate-700/50 hover:bg-slate-700/30 transition-colors focus:outline-none focus:bg-slate-700/40 focus:ring-1 focus:ring-inset focus:ring-indigo-500 ${isInjured ? 'opacity-40' : ''} ${onPlayerClick ? 'cursor-pointer' : ''}`}
      onClick={onPlayerClick ? () => onPlayerClick(player) : undefined}
      onKeyDown={onPlayerClick ? handleKeyDown : undefined}
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
          {/* First- and second-choice set-piece duties; full order in the drawer */}
          <SetPieceBadges player={player} maxOrder={2} />
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

      {/* Effective ownership bar (multiplier-weighted; can exceed 100%) */}
      <td className="px-3 py-2">
        <OwnershipBar value={player.effective_ownership} position={player.position} max={100} />
      </td>

      {/* Differential */}
      <td className={`px-3 py-2 text-sm font-semibold text-right tabular-nums ${diffColor}`}>
        {diffSign}{diff.toFixed(1)}%
      </td>

      {/* Next 5 GWs' fixture chips (backend windows by gameweek, so a double
          gameweek may yield more than 5 fixtures) */}
      <td className="px-3 py-2">
        <div className="flex gap-1 flex-wrap">
          {player.fixtures.map((f, i) => (
              <FixtureChip
                key={i}
                fixture={f}
                xg={f.xg}
                csPct={f.cs_pct}
                focusMode="overall"
                oppOvrForm={teams?.find(t => t.short_name === f.opp)?.ovr_form}
                currentGw={currentGw}
              />
            ))}
        </div>
      </td>
    </tr>
  );
});
