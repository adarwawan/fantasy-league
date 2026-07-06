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
  meta: {
    game_id:   string;
    gw:        number;
    window:    number;
    cached_at: string;
  };
}
