export interface TeamFixture {
  gw:         number;
  opp:        string;
  ha:         'H' | 'A';
  difficulty: 1 | 2 | 3 | 4 | 5;
  kickoff:    string;
  xg:         number | null;
  cs_pct:     number | null;
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
  xg_sum:     number | null;
  cs_avg:     number | null;
}

export interface TeamsResponse {
  teams:   Team[];
  meta: {
    game_id:   string;
    cached_at: string;
  };
}
