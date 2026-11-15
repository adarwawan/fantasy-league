import type { TeamBonusEntry } from '../types/stats';

export type BonusPosition = 'GK' | 'DEF' | 'MID' | 'FWD';
export const BONUS_POSITIONS: BonusPosition[] = ['GK', 'DEF', 'MID', 'FWD'];

/** Each match hands out 3+2+1 bonus points, so a club in a neutral match averages 3.0. */
export const NEUTRAL_BONUS_PER_MATCH = 3;

// Chip cut-offs on the top-3 earners' share of a club's bonus. Tuned on a
// 10-GW sample where clubs ranged 38%-73%; the middle band gets no chip.
const CONCENTRATED_MIN = 60;
const SPREAD_MAX = 45;
// Below this many club bonus points the shares are noise, so show no chip.
const MIN_BONUS_FOR_CHIP = 8;

/** Bonus per fixture played, so double and blank gameweeks stay correct. */
export function bonusPerMatch(e: TeamBonusEntry): number {
  return e.matches > 0 ? e.bonus / e.matches : 0;
}

/** Each position line's % of the club's bonus (0 when the club has none). */
export function positionShares(e: TeamBonusEntry): Record<BonusPosition, number> {
  const out = { GK: 0, DEF: 0, MID: 0, FWD: 0 };
  if (e.bonus <= 0) return out;
  for (const p of BONUS_POSITIONS) out[p] = (e.by_position[p] / e.bonus) * 100;
  return out;
}

/** % of the club's bonus earned by its (up to) three biggest earners. */
export function top3Share(e: TeamBonusEntry): number {
  if (e.bonus <= 0) return 0;
  return (e.top.reduce((s, p) => s + p.bonus, 0) / e.bonus) * 100;
}

export type ConcentrationChip = 'concentrated' | 'spread' | null;

export function concentrationChip(e: TeamBonusEntry): ConcentrationChip {
  if (e.bonus < MIN_BONUS_FOR_CHIP) return null;
  const share = top3Share(e);
  if (share >= CONCENTRATED_MIN) return 'concentrated';
  if (share <= SPREAD_MAX) return 'spread';
  return null;
}

/** Clubs sorted by bonus per match, highest first (ties: more bonus, then code). */
export function sortByBonusRate(teams: TeamBonusEntry[]): TeamBonusEntry[] {
  return [...teams].sort(
    (a, b) => bonusPerMatch(b) - bonusPerMatch(a) || b.bonus - a.bonus || a.team.localeCompare(b.team),
  );
}
