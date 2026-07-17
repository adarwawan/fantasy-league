export interface StatLeader {
  rank:  number;
  id:    string;
  name:  string;
  team:  string;
  value: number;
}

export interface StatCard {
  component: string;
  label:     string;
  points:    string;
  leaders:   StatLeader[];
}

export interface StatSection {
  position: 'GK' | 'DEF' | 'MID' | 'FWD';
  label:    string;
  cards:    StatCard[];
}

export interface StatsResponse {
  sections: StatSection[];
  meta: StatsMeta;
}

export interface StatsMeta {
  game_id:   string;
  gw:        number;
  window:    number;
  cached_at: string;
}

export interface TeamICTPlayer {
  id:         string;
  name:       string;
  position:   'GK' | 'DEF' | 'MID' | 'FWD';
  influence:  number;
  creativity: number;
  threat:     number;
  ict:        number;
  share:      number; // % of team total ICT over the window
  // Top-3 ranks within the team per component; absent outside the top 3.
  influence_rank?:  1 | 2 | 3;
  creativity_rank?: 1 | 2 | 3;
  threat_rank?:     1 | 2 | 3;
}

export interface TeamICTEntry {
  team:      string;
  total_ict: number;
  players:   TeamICTPlayer[];
}

export interface TeamICTResponse {
  teams: TeamICTEntry[];
  meta:  StatsMeta;
}
