import { useQuery } from '@tanstack/react-query';
import { fetchStats, fetchTeamICT } from '../api/stats';

export function useStats(game: string) {
  return useQuery({
    queryKey: ['stats', game],
    queryFn:  () => fetchStats(game),
    staleTime: 30 * 60 * 1000,
  });
}

export function useTeamICT(game: string, enabled: boolean) {
  return useQuery({
    queryKey: ['stats-team-ict', game],
    queryFn:  () => fetchTeamICT(game),
    staleTime: 30 * 60 * 1000,
    enabled,
  });
}
