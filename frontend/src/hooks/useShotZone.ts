import { useQuery } from '@tanstack/react-query';
import { fetchShotZoneTeam, fetchShotZoneTeams } from '../api/shotzone';

export function useShotZoneTeams() {
  return useQuery({
    queryKey:  ['shotzone', 'teams'],
    queryFn:   fetchShotZoneTeams,
    staleTime: 30 * 60 * 1000,
  });
}

export function useShotZoneTeam(team: string | undefined) {
  return useQuery({
    queryKey:  ['shotzone', 'team', team],
    queryFn:   () => fetchShotZoneTeam(team as string),
    enabled:   !!team,
    staleTime: 30 * 60 * 1000,
  });
}
