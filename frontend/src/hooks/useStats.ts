import { useQuery } from '@tanstack/react-query';
import { fetchStats } from '../api/stats';

export function useStats(game: string) {
  return useQuery({
    queryKey: ['stats', game],
    queryFn:  () => fetchStats(game),
    staleTime: 30 * 60 * 1000,
  });
}
