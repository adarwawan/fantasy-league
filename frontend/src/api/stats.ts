import { apiFetch } from './client';
import type { StatsResponse } from '../types/stats';

export function fetchStats(game: string): Promise<StatsResponse> {
  return apiFetch<StatsResponse>(`/api/${game}/stats`);
}
