export interface FixtureOdds {
  fixture_id:   string;
  gw:           number;
  home_team:    string;
  home_xg:      number;
  home_cs_pct:  number;
  away_team:    string;
  away_xg:      number;
  away_cs_pct:  number;
  kickoff_time: string; // ISO 8601
}
