import type { FixtureOdds } from '../../types/odds';
import { FixtureOddsRow }   from './FixtureOddsRow';

interface FixtureOddsCardProps {
  fixture: FixtureOdds;
}

function formatKickoff(iso: string): { day: string; date: string } {
  const d = new Date(iso);
  return {
    day:  d.toLocaleDateString('en-GB', { weekday: 'short' }),
    date: d.toLocaleDateString('en-GB', { day: '2-digit', month: 'short' }),
  };
}

export function FixtureOddsCard({ fixture }: FixtureOddsCardProps) {
  const { day, date } = formatKickoff(fixture.kickoff_time);

  const homeXgWins = fixture.home_xg  >= fixture.away_xg;
  const homeCSWins = fixture.home_cs_pct >= fixture.away_cs_pct;

  return (
    <div className="bg-slate-800 border border-slate-700/60 rounded-lg overflow-hidden">
      {/* Header: GW label + kickoff */}
      <div className="flex items-center justify-between px-3 py-1.5 bg-slate-700/40 border-b border-slate-700/60">
        <span className="text-xs text-slate-400 font-medium">GW {fixture.gw}</span>
        <span className="text-xs text-slate-400">{day} {date}</span>
        {/* Column headers */}
        <div className="flex gap-3 text-xs text-slate-500 font-medium">
          <span className="min-w-[3rem] text-center">xG</span>
          <span className="min-w-[3rem] text-center">CS%</span>
        </div>
      </div>

      <FixtureOddsRow
        teamName={fixture.home_team}
        xg={fixture.home_xg}
        csPct={fixture.home_cs_pct}
        xgHighlight={homeXgWins}
        csHighlight={homeCSWins}
      />
      <div className="border-t border-slate-700/40" />
      <FixtureOddsRow
        teamName={fixture.away_team}
        xg={fixture.away_xg}
        csPct={fixture.away_cs_pct}
        xgHighlight={!homeXgWins}
        csHighlight={!homeCSWins}
      />
    </div>
  );
}
