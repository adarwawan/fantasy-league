import type { Player } from '../types/player';

// Minutes security answers "will this player actually be on the pitch?" from
// recent start history, with the current availability status taking precedence:
// an injured or doubtful player is a risk regardless of how nailed-on he was.
//
// The start rate is fixture-level (see the backend), so it already treats double
// and blank gameweeks correctly. Cutoffs: >= 80% nailed (tolerates one rotation
// or absence in five), >= 50% rotation, below that bench risk.
export type MinutesTone = 'nailed' | 'rotation' | 'bench' | 'doubt' | 'out' | 'unknown';

export interface MinutesSecurity {
  tone:  MinutesTone;
  label: string;
  bg:    string;
  text:  string;
  title: string;
}

const NAILED_CUTOFF   = 0.8;
const ROTATION_CUTOFF = 0.5;

const TONE_STYLES: Record<MinutesTone, { bg: string; text: string }> = {
  nailed:   { bg: 'bg-emerald-500/15', text: 'text-emerald-300' },
  rotation: { bg: 'bg-amber-500/15',   text: 'text-amber-300'   },
  bench:    { bg: 'bg-red-500/15',     text: 'text-red-300'     },
  doubt:    { bg: 'bg-amber-500/15',   text: 'text-amber-300'   },
  out:      { bg: 'bg-slate-600/40',   text: 'text-slate-300'   },
  unknown:  { bg: 'bg-slate-700/40',   text: 'text-slate-500'   },
};

export function minutesSecurity(player: Player): MinutesSecurity {
  const pct = player.start_rate === null ? null : Math.round(player.start_rate * 100);
  const startsNote = pct === null ? 'no recent minutes data' : `${pct}% of starts`;

  const build = (tone: MinutesTone, label: string, title: string): MinutesSecurity => ({
    tone, label, title, ...TONE_STYLES[tone],
  });

  // Availability status overrides start history.
  if (player.status === 'injured') return build('out',   'Out',   player.news || 'Unavailable');
  if (player.status === 'doubt')   return build('doubt', 'Doubt', player.news || 'Fitness doubt');

  if (pct === null)              return build('unknown',  '—',          startsNote);
  if (pct >= NAILED_CUTOFF * 100)   return build('nailed',   'Nailed',     startsNote);
  if (pct >= ROTATION_CUTOFF * 100) return build('rotation', 'Rotation',   startsNote);
  return build('bench', 'Bench risk', startsNote);
}
