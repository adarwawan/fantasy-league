import { useQuery } from '@tanstack/react-query';
import { fetchTrendsLeaders, fetchTrendsSeries, fetchTrendsSession } from '../api/trends';

// Near a deadline the data moves every few minutes, so these refetch on an
// interval rather than sitting stale — the server's short cache TTL absorbs it.
export function useTrendsSession() {
  return useQuery({
    queryKey: ['trends', 'session'],
    queryFn:  fetchTrendsSession,
    refetchInterval: 60 * 1000,
  });
}

export function useTrendsLeaders(window: string, direction: 'in' | 'out' = 'in', limit = 25) {
  return useQuery({
    queryKey: ['trends', 'leaders', window, direction, limit],
    queryFn:  () => fetchTrendsLeaders(window, direction, limit),
    refetchInterval: 60 * 1000,
  });
}

export function useTrendsSeries(playerExtId: number | null) {
  return useQuery({
    queryKey: ['trends', 'series', playerExtId],
    queryFn:  () => fetchTrendsSeries(playerExtId as number),
    enabled:  playerExtId != null,
    staleTime: 30 * 1000,
  });
}
