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
  must_have:         boolean;
  fixtures:          Fixture[];
  recent_points:     GWPoints[];
}

export interface GWPoints {
  gw:     number;
  points: number;
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
