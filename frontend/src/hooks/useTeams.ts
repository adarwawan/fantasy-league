import { useQuery } from '@tanstack/react-query';
import { fetchTeams } from '../api/teams';

export function useTeams(game: string, window = 5, sort = 'ovr_form') {
  return useQuery({
    queryKey:  ['teams', game, window, sort],
    queryFn:   () => fetchTeams(game, window, sort),
    staleTime: 30 * 60 * 1000,
  });
}
