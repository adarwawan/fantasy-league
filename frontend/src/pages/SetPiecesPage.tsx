import { useState } from 'react';
import { useSetPieceTeams } from '../hooks/useSetPiece';
import { QueryBoundary } from '../components/ui/QueryBoundary';
import { EmptyState } from '../components/ui/EmptyState';
import { CardGridSkeleton } from '../components/ui/Skeletons';
import { PageHeader } from '../components/layout/PageHeader';
import { usePageTitle } from '../hooks/usePageTitle';
import { FilterBar } from '../components/ui/FilterBar';
import { Field } from '../components/ui/Field';
import { SearchInput } from '../components/ui/SearchInput';
import { TeamCard } from '../components/setpiece/TeamCard';

export function SetPiecesPage() {
  const { data, isLoading, isError, refetch } = useSetPieceTeams();
  const [query, setQuery] = useState('');

  usePageTitle('Set Pieces');

  const teams = data?.teams ?? [];
  const filtered = query
    ? teams.filter((t) => t.team.toLowerCase().includes(query.toLowerCase()))
    : teams;

  return (
    <div>
      <PageHeader
        title="Set Pieces"
        description="Observed from Understat shot data — each team's real penalty / free-kick takers and set-piece target men, independent of FPL's declared order."
        meta={data ? `Last ${data.window_matches} matches` : undefined}
      />

      {teams.length > 0 && (
        <FilterBar summary={`${filtered.length} of ${teams.length} teams`}>
          <Field label="Search team" htmlFor="setpiece-search">
            <SearchInput id="setpiece-search" value={query} onChange={setQuery} placeholder="e.g. Arsenal" />
          </Field>
        </FilterBar>
      )}

      <QueryBoundary
        isLoading={isLoading}
        isError={isError}
        isEmpty={teams.length === 0}
        onRetry={() => refetch()}
        what="set-piece data"
        emptyHint="Signals appear once matches have been played and synced."
        skeleton={<CardGridSkeleton count={4} height="h-72" className="grid gap-4 lg:grid-cols-2" />}
      >
        {filtered.length === 0 ? (
          <EmptyState what="teams" filtered onClear={() => setQuery('')} />
        ) : (
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
      </QueryBoundary>
    </div>
  );
}
