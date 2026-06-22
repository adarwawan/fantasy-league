import type { FixtureOdds }  from '../../types/odds';
import { FixtureOddsCard }   from './FixtureOddsCard';

interface FixtureOddsGridProps {
  fixtures: FixtureOdds[];
}

export function FixtureOddsGrid({ fixtures }: FixtureOddsGridProps) {
  if (fixtures.length === 0) {
    return (
      <p className="text-slate-400 text-sm mt-8 text-center">
        Odds unavailable — no fixture data returned for this gameweek.
      </p>
    );
  }

  const byGW = fixtures.reduce<Map<number, FixtureOdds[]>>((acc, f) => {
    const list = acc.get(f.gw) ?? [];
    list.push(f);
    acc.set(f.gw, list);
    return acc;
  }, new Map());

  const sortedGWs = Array.from(byGW.keys()).sort((a, b) => a - b);

  return (
    <div className="space-y-6">
      {sortedGWs.map(gw => (
        <section key={gw}>
          <h2 className="text-sm font-semibold text-slate-400 uppercase tracking-wide mb-3">
            Gameweek {gw}
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {byGW.get(gw)!.map(f => (
              <FixtureOddsCard key={f.fixture_id || `${f.home_team}-${f.away_team}`} fixture={f} />
            ))}
          </div>
        </section>
      ))}
    </div>
  );
}
