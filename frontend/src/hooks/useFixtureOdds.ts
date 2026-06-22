import { useQuery } from '@tanstack/react-query';
import { fetchFixtureOdds } from '../api/odds';

export function useFixtureOdds(game: string, currentGW: number | null) {
  const gws = currentGW != null ? [currentGW, currentGW + 1] : undefined;
  return useQuery({
    queryKey: ['fixture-odds', game, gws],
    queryFn:  () => fetchFixtureOdds(game, gws),
    staleTime: 15 * 60 * 1000,
    enabled:  currentGW != null,
  });
}
