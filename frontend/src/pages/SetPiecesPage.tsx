import { useEffect, useState } from 'react';
import { useSetPieceTeams } from '../hooks/useSetPiece';
import { ErrorState } from '../components/common/ErrorState';
import { TeamCard } from '../components/setpiece/TeamCard';

export function SetPiecesPage() {
  const { data, isLoading, isError, refetch } = useSetPieceTeams();
  const [query, setQuery] = useState('');

  useEffect(() => {
    document.title = 'Set Pieces — Observed';
  }, []);

  const teams = data?.teams ?? [];
  const filtered = query
    ? teams.filter((t) => t.team.toLowerCase().includes(query.toLowerCase()))
    : teams;

  return (
    <div>
      <div className="mb-4 flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold text-slate-100">Set Pieces</h1>
          <p className="text-sm text-slate-400 mt-1 max-w-2xl">
            Observed from Understat shot data — each team's real penalty / free-kick takers and
            set-piece target men, independent of FPL's declared order.
          </p>
        </div>
        {teams.length > 0 && (
          <input
            type="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Filter teams…"
            className="w-full sm:w-48 px-3 py-1.5 text-sm rounded-md bg-slate-900 border border-slate-700/60 text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-violet-500/50"
          />
        )}
      </div>

      {isLoading && <LoadingGrid />}

      {isError && (
        <ErrorState
          message="Failed to load set-piece data. The source may be temporarily unavailable."
          onRetry={() => refetch()}
        />
      )}

      {!isLoading && !isError && teams.length === 0 && <EmptyState />}

      {!isLoading && !isError && filtered.length > 0 && (
        <div className="grid gap-4 lg:grid-cols-2 items-start">
          {filtered.map((t) => (
            <TeamCard
              key={t.team}
              team={t}
              windowMatches={data?.window_matches ?? 6}
              updatedAt={data?.updated_at}
            />
          ))}
        </div>
      )}

      {!isLoading && !isError && teams.length > 0 && filtered.length === 0 && (
        <p className="text-sm text-slate-500 py-8 text-center">No team matches “{query}”.</p>
      )}
    </div>
  );
}

function LoadingGrid() {
  return (
    <div className="grid gap-4 lg:grid-cols-2">
      {Array.from({ length: 4 }).map((_, i) => (
        <div key={i} className="rounded-2xl border border-slate-700/60 bg-slate-900 p-6 h-72 animate-pulse">
          <div className="flex items-center gap-3 mb-6">
            <div className="h-11 w-11 rounded-xl bg-slate-700" />
            <div className="h-4 w-32 rounded bg-slate-700" />
          </div>
          <div className="grid grid-cols-2 gap-3 mb-6">
            <div className="h-20 rounded-xl bg-slate-800" />
            <div className="h-20 rounded-xl bg-slate-800" />
          </div>
          <div className="h-24 rounded-xl bg-slate-800" />
        </div>
      ))}
    </div>
  );
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center gap-2 py-16 text-center">
      <span className="text-3xl" aria-hidden="true">🎯</span>
      <p className="text-slate-300 text-sm">No set-piece data yet.</p>
      <p className="text-slate-500 text-xs">Signals appear once matches have been played and synced.</p>
    </div>
  );
}
