import type { TakenRow } from '../../types/shotzone';

interface Props {
  rows:     TakenRow[];
  search:   string;
  minShots: number;
}

// Top 5 by shots (default sort), matching the set-piece board's leaderboard
// shape.
export function TakenTable({ rows, search, minShots }: Props) {
  const q = search.trim().toLowerCase();
  const filtered = rows
    .filter((r) => r.shots >= minShots && (q === '' || r.player_name.toLowerCase().includes(q)))
    .sort((a, b) => b.shots - a.shots)
    .slice(0, 5);

  if (filtered.length === 0) {
    return <div className="text-sm text-slate-500 py-4">No qualifying shots.</div>;
  }

  return (
    <div className="rounded-xl border border-slate-700/50 overflow-hidden">
      <table className="w-full text-sm">
        <thead>
          <tr className="text-xs text-slate-500 border-b border-slate-700/60">
            <th className="text-left font-medium px-3 py-2 w-8">#</th>
            <th className="text-left font-medium px-3 py-2">Player</th>
            <th className="text-right font-medium px-3 py-2">Shots</th>
            <th className="text-right font-medium px-3 py-2">Goals</th>
            <th className="text-right font-medium px-3 py-2">xG</th>
            <th className="text-right font-medium px-3 py-2">xG/Shot</th>
            <th className="text-right font-medium px-3 py-2">G&minus;xG</th>
          </tr>
        </thead>
        <tbody>
          {filtered.map((r, i) => {
            const xgps = r.shots > 0 ? r.xg / r.shots : 0;
            const diff = r.goals - r.xg;
            const diffColor = diff > 0.3 ? 'text-emerald-400' : diff < -0.3 ? 'text-rose-400' : 'text-slate-400';
            return (
              <tr key={r.player_id} className="border-b border-slate-700/40 last:border-0">
                <td className="px-3 py-2.5 text-slate-500 tabular-nums">{i + 1}</td>
                <td className="px-3 py-2.5 font-semibold text-slate-100">{r.player_name}</td>
                <td className="px-3 py-2.5 text-right tabular-nums text-slate-200">{r.shots}</td>
                <td className="px-3 py-2.5 text-right tabular-nums text-slate-200">{r.goals}</td>
                <td className="px-3 py-2.5 text-right tabular-nums text-slate-400">{r.xg.toFixed(2)}</td>
                <td className="px-3 py-2.5 text-right tabular-nums text-slate-400">{xgps.toFixed(2)}</td>
                <td className={`px-3 py-2.5 text-right tabular-nums font-medium ${diffColor}`}>
                  {diff > 0 ? '+' : ''}{diff.toFixed(2)}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
