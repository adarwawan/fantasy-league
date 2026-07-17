import type { TeamICTEntry, TeamICTPlayer } from '../../types/stats';

interface Props {
  entry: TeamICTEntry;
}

/**
 * Star badge for a team-internal top-3 rank in one ICT component: a tinted
 * pill (same idiom as the points chips on the stat cards) holding a star
 * glyph and the rank at full text size, so the digit stays legible.
 */
function StarBadge({ rank, component, colorClasses }: { rank: number; component: string; colorClasses: string }) {
  return (
    <span
      title={`#${rank} ${component} in team`}
      className={`flex shrink-0 items-center gap-0.5 rounded px-1 py-px text-[10px] font-bold tabular-nums ${colorClasses}`}
    >
      <svg viewBox="0 0 24 24" className="h-2.5 w-2.5" fill="currentColor" aria-hidden>
        <path d="M12 1.5l3.1 6.53 7.15.92-5.25 4.98 1.35 7.07L12 17.6 5.65 21l1.35-7.07-5.25-4.98 7.15-.92z" />
      </svg>
      {rank}
    </span>
  );
}

/** The player's star badges: one per ICT component he is top 3 in team for. */
function PlayerBadges({ p }: { p: TeamICTPlayer }) {
  return (
    <span className="flex shrink-0 items-center gap-1">
      {p.influence_rank && (
        <StarBadge rank={p.influence_rank} component="influence" colorClasses="text-emerald-300 bg-emerald-400/10" />
      )}
      {p.creativity_rank && (
        <StarBadge rank={p.creativity_rank} component="creativity" colorClasses="text-sky-300 bg-sky-400/10" />
      )}
      {p.threat_rank && (
        <StarBadge rank={p.threat_rank} component="threat" colorClasses="text-rose-300 bg-rose-400/10" />
      )}
    </span>
  );
}

/**
 * One team's ICT share card: the players carrying the team's underlying
 * attacking involvement over the recent-GW window. Each row shows the player's
 * share of the team's total ICT as a bar, split into Influence / Creativity /
 * Threat segments so the composition is visible at a glance (a threat-heavy
 * bar is a finisher, a creativity-heavy bar is the playmaker).
 */
export function TeamICTCard({ entry }: Props) {
  return (
    <div className="rounded-xl border border-slate-700/60 bg-slate-800/60 p-3">
      <div className="flex items-center justify-between gap-2 mb-2">
        <h3 className="text-sm font-semibold text-slate-100">{entry.team}</h3>
        <span className="shrink-0 text-[11px] font-medium text-slate-400 tabular-nums">
          team ICT {entry.total_ict}
        </span>
      </div>

      {entry.players.length === 0 ? (
        <p className="text-xs text-slate-500 py-2">No data yet.</p>
      ) : (
        <ol className="space-y-1.5">
          {entry.players.map((p, i) => (
            <li key={p.id} className="text-sm">
              {/* Players past index 4 are badge-holders outside the top 5 by
                  combined ICT (see teamICTLeaderLimit in the backend). */}
              {i === 5 && <div className="mb-1.5 border-t border-dashed border-slate-700/60" />}
              <div className="flex items-center gap-2">
                <span className="min-w-0 flex-1 truncate text-slate-200">
                  {p.name}
                  <span className="ml-1.5 text-[10px] uppercase text-slate-500">{p.position}</span>
                </span>
                <PlayerBadges p={p} />
                <span className="shrink-0 text-[11px] text-slate-500 tabular-nums">{p.ict}</span>
                <span className="w-11 shrink-0 text-right font-semibold text-slate-100 tabular-nums">
                  {p.share}%
                </span>
              </div>
              <div className="mt-0.5 flex h-1.5 w-full overflow-hidden rounded-full bg-slate-700/40">
                <div
                  className="bg-emerald-400/80"
                  style={{ width: `${(p.share * p.influence) / p.ict}%` }}
                />
                <div
                  className="bg-sky-400/80"
                  style={{ width: `${(p.share * p.creativity) / p.ict}%` }}
                />
                <div
                  className="bg-rose-400/80"
                  style={{ width: `${(p.share * p.threat) / p.ict}%` }}
                />
              </div>
            </li>
          ))}
        </ol>
      )}
    </div>
  );
}
