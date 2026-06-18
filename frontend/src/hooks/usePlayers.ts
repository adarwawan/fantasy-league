import { useQuery } from '@tanstack/react-query';
import { fetchPlayers, type PlayerQueryParams } from '../api/players';

export function usePlayers(game: string, params: PlayerQueryParams = {}) {
  return useQuery({
    queryKey: ['players', game, params],
    queryFn:  () => fetchPlayers(game, params),
    staleTime: 30 * 60 * 1000,
  });
}
