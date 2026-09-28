export type Zone = 'center' | 'left' | 'right' | 'outside';
export type ZoneTab = 'overall' | Zone;

export interface TakenRow {
  player_id:   string;
  player_name: string;
  rank:        number;
  shots:       number;
  goals:       number;
  xg:          number;
  last_seen?:  string;
}

export interface ConcededTotals {
  shots:      number;
  goals:      number;
  xg:         number;
  last_seen?: string;
}

export interface ZoneTakenSet {
  overall: TakenRow[];
  center:  TakenRow[];
  left:    TakenRow[];
  right:   TakenRow[];
  outside: TakenRow[];
}

export interface ZoneConcededSet {
  overall: ConcededTotals;
  center:  ConcededTotals;
  left:    ConcededTotals;
  right:   ConcededTotals;
  outside: ConcededTotals;
}

export interface ShotZoneTeamsResponse {
  window_matches: number;
  updated_at?:    string;
  teams:          string[];
}

export interface ShotZoneTeamDetail {
  team:           string;
  window_matches: number;
  updated_at?:    string;
  taken:          ZoneTakenSet;
  conceded:       ZoneConcededSet;
}
