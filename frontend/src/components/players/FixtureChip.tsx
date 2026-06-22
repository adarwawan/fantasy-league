import type { Fixture } from '../../types/player';
import { fdrColours } from '../../utils/fdr';

export function FixtureChip({ fixture }: { fixture: Fixture }) {
  const colours = fdrColours[fixture.difficulty] ?? fdrColours[3];
  return (
    <span
      className={`inline-flex flex-col items-center px-1.5 py-0.5 rounded text-xs font-medium leading-tight ${colours.bg} ${colours.text}`}
      title={`GW${fixture.gw} · ${fixture.opp} ${fixture.ha === 'H' ? 'Home' : 'Away'}`}
    >
      <span className="font-bold">{fixture.opp} {fixture.ha}</span>
      <span className="opacity-70 text-[10px]">GW{fixture.gw}</span>
    </span>
  );
}
