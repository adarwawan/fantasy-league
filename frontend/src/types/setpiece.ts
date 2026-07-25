export type Role = 'taker' | 'target';
export type Duty = 'penalty' | 'dfk' | 'corner' | 'setpiece' | 'all';

export interface TakerRow {
  player_id:   string;
  player_name: string;
  rank:        number;
  attempts:    number;
  goals:       number;
  last_taken?: string;
  confidence:  number;
}

export interface TargetRow {
  player_id:      string;
  player_name:    string;
  rank:           number;
  shots:          number;
  goals:          number;
  xg:             number;
  header_pct?:    number;
  weighted_score: number;
  duty:           Duty;
  last_seen?:     string;
  confidence:     number;
}

export interface SetPieceTeam {
  team:   string;
  takers: {
    penalty: TakerRow[];
    dfk:     TakerRow[];
  };
  targets: TargetRow[]; // duty='all', league-comparable
  targets_by_duty: Partial<Record<Duty, TargetRow[]>>;
}

export interface SetPieceEvent {
  match_date:  string;
  minute:      number;
  role:        Role;
  duty:        Duty;
  player_name: string;
  is_header:   boolean;
  xg:          number;
}

export interface SetPieceTeamDetail extends SetPieceTeam {
  recent_events: SetPieceEvent[];
}

export interface SetPieceTeamsResponse {
  window_matches: number;
  updated_at?:    string;
  teams:          SetPieceTeam[];
}
