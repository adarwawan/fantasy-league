export interface TrendsSession {
  gameweek:   number;
  started_at: string;
  ends_at:    string;
  active:     boolean;
  poll_count: number;
}

export interface LeaderRow {
  player_ext_id: number;
  name:          string;
  team:          string;
  position:      string;
  selected_pct:  number;
  now_cost:      number;
  net_transfers: number;
  // change over the window that the board ranks by. Unit depends on metric:
  // transfer count (transfers) or basis points of ownership (ownership).
  rank_delta:    number;
}

export type TrendsMetric = 'transfers' | 'ownership';

export interface LeadersResponse {
  active:    boolean;
  gameweek?: number;
  window?:   string;
  direction?: 'in' | 'out';
  metric?:   TrendsMetric;
  leaders:   LeaderRow[];
}

export interface SeriesPoint {
  captured_at:   string;
  selected_pct:  number;
  transfers_in:  number;
  transfers_out: number;
  net_transfers: number;
  now_cost:      number;
}

export interface SeriesResponse {
  active:        boolean;
  gameweek?:     number;
  player_ext_id?: number;
  series:        SeriesPoint[];
}

// SessionResponse is either an active session or {active:false}.
export type SessionResponse = TrendsSession | { active: false };
