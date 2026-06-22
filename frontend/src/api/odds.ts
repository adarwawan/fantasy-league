import { apiFetch } from './client';
import type { FixtureOdds } from '../types/odds';

export function fetchFixtureOdds(game: string, gws?: number[]): Promise<FixtureOdds[]> {
  const qs = gws && gws.length > 0 ? `?gw=${gws.join(',')}` : '';
  return apiFetch<FixtureOdds[]>(`/api/${game}/fixtures/odds${qs}`);
}
