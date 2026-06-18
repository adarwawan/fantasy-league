export interface Fixture {
  id:              string;
  game_id:         string;
  gw:              number;
  home_team_id:    string;
  away_team_id:    string;
  home_difficulty: number;
  away_difficulty: number;
  kickoff_time:    string;
  finished:        boolean;
}

export interface FixturesResponse {
  fixtures: Fixture[];
  meta: {
    game_id:   string;
    from_gw:   number;
    to_gw:     number;
  };
}
