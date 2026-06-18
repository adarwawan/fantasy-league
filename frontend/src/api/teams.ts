import { apiFetch } from './client';
import type { TeamsResponse } from '../types/team';

export function fetchTeams(game: string): Promise<TeamsResponse> {
  return apiFetch<TeamsResponse>(`/api/${game}/teams`);
}
