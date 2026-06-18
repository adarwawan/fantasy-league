export interface TeamFixture {
  gw:         number;
  opp:        string;
  ha:         'H' | 'A';
  difficulty: 1 | 2 | 3 | 4 | 5;
  kickoff:    string;
}

export interface Team {
  id:         string;
  game_id:    string;
  name:       string;
  short_name: string;
  att_form:   number;
  def_form:   number;
  ovr_form:   number;
  fixtures:   TeamFixture[];
}

export interface TeamsResponse {
  teams:   Team[];
  meta: {
    game_id:   string;
    cached_at: string;
  };
}
