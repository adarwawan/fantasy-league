import { GoalsCell } from './GoalsCell';
import { CSCell }    from './CSCell';

interface FixtureOddsRowProps {
  teamName:       string;
  xg:             number;
  csPct:          number;
  xgHighlight:    boolean;
  csHighlight:    boolean;
}

export function FixtureOddsRow({ teamName, xg, csPct, xgHighlight, csHighlight }: FixtureOddsRowProps) {
  return (
    <div className="flex items-center gap-3 px-3 py-2">
      <span className="flex-1 text-sm font-medium text-slate-100 truncate">{teamName}</span>
      <GoalsCell value={xg}    highlight={xgHighlight} />
      <CSCell    value={csPct} highlight={csHighlight} />
    </div>
  );
}
