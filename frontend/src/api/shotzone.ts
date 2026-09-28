import { apiFetch } from './client';
import type { ShotZoneTeamDetail, ShotZoneTeamsResponse } from '../types/shotzone';

// The shot-zone module is PL-wide and isolated from the per-game pipeline
// (open-play companion to the set-piece board), so its routes live under
// /api/shot-zones (no game param) — same pattern as api/setpiece.ts.
export function fetchShotZoneTeams(): Promise<ShotZoneTeamsResponse> {
  return apiFetch<ShotZoneTeamsResponse>('/api/shot-zones/teams');
}

export function fetchShotZoneTeam(team: string): Promise<ShotZoneTeamDetail> {
  return apiFetch<ShotZoneTeamDetail>(`/api/shot-zones/teams/${encodeURIComponent(team)}`);
}
