export interface Fixture {
  gw:         number;
  opp:        string;
  ha:         'H' | 'A';
  difficulty: 1 | 2 | 3 | 4 | 5;
  kickoff:    string;
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
  top_n_size:        number;
  status:            'available' | 'doubt' | 'injured';
  news:              string;
  fixtures:          Fixture[];
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
