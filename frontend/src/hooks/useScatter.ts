import { useQuery } from '@tanstack/react-query';
import { apiFetch } from '../api/client';
import type { PlayersResponse } from '../types/player';

export function useScatter(game: string) {
  return useQuery({
    queryKey:  ['scatter', game],
    queryFn:   () => apiFetch<PlayersResponse>(`/api/${game}/players/scatter`),
    staleTime: 30 * 60 * 1000,
  });
}
