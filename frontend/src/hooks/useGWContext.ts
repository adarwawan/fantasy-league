import { useMemo } from 'react';
import { usePlayers } from './usePlayers';

export function useGWContext(game: string) {
  const { data, isLoading } = usePlayers(game, {});

  const deadline = useMemo(() => {
    const players = data?.players ?? [];
    const kickoffs = players
      .map(p => p.fixtures[0]?.kickoff)
      .filter(Boolean)
      .map(k => new Date(k!).getTime())
      .filter(t => t > Date.now());

    if (!kickoffs.length) return null;
    // FPL deadline = 90 min before first match of the GW
    return new Date(Math.min(...kickoffs) - 90 * 60 * 1000);
  }, [data]);

  return {
    gw: data?.meta.gw ?? null,
    cachedAt: data?.meta.cached_at ?? null,
    deadline,
    isLoading,
  };
}
