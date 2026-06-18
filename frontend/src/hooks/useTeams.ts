import { useQuery } from '@tanstack/react-query';
import { fetchTeams } from '../api/teams';

export function useTeams(game: string) {
  return useQuery({
    queryKey:  ['teams', game],
    queryFn:   () => fetchTeams(game),
    staleTime: 30 * 60 * 1000,
  });
}
