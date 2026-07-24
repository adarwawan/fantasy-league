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

export function PlayerCard({
  player,
  teams,
  currentGw,
  onClick,
}: {
  player: Player;
  teams?: Team[];
  currentGw?: number;
  onClick?: (p: Player) => void;
}) {
  const diff = player.top_n_ownership - player.global_ownership;
  const diffColor = diff > 0 ? 'text-emerald-400' : diff < 0 ? 'text-red-400' : 'text-slate-400';
  const diffSign  = diff > 0 ? '+' : '';
  const isInjured = player.status === 'injured';

  return (
    <button
      type="button"
      onClick={onClick ? () => onClick(player) : undefined}
      aria-label={onClick ? `View ${player.name}` : undefined}
      className={`w-full text-left rounded-lg border border-slate-700/50 bg-slate-800/50 px-3 py-2 focus:outline-none focus:ring-1 focus:ring-inset focus:ring-indigo-500 ${isInjured ? 'opacity-40' : ''}`}
    >
      {/* Line 1 — name + team/pos */}
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-1.5 min-w-0">
          <span className={`shrink-0 w-2 h-2 rounded-full ${STATUS_DOT[player.status]}`} title={player.news || player.status} />
          <span className="text-sm font-semibold text-slate-100 truncate">{player.name}</span>
          {/* First- and second-choice set-piece duties; full order in the drawer */}
          <SetPieceBadges player={player} maxOrder={2} />
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          <span className="text-xs text-slate-400">{player.team.short_name}</span>
          <PositionBadge position={player.position} />
        </div>
      </div>

      {/* Line 2 — inline stats */}
      <div className="mt-1 flex items-center gap-x-3 text-xs tabular-nums">
        <span className="text-slate-300">£{player.price.toFixed(1)}m</span>
        <span className="text-slate-400">Form <span className="text-slate-200">{player.form.toFixed(1)}</span></span>
        <span className="text-slate-400">Own <span className="text-slate-200">{player.global_ownership.toFixed(1)}%</span></span>
        <span className="text-slate-400">EO <span className="text-slate-200">{player.effective_ownership.toFixed(1)}%</span></span>
        <span className={`ml-auto font-semibold ${diffColor}`}>{diffSign}{diff.toFixed(1)}%</span>
      </div>

      {/* Line 3 — fixtures (next 5 GWs; double gameweeks may exceed 5) */}
      <div className="mt-1.5 flex flex-wrap gap-1">
        {player.fixtures.map((f, i) => (
          <FixtureChip
            key={i}
            fixture={f}
            xg={f.xg}
            csPct={f.cs_pct}
            focusMode="overall"
            oppOvrForm={teams?.find(t => t.short_name === f.opp)?.ovr_form}
            currentGw={currentGw}
            compact
          />
        ))}
      </div>
    </button>
  );
}
