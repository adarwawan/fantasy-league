import { findTeamMeta, readableText } from '../setpiece/teamMeta';

const SIZE = {
  xs: 'h-5 min-w-[2rem] px-1 rounded text-[10px]', // dense table rows / card meta
  sm: 'h-6 w-6 rounded-md text-[10px]',
  md: 'h-8 w-8 rounded-lg text-xs',
  lg: 'h-11 w-11 rounded-xl text-sm',
} as const;

const FALLBACK = '#475569'; // slate-600 — teams we hold no colour for (WCF/UCLF)

interface Props {
  /** Full team name (e.g. "Man City"). */
  name?: string;
  /** 3-letter code (e.g. "MCI"); used when only the code is known. */
  code?: string;
  size?: keyof typeof SIZE;
}

/**
 * Team chip: club colour + 3-letter code. The one way a team is shown outside
 * of prose. Unknown teams (other games) get a neutral chip with their code.
 */
export function TeamBadge({ name, code, size = 'md' }: Props) {
  const meta = findTeamMeta(name, code);
  const color = meta?.color ?? FALLBACK;
  const label = meta?.code ?? code ?? (name ?? '').slice(0, 3).toUpperCase();
  return (
    <span
      title={name ?? code}
      className={`inline-flex shrink-0 items-center justify-center font-bold tracking-wide ${SIZE[size]}`}
      style={{ backgroundColor: color, color: readableText(color) }}
    >
      {label}
    </span>
  );
}
