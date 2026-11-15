import { teamNameByCode } from '../setpiece/teamMeta';
import { TeamBadge } from '../ui/TeamBadge';
import { Card } from '../ui/Card';
import { InlineEmpty } from '../ui/EmptyState';
import type { TeamBonusEntry } from '../../types/stats';
import {
  BONUS_POSITIONS,
  bonusPerMatch,
  concentrationChip,
  positionShares,
  top3Share,
} from '../../utils/bonus';

// Same palette as PositionBadge and the scatter plot.
export const BONUS_POS_COLOR = {
  GK:  '#34d399',
  DEF: '#60a5fa',
  MID: '#a78bfa',
  FWD: '#f87171',
} as const;

const CHIP = {
  concentrated: { label: 'Concentrated', cls: 'bg-amber-400/10 text-amber-300', hint: 'Top 3 earn most of the club’s bonus' },
  spread:       { label: 'Spread',       cls: 'bg-emerald-400/10 text-emerald-300', hint: 'Bonus is shared across the squad' },
} as const;

/** One club: bonus rate, which position line earned it, and its top 3 earners. */
export function TeamBonusCard({ entry }: { entry: TeamBonusEntry }) {
  const shares = positionShares(entry);
  const chip = concentrationChip(entry);
  const top3 = Math.round(top3Share(entry));

  return (
    <Card>
      <div className="flex items-center justify-between gap-2">
        <h3 className="flex min-w-0 items-center gap-2 text-sm font-semibold text-slate-100">
          <TeamBadge code={entry.team} size="sm" />
          <span className="truncate">{teamNameByCode(entry.team)}</span>
        </h3>
        <span className="shrink-0 tabular-nums">
          <span className="text-lg font-semibold text-slate-100">{bonusPerMatch(entry).toFixed(1)}</span>
          <span className="text-[11px] text-slate-500"> /match</span>
        </span>
      </div>
      <div className="mt-0.5 text-[11px] text-slate-500 tabular-nums">
        {entry.bonus} bonus · {entry.matches} {entry.matches === 1 ? 'match' : 'matches'}
      </div>

      {entry.bonus > 0 && (
        <>
          <div
            className="mt-3 flex h-2.5 gap-0.5 overflow-hidden rounded"
            role="img"
            aria-label={BONUS_POSITIONS.map((p) => `${p} ${Math.round(shares[p])}%`).join(', ')}
          >
            {BONUS_POSITIONS.map((p) =>
              entry.by_position[p] > 0 ? (
                <div key={p} style={{ flex: entry.by_position[p], backgroundColor: BONUS_POS_COLOR[p] }} />
              ) : null,
            )}
          </div>
          <div className="mt-1.5 flex flex-wrap gap-x-3 text-[11px] text-slate-400 tabular-nums">
            {BONUS_POSITIONS.map((p) => (
              <span key={p}>
                {p} {Math.round(shares[p])}%
              </span>
            ))}
          </div>
        </>
      )}

      <div className="mt-3 flex items-center gap-2">
        {chip && (
          <span title={CHIP[chip].hint} className={`rounded px-2 py-0.5 text-[11px] font-medium ${CHIP[chip].cls}`}>
            {CHIP[chip].label}
          </span>
        )}
        {entry.bonus > 0 && <span className="text-[11px] text-slate-500">top 3 = {top3}%</span>}
      </div>

      {entry.top.length === 0 ? (
        <InlineEmpty>No bonus yet.</InlineEmpty>
      ) : (
        <ol className="mt-2 space-y-0.5 text-sm">
          {entry.top.map((p) => (
            <li key={p.id} className="flex items-center gap-2">
              <span className="min-w-0 flex-1 truncate text-slate-200">
                {p.name}
                <span className="ml-1.5 text-[10px] uppercase text-slate-500">{p.position}</span>
              </span>
              <span className="shrink-0 font-semibold text-slate-100 tabular-nums">{p.bonus}</span>
            </li>
          ))}
        </ol>
      )}
    </Card>
  );
}
