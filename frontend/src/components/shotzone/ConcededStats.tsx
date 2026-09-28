import type { ConcededTotals } from '../../types/shotzone';

// Shots-conceded mode shows only the team's aggregate totals for the zone —
// no player breakdown — a defensive-weakness signal, not a leaderboard.
export function ConcededStats({ totals }: { totals: ConcededTotals }) {
  const xgps = totals.shots > 0 ? totals.xg / totals.shots : 0;
  const diff = totals.goals - totals.xg;
  const diffColor = diff > 0.3 ? 'text-rose-400' : diff < -0.3 ? 'text-emerald-400' : 'text-slate-400';

  return (
    <div className="grid grid-cols-5 gap-3 tabular-nums">
      <Stat label="Shots" value={totals.shots} />
      <Stat label="Goals" value={totals.goals} />
      <Stat label="xG" value={totals.xg.toFixed(2)} />
      <Stat label="xG/Shot" value={xgps.toFixed(2)} />
      <Stat label="G−xG" value={`${diff > 0 ? '+' : ''}${diff.toFixed(2)}`} valueClass={diffColor} />
    </div>
  );
}

function Stat({ label, value, valueClass = 'text-slate-100' }: { label: string; value: string | number; valueClass?: string }) {
  return (
    <div>
      <div className="text-xs text-slate-500 mb-0.5">{label}</div>
      <div className={`text-base font-medium ${valueClass}`}>{value}</div>
    </div>
  );
}
