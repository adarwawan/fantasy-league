import { apiFetch } from './client';
import type { TeamsResponse } from '../types/team';

export function fetchTeams(game: string, window = 5, sort = 'ovr_form'): Promise<TeamsResponse> {
  return apiFetch<TeamsResponse>(`/api/${game}/teams?window=${window}&sort=${sort}`);
}
