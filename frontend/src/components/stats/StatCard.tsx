import { Card } from '../ui/Card';
import { TeamBadge } from '../ui/TeamBadge';
import { InlineEmpty } from '../ui/EmptyState';
import type { StatCard as StatCardData } from '../../types/stats';

interface Props {
  card: StatCardData;
}

/**
 * A single scoring-component card: header with the component label and its FPL
 * point value, followed by the ranked top players (by that component summed over
 * the recent-GW window).
 */
export function StatCard({ card }: Props) {
  return (
    <Card>
      <div className="flex items-center justify-between gap-2 mb-2">
        <h3 className="text-sm font-semibold text-slate-100 truncate">{card.label}</h3>
        <span className="shrink-0 text-[11px] font-medium text-emerald-300 bg-emerald-400/10 rounded px-1.5 py-0.5 tabular-nums">
          {card.points}
        </span>
      </div>

      {card.leaders.length === 0 ? (
        <InlineEmpty>No data yet.</InlineEmpty>
      ) : (
        <ol className="space-y-1">
          {card.leaders.map((l) => (
            <li key={l.id} className="flex items-center gap-2 text-sm">
              <span className="w-4 shrink-0 text-right text-xs text-slate-500 tabular-nums">
                {l.rank}
              </span>
              <span className="min-w-0 flex-1 truncate text-slate-200">{l.name}</span>
              <TeamBadge code={l.team} size="xs" />
              <span className="w-8 shrink-0 text-right font-semibold text-slate-100 tabular-nums">
                {l.value}
              </span>
            </li>
          ))}
        </ol>
      )}
    </Card>
  );
}
