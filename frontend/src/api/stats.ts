import { apiFetch } from './client';
import type { StatsResponse, TeamBonusResponse, TeamICTResponse } from '../types/stats';

export function fetchStats(game: string): Promise<StatsResponse> {
  return apiFetch<StatsResponse>(`/api/${game}/stats`);
}

export function fetchTeamICT(game: string): Promise<TeamICTResponse> {
  return apiFetch<TeamICTResponse>(`/api/${game}/stats/teams`);
}

export function fetchTeamBonus(game: string): Promise<TeamBonusResponse> {
  return apiFetch<TeamBonusResponse>(`/api/${game}/stats/bonus`);
}
