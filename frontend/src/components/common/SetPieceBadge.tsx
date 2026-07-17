import type { Player } from '../../types/player';

export type SetPieceDuty = 'penalties' | 'freekicks' | 'corners';

export const DUTY_META: Record<SetPieceDuty, { code: string; label: string; colorClasses: string }> = {
  penalties: { code: 'PEN', label: 'Penalty taker',                     colorClasses: 'text-amber-300 bg-amber-400/10' },
  freekicks: { code: 'FK',  label: 'Direct free-kick taker',           colorClasses: 'text-sky-300 bg-sky-400/10' },
  corners:   { code: 'CK',  label: 'Corner & indirect free-kick taker', colorClasses: 'text-violet-300 bg-violet-400/10' },
};

/** The player's duties in display order, only where a rank is assigned. */
export function setPieceDuties(player: Player): { duty: SetPieceDuty; order: number }[] {
  const orders: [SetPieceDuty, number | null][] = [
    ['penalties', player.penalties_order],
    ['freekicks', player.direct_freekicks_order],
    ['corners',   player.corners_indirect_freekicks_order],
  ];
  return orders
    .filter((entry): entry is [SetPieceDuty, number] => entry[1] != null)
    .map(([duty, order]) => ({ duty, order }));
}

/**
 * Set-piece duty pill (same idiom as the ICT star badges): a letter code for
 * the duty, plus the rank when the player is not first choice. Backup takers
 * are dimmed so first-choice duties stand out at a glance.
 */
export function SetPieceBadge({ duty, order }: { duty: SetPieceDuty; order: number }) {
  const { code, label, colorClasses } = DUTY_META[duty];
  return (
    <span
      title={`${label} #${order} in team`}
      className={`flex shrink-0 items-center rounded px-1 py-px text-[10px] font-bold tabular-nums ${colorClasses} ${order > 1 ? 'opacity-50' : ''}`}
    >
      {code}
      {order > 1 && order}
    </span>
  );
}

/** The player's set-piece badges, capped at maxOrder (e.g. 2 in the table row). */
export function SetPieceBadges({ player, maxOrder }: { player: Player; maxOrder?: number }) {
  const duties = setPieceDuties(player).filter(d => maxOrder == null || d.order <= maxOrder);
  if (duties.length === 0) return null;
  return (
    <span className="flex shrink-0 items-center gap-1">
      {duties.map(({ duty, order }) => (
        <SetPieceBadge key={duty} duty={duty} order={order} />
      ))}
    </span>
  );
}
