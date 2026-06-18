import { useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useTeams } from '../hooks/useTeams';
import { usePlayers } from '../hooks/usePlayers';
import { TeamFormTable } from '../components/teams/TeamFormTable';

export function TeamsPage() {
  const { game = 'fpl' } = useParams<{ game: string }>();

  useEffect(() => {
    document.title = `${game.toUpperCase()} — Teams`;
  }, [game]);

  const { data: teamsData, isLoading: teamsLoading, isError: teamsError } = useTeams(game);
  const { data: playersData } = usePlayers(game, {});

  if (teamsLoading) {
    return <div className="text-gray-500 py-8 text-center">Loading teams…</div>;
  }

  if (teamsError || !teamsData) {
    return <div className="text-red-500 py-8 text-center">Failed to load teams.</div>;
  }

  return (
    <div>
      <h1 className="text-xl font-semibold text-gray-900 mb-4">
        {game.toUpperCase()} — Teams
      </h1>
      <TeamFormTable
        teams={teamsData.teams}
        players={playersData?.players ?? []}
      />
    </div>
  );
}
