import { LeaderRow } from './LeaderRow';
import type { LeaderRow as LeaderRowType, TrendsMetric } from '../../types/trends';

interface Props {
  title: string;
  accent: 'in' | 'out';
  metric: TrendsMetric;
  rows: LeaderRowType[];
  expanded: number | null;
  onToggle: (id: number) => void;
}

// LeaderBoard renders one ranked column (inflows or outflows). Bars scale to the
// column's own peak magnitude so each side reads on its own terms.
export function LeaderBoard({ title, accent, metric, rows, expanded, onToggle }: Props) {
  const maxMagnitude = rows.reduce((m, r) => Math.max(m, Math.abs(r.rank_delta)), 0);
  const dot = accent === 'out' ? 'bg-red-400' : 'bg-emerald-400';

  return (
    <div className="rounded-lg border border-slate-800/60 overflow-hidden">
      <div className="flex items-center gap-2 px-3 py-2 border-b border-slate-800 bg-slate-900/40">
        <span className={`inline-block w-2 h-2 rounded-full ${dot}`} />
        <span className="text-sm font-medium text-slate-200">{title}</span>
      </div>
      <table className="w-full text-sm">
        <thead>
          <tr className="text-xs uppercase tracking-wide text-slate-500 border-b border-slate-800">
            <th className="py-2 pl-3 pr-2 text-left font-medium w-8">#</th>
            <th className="py-2 pr-2 text-left font-medium">Player</th>
            <th className="py-2 pr-2 text-right font-medium">Own</th>
            <th className="py-2 pr-3 text-right font-medium">Δ</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <LeaderRow
              key={row.player_ext_id}
              row={row}
              rank={i + 1}
              maxMagnitude={maxMagnitude}
              metric={metric}
              expanded={expanded === row.player_ext_id}
              onToggle={() => onToggle(row.player_ext_id)}
            />
          ))}
        </tbody>
      </table>
      {rows.length === 0 && (
        <div className="text-sm text-slate-500 py-8 text-center">No movement captured yet.</div>
      )}
    </div>
  );
}
