import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useTeams } from '../hooks/useTeams';
import { usePlayers } from '../hooks/usePlayers';
import { TeamFormTable } from '../components/teams/TeamFormTable';
import { QueryBoundary } from '../components/ui/QueryBoundary';
import { TableSkeleton } from '../components/ui/Skeletons';
import { PageHeader } from '../components/layout/PageHeader';
import { usePageTitle } from '../hooks/usePageTitle';
import type { FocusMode } from '../components/players/FixtureChip';

const TEAM_SKELETON_COLS = ['w-6', 'w-28', 'w-16', 'w-16', 'w-16', 'w-40'];

const TEAMS_DESCRIPTION =
  'Attack, defence and overall form for every team over a rolling window, with upcoming fixture difficulty.';

const focusToSort: Record<FocusMode, string> = {
  attack:  'xg_sum',
  defense: 'cs_avg',
  overall: 'ovr_form',
};

export function TeamsPage() {
  const { game = 'fpl' } = useParams<{ game: string }>();
  const [focusMode, setFocusMode] = useState<FocusMode>('overall');
  const [window, setWindow]       = useState(5);

  usePageTitle('Teams');

  const sort = focusToSort[focusMode];
  const { data: teamsData, isLoading: teamsLoading, isError: teamsError, refetch } = useTeams(game, window, sort);
  const { data: playersData } = usePlayers(game, {});

  return (
    <div>
      <PageHeader title="Teams" description={TEAMS_DESCRIPTION} />
      <QueryBoundary
        isLoading={teamsLoading}
        isError={teamsError || !teamsData}
        onRetry={() => refetch()}
        what="teams"
        skeleton={<TableSkeleton cols={TEAM_SKELETON_COLS} label="Loading teams" />}
      >
        {teamsData && (
          <TeamFormTable
            teams={teamsData.teams}
            players={playersData?.players ?? []}
            focusMode={focusMode}
            window={window}
            currentGw={playersData?.meta.gw}
            onFocusChange={setFocusMode}
            onWindowChange={setWindow}
          />
        )}
      </QueryBoundary>
    </div>
  );
}
