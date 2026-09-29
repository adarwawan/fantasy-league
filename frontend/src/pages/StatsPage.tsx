import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useStats, useTeamICT } from '../hooks/useStats';
import { StatCard } from '../components/stats/StatCard';
import { TeamICTCard } from '../components/stats/TeamICTCard';
import { QueryBoundary } from '../components/ui/QueryBoundary';
import { CardGridSkeleton } from '../components/ui/Skeletons';
import { SegmentedControl } from '../components/ui/SegmentedControl';
import { SectionHeading } from '../components/ui/SectionHeading';
import { PageHeader } from '../components/layout/PageHeader';
import { usePageTitle } from '../hooks/usePageTitle';

type StatsView = 'position' | 'team';

const ictLegend = [
  { label: 'Influence',  dot: 'bg-emerald-400' },
  { label: 'Creativity', dot: 'bg-sky-400' },
  { label: 'Threat',     dot: 'bg-rose-400' },
];

export function StatsPage() {
  const { game = 'fpl' } = useParams<{ game: string }>();
  const [view, setView] = useState<StatsView>('position');
  const stats = useStats(game);
  const teamICT = useTeamICT(game, view === 'team');

  usePageTitle('Stats');

  const active = view === 'position' ? stats : teamICT;
  const window = stats.data?.meta.window ?? teamICT.data?.meta.window;

  return (
    <div>
      <PageHeader
        title="Stats"
        description="Top players for each FPL scoring component, and how each team's Influence, Creativity and Threat is shared out."
        meta={window != null ? `Last ${window} gameweeks` : undefined}
      />

      <div className="flex items-center justify-between gap-2 mb-4">
        <SegmentedControl
          label="Stats view"
          value={view}
          onChange={setView}
          options={[
            { value: 'position', label: 'Points Leaders' },
            { value: 'team', label: 'Team ICT Share' },
          ]}
        />

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

      <QueryBoundary
        isLoading={active.isLoading}
        isError={active.isError || !active.data}
        onRetry={() => active.refetch()}
        what="stats"
        skeleton={<CardGridSkeleton />}
      >
        {view === 'position' && stats.data ? (
          <div className="space-y-6">
            {stats.data.sections.map((section) => (
              <section key={section.position}>
                <SectionHeading>{section.label}</SectionHeading>
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
      </QueryBoundary>
    </div>
  );
}
