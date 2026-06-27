import type { Fixture } from '../../types/player';
import { fdrColours, ovrFormColour } from '../../utils/fdr';

export type FocusMode = 'attack' | 'defense' | 'overall';

function chipTier(
  xg: number | null,
  csPct: number | null,
  difficulty: number,
  focusMode: FocusMode,
): 'easy' | 'medium' | 'hard' {
  if (xg === null || csPct === null) {
    // FDR fallback
    if (difficulty <= 2) return 'easy';
    if (difficulty === 3) return 'medium';
    return 'hard';
  }
  if (focusMode === 'attack') {
    if (xg >= 2.0) return 'easy';
    if (xg >= 1.2) return 'medium';
    return 'hard';
  }
  if (focusMode === 'defense') {
    if (csPct >= 50) return 'easy';
    if (csPct >= 25) return 'medium';
    return 'hard';
  }
  // overall: normalised combined score (cs_pct is 0–100, cap at 50)
  const score = (xg / 3.0) * 0.5 + (csPct / 50) * 0.5;
  if (score >= 0.6) return 'easy';
  if (score >= 0.35) return 'medium';
  return 'hard';
}

const tierColours: Record<'easy' | 'medium' | 'hard', { bg: string; text: string }> = {
  easy:   { bg: 'bg-green-200',  text: 'text-green-900'  },
  medium: { bg: 'bg-amber-100',  text: 'text-amber-900'  },
  hard:   { bg: 'bg-red-200',    text: 'text-red-900'    },
};

interface FixtureChipProps {
  fixture:      Fixture;
  xg?:          number | null;
  csPct?:       number | null;
  focusMode?:   FocusMode;
  oppOvrForm?:  number;
}

export function FixtureChip({ fixture, xg, csPct, focusMode = 'overall', oppOvrForm }: FixtureChipProps) {
  const resolvedXG    = xg   ?? null;
  const resolvedCSPct = csPct ?? null;

  // When odds are available use the chipTier logic; otherwise fall back to
  // opponent form (if provided) or raw FDR.
  let colours: { bg: string; text: string };
  if (resolvedXG !== null && resolvedCSPct !== null) {
    colours = tierColours[chipTier(resolvedXG, resolvedCSPct, fixture.difficulty, focusMode)];
  } else if (oppOvrForm !== undefined) {
    colours = ovrFormColour(oppOvrForm);
  } else {
    colours = fdrColours[fixture.difficulty] ?? fdrColours[3];
  }

  const hasOdds = resolvedXG !== null && resolvedCSPct !== null;
  const xgDim   = hasOdds && focusMode === 'defense';
  const csDim   = hasOdds && focusMode === 'attack';

  return (
    <span
      className={`inline-flex flex-col items-center px-1.5 py-0.5 rounded text-xs font-medium leading-tight ${colours.bg} ${colours.text}`}
      title={`GW${fixture.gw} · ${fixture.opp} ${fixture.ha === 'H' ? 'Home' : 'Away'}${hasOdds ? ` · xG ${resolvedXG!.toFixed(2)} · CS ${resolvedCSPct!.toFixed(0)}%` : ''}`}
    >
      <span className="font-bold">{fixture.opp} {fixture.ha}</span>
      <span className="opacity-70 text-[10px]">GW{fixture.gw}</span>
      {hasOdds && (
        <span className="flex gap-1 text-[9px] mt-0.5">
          <span className={xgDim ? 'opacity-40' : 'font-semibold'}>
            {resolvedXG!.toFixed(2)}xG
          </span>
          <span className={csDim ? 'opacity-40' : 'font-semibold'}>
            {resolvedCSPct!.toFixed(0)}%CS
          </span>
        </span>
      )}
    </span>
  );
}
