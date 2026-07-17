import { apiFetch } from './client';
import type { StatsResponse, TeamICTResponse } from '../types/stats';

export function fetchStats(game: string): Promise<StatsResponse> {
  return apiFetch<StatsResponse>(`/api/${game}/stats`);
}

export function fetchTeamICT(game: string): Promise<TeamICTResponse> {
  return apiFetch<TeamICTResponse>(`/api/${game}/stats/teams`);
}
