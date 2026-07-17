import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useStats, useTeamICT } from '../hooks/useStats';
import { StatCard } from '../components/stats/StatCard';
import { TeamICTCard } from '../components/stats/TeamICTCard';
import { ErrorState } from '../components/common/ErrorState';

type StatsView = 'position' | 'team';

const ictLegend = [
  { label: 'Influence',  dot: 'bg-emerald-400' },
  { label: 'Creativity', dot: 'bg-sky-400' },
  { label: 'Threat',     dot: 'bg-rose-400' },
];

function LoadingGrid() {
  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: 6 }).map((_, i) => (
        <div key={i} className="h-40 rounded-xl border border-slate-700/50 bg-slate-800/40 animate-pulse" />
      ))}
    </div>
  );
}

export function StatsPage() {
  const { game = 'fpl' } = useParams<{ game: string }>();
  const [view, setView] = useState<StatsView>('position');
  const stats = useStats(game);
  const teamICT = useTeamICT(game, view === 'team');

  useEffect(() => {
    document.title = `${game.toUpperCase()} — Stats`;
  }, [game]);

  const active = view === 'position' ? stats : teamICT;
  const window = stats.data?.meta.window ?? teamICT.data?.meta.window;

  return (
    <div>
      <div className="flex items-baseline justify-between gap-2 mb-3">
        <h1 className="text-xl font-semibold text-slate-100">{game.toUpperCase()} — Stats</h1>
        {window != null && <span className="text-xs text-slate-500">Last {window} GWs</span>}
      </div>

      <div className="flex items-center justify-between gap-2 mb-4">
        <div className="inline-flex rounded-lg border border-slate-700/60 bg-slate-800/60 p-0.5">
          {(
            [
              ['position', 'Points Leaders'],
              ['team', 'Team ICT Share'],
            ] as const
          ).map(([key, label]) => (
            <button
              key={key}
              onClick={() => setView(key)}
              className={`rounded-md px-3 py-1 text-xs font-medium transition-colors ${
                view === key
                  ? 'bg-slate-600/80 text-slate-100'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              {label}
            </button>
          ))}
        </div>

        {view === 'team' && (
          <div className="flex items-center gap-3">
            {ictLegend.map((l) => (
              <span key={l.label} className="flex items-center gap-1 text-[11px] text-slate-400">
                <span className={`h-2 w-2 rounded-full ${l.dot}`} />
                {l.label}
              </span>
            ))}
          </div>
        )}
      </div>

      {active.isLoading ? (
        <LoadingGrid />
      ) : active.isError || !active.data ? (
        <ErrorState
          message="Failed to load stats. Check your connection and try again."
          onRetry={() => active.refetch()}
        />
      ) : view === 'position' && stats.data ? (
        <div className="space-y-6">
          {stats.data.sections.map((section) => (
            <section key={section.position}>
              <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-400 mb-2">
                {section.label}
              </h2>
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {section.cards.map((card) => (
                  <StatCard key={`${section.position}-${card.component}`} card={card} />
                ))}
              </div>
            </section>
          ))}
        </div>
      ) : teamICT.data ? (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {teamICT.data.teams.map((entry) => (
            <TeamICTCard key={entry.team} entry={entry} />
          ))}
        </div>
      ) : null}
    </div>
  );
}
