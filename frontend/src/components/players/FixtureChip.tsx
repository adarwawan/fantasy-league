import type { Fixture } from '../../types/player';
import { fdrColours } from '../../utils/fdr';

export function FixtureChip({ fixture }: { fixture: Fixture }) {
  const colours = fdrColours[fixture.difficulty] ?? fdrColours[3];
  return (
    <span
      className={`inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium ${colours.bg} ${colours.text}`}
    >
      {fixture.opp} {fixture.ha}
    </span>
  );
}
