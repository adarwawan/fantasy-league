import { apiFetch } from './client';

export interface EntryPick {
  player_id:  string;
  is_captain: boolean;
  multiplier: number;
}

export interface EntryResponse {
  entry_id:   string;
  team_value: number;
  bank:       number;
  gw:         number;
  picks:      EntryPick[];
}

// fetchEntry loads a manager's current squad + budget to seed the planner.
// FPL-only; other games return 404.
export function fetchEntry(game: string, entryId: string): Promise<EntryResponse> {
  return apiFetch<EntryResponse>(`/api/${game}/entry/${encodeURIComponent(entryId)}`);
}
