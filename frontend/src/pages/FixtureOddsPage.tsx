import { useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useFixtureOdds } from '../hooks/useFixtureOdds';
import { useGWContext }    from '../hooks/useGWContext';
import { FixtureOddsGrid } from '../components/odds/FixtureOddsGrid';
import { ErrorState }      from '../components/common/ErrorState';

const ODDS_ENABLED = 'true';

const SKELETON_CARDS = Array.from({ length: 6 });

function OddsSkeleton() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
      {SKELETON_CARDS.map((_, i) => (
        <div key={i} className="bg-slate-800 border border-slate-700/60 rounded-lg overflow-hidden animate-pulse">
          <div className="h-7 bg-slate-700/40 border-b border-slate-700/60" />
          <div className="px-3 py-3 flex items-center gap-3">
            <div className="flex-1 h-3.5 bg-slate-700 rounded w-28" />
            <div className="h-3.5 w-12 bg-slate-700 rounded" />
            <div className="h-3.5 w-12 bg-slate-700 rounded" />
          </div>
          <div className="border-t border-slate-700/40" />
          <div className="px-3 py-3 flex items-center gap-3">
            <div className="flex-1 h-3.5 bg-slate-700 rounded w-24" />
            <div className="h-3.5 w-12 bg-slate-700 rounded" />
            <div className="h-3.5 w-12 bg-slate-700 rounded" />
          </div>
        </div>
      ))}
    </div>
  );
}

export function FixtureOddsPage() {
  const { game = 'fpl' } = useParams<{ game: string }>();

  useEffect(() => {
    document.title = `${game.toUpperCase()} — Odds`;
  }, [game]);

  if (!ODDS_ENABLED) {
    return (
      <div className="text-slate-400 text-sm mt-8 text-center">
        Fixture odds are not currently enabled.
      </div>
    );
  }

  return <FixtureOddsPageInner game={game} />;
}

function FixtureOddsPageInner({ game }: { game: string }) {
  const { gw: currentGW } = useGWContext(game);
  const { data, isLoading, isError, refetch } = useFixtureOdds(game, currentGW);

  if (isLoading) {
    return (
      <div>
        <h1 className="text-xl font-semibold text-slate-100 mb-4">{game.toUpperCase()} — Odds</h1>
        <OddsSkeleton />
      </div>
    );
  }

  if (isError) {
    return (
      <ErrorState
        message="Failed to load fixture odds. Check your connection and try again."
        onRetry={() => refetch()}
      />
    );
  }

  return (
    <div>
      <h1 className="text-xl font-semibold text-slate-100 mb-4">{game.toUpperCase()} — Odds</h1>
      <FixtureOddsGrid fixtures={data ?? []} />
    </div>
  );
}
