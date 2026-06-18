import { apiFetch } from './client';
import type { FixturesResponse } from '../types/fixture';

export function fetchFixtures(game: string, fromGw?: number, toGw?: number): Promise<FixturesResponse> {
  const qs = new URLSearchParams();
  if (fromGw != null) qs.set('from_gw', String(fromGw));
  if (toGw   != null) qs.set('to_gw',   String(toGw));
  const query = qs.toString();
  return apiFetch<FixturesResponse>(`/api/${game}/fixtures${query ? `?${query}` : ''}`);
}
