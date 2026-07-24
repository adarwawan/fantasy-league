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

// Odds-based difficulty tiers reuse the FDR palette so odds chips and plain FDR
// chips read as one scheme: easy = FDR 2 green, medium = FDR 3 grey, hard = FDR 5 maroon.
const tierColours: Record<'easy' | 'medium' | 'hard', { bg: string; text: string }> = {
  easy:   fdrColours[2],
  medium: fdrColours[3],
  hard:   fdrColours[5],
};

// Opponent form is only trustworthy once enough of the season has been played;
// before this gameweek there isn't enough data, so we colour by static FDR.
const FORM_FALLBACK_FROM_GW = 5;

interface FixtureChipProps {
  fixture:      Fixture;
  xg?:          number | null;
  csPct?:       number | null;
  focusMode?:   FocusMode;
  /** Opponent's overall form, used as the first no-odds fallback (after GW 5). */
  oppOvrForm?:  number;
  /** Current season gameweek — gates the form fallback (needs > 5). */
  currentGw?:   number;
  /** Compact single-line variant: difficulty shown as a left accent, no xG/CS. */
  compact?:     boolean;
}

export function FixtureChip({ fixture, xg, csPct, focusMode = 'overall', oppOvrForm, currentGw, compact }: FixtureChipProps) {
  const resolvedXG    = xg   ?? null;
  const resolvedCSPct = csPct ?? null;

  // Colour resolution, in priority order:
  //   1. odds present            → chipTier (xG / CS based)
  //   2. opponent form, after    → ovrFormColour  (only once currentGw > 5)
  //      GW 5
  //   3. otherwise               → static FDR
  // Both the Players and Teams tabs feed the same inputs, so a given fixture
  // reads identically on both surfaces.
  let colours: { bg: string; text: string };
  if (resolvedXG !== null && resolvedCSPct !== null) {
    colours = tierColours[chipTier(resolvedXG, resolvedCSPct, fixture.difficulty, focusMode)];
  } else if (oppOvrForm !== undefined && currentGw !== undefined && currentGw > FORM_FALLBACK_FROM_GW) {
    colours = ovrFormColour(oppOvrForm);
  } else {
    colours = fdrColours[fixture.difficulty] ?? fdrColours[3];
  }

  const hasOdds = resolvedXG !== null && resolvedCSPct !== null;

  if (compact) {
    return (
      <span
        className={`inline-flex items-center rounded px-1.5 py-0.5 text-[11px] leading-tight ${colours.bg} ${colours.text}`}
        title={`GW${fixture.gw} · ${fixture.opp} ${fixture.ha === 'H' ? 'Home' : 'Away'}${hasOdds ? ` · xG ${resolvedXG!.toFixed(2)} · CS ${resolvedCSPct!.toFixed(0)}%` : ''}`}
      >
        <span className="font-bold">{fixture.opp} {fixture.ha}</span>
        <span className="ml-1 opacity-70">GW{fixture.gw}</span>
      </span>
    );
  }
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
