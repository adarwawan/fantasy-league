export interface Fixture {
  gw:         number;
  opp:        string;
  ha:         'H' | 'A';
  difficulty: 1 | 2 | 3 | 4 | 5;
  kickoff:    string;
  xg:         number | null;
  cs_pct:     number | null;
}

export interface Player {
  id:                string;
  game_id:           string;
  name:              string;
  team: {
    id:         string;
    short_name: string;
    name:       string;
  };
  position:          'GK' | 'DEF' | 'MID' | 'FWD';
  price:             number;
  form:              number;
  global_ownership:  number;
  top_n_ownership:   number;
  effective_ownership: number;
  top_n_size:        number;
  status:            'available' | 'doubt' | 'injured';
  news:              string;
  // Set-piece taker ranks (1 = first choice), null when the player has no duty.
  penalties_order:                  number | null;
  direct_freekicks_order:           number | null;
  corners_indirect_freekicks_order: number | null;
  must_have:         boolean;
  fixtures:          Fixture[];
  recent_points:     GWPoints[];
  // Minutes-security signal over the recent window.
  recent_minutes:    GWMinutes[];
  start_rate:        number | null; // fixtures started ÷ fixtures the club played; null = no data
  avg_minutes:       number;
}

export interface GWPoints {
  gw:     number;
  points: number;
}

export interface GWMinutes {
  gw:       number;
  minutes:  number;
  starts:   number; // fixtures started that GW (0, or up to 2 in a double)
  fixtures: number; // fixtures the club played that GW (0 = blank, 2 = double)
}

export interface PlayersResponse {
  players: Player[];
  meta: {
    game_id:    string;
    gw:         number;
    top_n_size: number;
    cached_at:  string;
    total:      number;
  };
}
