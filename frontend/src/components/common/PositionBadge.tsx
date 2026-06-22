import type { Player } from '../../types/player';

const STYLES: Record<Player['position'], string> = {
  GK:  'bg-emerald-400/15 text-emerald-400 ring-emerald-400/30',
  DEF: 'bg-blue-400/15 text-blue-400 ring-blue-400/30',
  MID: 'bg-purple-400/15 text-purple-400 ring-purple-400/30',
  FWD: 'bg-red-400/15 text-red-400 ring-red-400/30',
};

export function PositionBadge({ position }: { position: Player['position'] }) {
  return (
    <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-xs font-bold ring-1 ${STYLES[position]}`}>
      {position}
    </span>
  );
}
