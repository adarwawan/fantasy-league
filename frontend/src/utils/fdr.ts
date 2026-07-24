// Fixture Difficulty Rating colours: a soft diverging green → grey → red scale.
// Lightness is symmetric (dark green → soft green → calm grey → soft rose → deep
// red), so the easy/hard poles carry the visual weight while the light tints stay
// gentle on the eyes. Each fill pairs with a same-family text colour for strong,
// glare-free contrast on both light tables and dark cards.
export const fdrColours: Record<number, { bg: string; text: string }> = {
  1: { bg: 'bg-[#157f3c]', text: 'text-white'     },
  2: { bg: 'bg-[#7ed3a0]', text: 'text-[#0f3d24]' },
  3: { bg: 'bg-[#e6e8ec]', text: 'text-[#3a4250]' },
  4: { bg: 'bg-[#f3a0aa]', text: 'text-[#5c111c]' },
  5: { bg: 'bg-[#a5163a]', text: 'text-white'     },
};

export function fdrLabel(difficulty: number): string {
  const labels: Record<number, string> = { 1: 'Very Easy', 2: 'Easy', 3: 'Medium', 4: 'Hard', 5: 'Very Hard' };
  return labels[difficulty] ?? 'Unknown';
}

/**
 * Map an opponent's ovr_form to chip colours: high form = tough opponent = red.
 * Reuses the FDR palette so form-based chips sit on the same scale as plain FDR
 * chips. Only meaningful once enough of the season has been played (see the
 * GW-5 gate in FixtureChip) — early on there isn't enough form to trust.
 */
export function ovrFormColour(ovr: number): { bg: string; text: string } {
  if (ovr >= 2.5) return fdrColours[5]; // strong opponent = hard
  if (ovr >= 1.5) return fdrColours[3]; // average opponent = neutral
  return fdrColours[2];                 // weak opponent = easy
}
