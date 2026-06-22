import type { Player } from '../../types/player';
import { PositionBadge } from '../common/PositionBadge';
import { FixtureChip } from './FixtureChip';

const STATUS_DOT: Record<Player['status'], string> = {
  available: 'bg-green-500',
  doubt:     'bg-amber-400',
  injured:   'bg-red-500',
};

export function PlayerCard({ player }: { player: Player }) {
  const diff = player.top_n_ownership - player.global_ownership;
  const diffColor = diff > 0 ? 'text-emerald-400' : diff < 0 ? 'text-red-400' : 'text-slate-400';
  const diffSign  = diff > 0 ? '+' : '';
  const isInjured = player.status === 'injured';

  return (
    <div className={`rounded-lg border border-slate-700/50 bg-slate-800/50 p-3 ${isInjured ? 'opacity-40' : ''}`}>
      {/* Header row */}
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <span className={`shrink-0 w-2 h-2 rounded-full ${STATUS_DOT[player.status]}`} title={player.news || player.status} />
          <span className="text-sm font-semibold text-slate-100">{player.name}</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-slate-400">{player.team.short_name}</span>
          <PositionBadge position={player.position} />
        </div>
      </div>

      {/* Stats row */}
      <div className="grid grid-cols-4 gap-2 mb-2 text-center">
        <div>
          <p className="text-[10px] text-slate-500 uppercase">Price</p>
          <p className="text-xs text-slate-300 tabular-nums">£{player.price.toFixed(1)}m</p>
        </div>
        <div>
          <p className="text-[10px] text-slate-500 uppercase">Form</p>
          <p className="text-xs text-slate-300 tabular-nums">{player.form.toFixed(1)}</p>
        </div>
        <div>
          <p className="text-[10px] text-slate-500 uppercase">Global</p>
          <p className="text-xs text-slate-300 tabular-nums">{player.global_ownership.toFixed(1)}%</p>
        </div>
        <div>
          <p className="text-[10px] text-slate-500 uppercase">Diff</p>
          <p className={`text-xs font-semibold tabular-nums ${diffColor}`}>{diffSign}{diff.toFixed(1)}%</p>
        </div>
      </div>

      {/* Fixtures */}
      <div className="flex gap-1 flex-wrap">
        {player.fixtures.slice(0, 5).map((f, i) => (
          <FixtureChip key={i} fixture={f} />
        ))}
      </div>
    </div>
  );
}
