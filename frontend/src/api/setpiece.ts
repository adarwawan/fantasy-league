import { apiFetch } from './client';
import type { SetPieceTeamDetail, SetPieceTeamsResponse } from '../types/setpiece';

// The set-piece module is PL-wide and isolated from the per-game pipeline, so
// its routes live under /api/setpiece (no game param).
export function fetchSetPieceTeams(): Promise<SetPieceTeamsResponse> {
  return apiFetch<SetPieceTeamsResponse>('/api/setpiece/teams');
}

export function fetchSetPieceTeam(understatTeam: string): Promise<SetPieceTeamDetail> {
  return apiFetch<SetPieceTeamDetail>(`/api/setpiece/teams/${encodeURIComponent(understatTeam)}`);
}
