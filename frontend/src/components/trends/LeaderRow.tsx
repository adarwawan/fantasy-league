import { PositionBadge } from '../common/PositionBadge';
import { Sparkline } from './Sparkline';
import { useTrendsSeries } from '../../hooks/useTrends';
import type { LeaderRow as LeaderRowType, TrendsMetric } from '../../types/trends';
import type { Player } from '../../types/player';

const KNOWN_POS = new Set(['GK', 'DEF', 'MID', 'FWD']);

// fmtDelta renders rank_delta in the active metric's unit: percentage points for
// ownership (basis points ÷ 100), or a transfer count (k-abbreviated).
function fmtDelta(n: number, metric: TrendsMetric): string {
  const sign = n > 0 ? '+' : n < 0 ? '−' : '';
  const abs = Math.abs(n);
  if (metric === 'ownership') return `${sign}${(abs / 100).toFixed(1)}%`;
  if (abs >= 1000) return `${sign}${(abs / 1000).toFixed(1)}k`;
  return `${sign}${abs}`;
}

interface Props {
  row: LeaderRowType;
  rank: number;
  maxMagnitude: number;
  metric: TrendsMetric;
  expanded: boolean;
  onToggle: () => void;
}

export function LeaderRow({ row, rank, maxMagnitude, metric, expanded, onToggle }: Props) {
  const { data: series } = useTrendsSeries(expanded ? row.player_ext_id : null);
  const magnitude = Math.abs(row.rank_delta);
  const barPct = maxMagnitude > 0 ? Math.max(2, (magnitude / maxMagnitude) * 100) : 0;
  const outflow = row.rank_delta < 0;
  const barColor = outflow ? 'bg-red-400/80' : 'bg-emerald-400/80';
  const textColor = outflow ? 'text-red-400' : 'text-emerald-400';
  const isOwnership = metric === 'ownership';
  const sparkValues = series?.series.map((p) => (isOwnership ? p.selected_pct : p.net_transfers)) ?? [];
  const sparkLabel = isOwnership ? 'Ownership % this session' : 'Net transfers this session';

  return (
    <>
      <tr
        onClick={onToggle}
        className="border-b border-slate-800/60 hover:bg-slate-800/40 cursor-pointer transition-colors"
      >
        <td className="py-2 pl-3 pr-2 text-slate-500 tabular-nums text-sm w-8">{rank}</td>
        <td className="py-2 pr-2">
          <div className="flex items-center gap-2">
            {KNOWN_POS.has(row.position) && <PositionBadge position={row.position as Player['position']} />}
            <span className="text-slate-100 font-medium">{row.name}</span>
            <span className="text-xs text-slate-500">{row.team}</span>
          </div>
        </td>
        <td className="py-2 pr-2 text-right tabular-nums text-slate-300 text-sm">{row.selected_pct.toFixed(1)}%</td>
        <td className="py-2 pr-3 w-32">
          <div className="flex items-center gap-2 justify-end">
            <div className="flex-1 h-1.5 rounded-full bg-slate-800 overflow-hidden hidden sm:block">
              <div className={`h-full rounded-full ${barColor}`} style={{ width: `${barPct}%` }} />
            </div>
            <span className={`${textColor} font-semibold tabular-nums text-sm w-12 text-right`}>
              {fmtDelta(row.rank_delta, metric)}
            </span>
          </div>
        </td>
      </tr>
      {expanded && (
        <tr className="border-b border-slate-800/60 bg-slate-900/40">
          <td colSpan={4} className="px-3 py-3">
            <div className="flex items-center gap-4">
              <div>
                <div className="text-xs text-slate-500 mb-1">{sparkLabel}</div>
                <Sparkline values={sparkValues} width={220} height={44} />
              </div>
              <div className="text-xs text-slate-400 leading-5">
                {series?.series.length
                  ? `${series.series.length} snapshots · ${
                      isOwnership
                        ? `now ${row.selected_pct.toFixed(1)}%`
                        : `latest net ${row.net_transfers.toLocaleString()}`
                    }`
                  : 'Loading series…'}
              </div>
            </div>
          </td>
        </tr>
      )}
    </>
  );
}
