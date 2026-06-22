import { useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useTeams } from '../hooks/useTeams';
import { usePlayers } from '../hooks/usePlayers';
import { TeamFormTable } from '../components/teams/TeamFormTable';
import { SkeletonRow } from '../components/common/SkeletonRow';

// expand · TEAM · ATT FORM · DEF FORM · OVR FORM · NEXT 5 GWS
const TEAM_SKELETON_COLS = ['w-6', 'w-28', 'w-16', 'w-16', 'w-16', 'w-40'];

export function TeamsPage() {
  const { game = 'fpl' } = useParams<{ game: string }>();

  useEffect(() => {
    document.title = `${game.toUpperCase()} — Teams`;
  }, [game]);

  const { data: teamsData, isLoading: teamsLoading, isError: teamsError } = useTeams(game);
  const { data: playersData } = usePlayers(game, {});

  if (teamsLoading) {
    return (
      <div>
        <h1 className="text-xl font-semibold text-slate-100 mb-4">{game.toUpperCase()} — Teams</h1>
        <div className="overflow-x-auto rounded-lg border border-slate-700/50">
          <table className="w-full text-sm">
            <tbody>
              {Array.from({ length: 8 }).map((_, i) => (
                <SkeletonRow key={i} cols={TEAM_SKELETON_COLS} />
              ))}
            </tbody>
          </table>
        </div>
      </div>
    );
  }

  if (teamsError || !teamsData) {
    return <div className="text-red-500 py-8 text-center">Failed to load teams.</div>;
  }

  return (
    <div>
      <h1 className="text-xl font-semibold text-slate-100 mb-4">
        {game.toUpperCase()} — Teams
      </h1>
      <TeamFormTable
        teams={teamsData.teams}
        players={playersData?.players ?? []}
      />
    </div>
  );
}
