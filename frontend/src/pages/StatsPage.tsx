import { useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useStats } from '../hooks/useStats';
import { StatCard } from '../components/stats/StatCard';
import { ErrorState } from '../components/common/ErrorState';

export function StatsPage() {
  const { game = 'fpl' } = useParams<{ game: string }>();
  const { data, isLoading, isError, refetch } = useStats(game);

  useEffect(() => {
    document.title = `${game.toUpperCase()} — Stats`;
  }, [game]);

  if (isLoading) {
    return (
      <div>
        <h1 className="text-xl font-semibold text-slate-100 mb-4">{game.toUpperCase()} — Stats</h1>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="h-40 rounded-xl border border-slate-700/50 bg-slate-800/40 animate-pulse" />
          ))}
        </div>
      </div>
    );
  }

  if (isError || !data) {
    return (
      <ErrorState
        message="Failed to load stats. Check your connection and try again."
        onRetry={() => refetch()}
      />
    );
  }

  return (
    <div>
      <div className="flex items-baseline justify-between gap-2 mb-4">
        <h1 className="text-xl font-semibold text-slate-100">{game.toUpperCase()} — Stats</h1>
        <span className="text-xs text-slate-500">Last {data.meta.window} GWs</span>
      </div>

      <div className="space-y-6">
        {data.sections.map((section) => (
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
    </div>
  );
}
