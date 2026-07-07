import type { Player } from '../../types/player';

const POS_COLORS: Record<string, string> = {
  GK:  '#34d399',
  DEF: '#60a5fa',
  MID: '#a78bfa',
  FWD: '#f87171',
};

interface Props {
  player:      Player;
  avgFdr:      number;
  onDetails:   () => void;
  onClose:     () => void;
}

/**
 * Docked detail card for touch devices. Pinned to the bottom of the viewport
 * so it never covers the dot being inspected (unlike a floating tooltip).
 * Shows a quick peek; "Details" opens the full player drawer.
 */
export function PlayerPeekCard({ player, avgFdr, onDetails, onClose }: Props) {
  return (
    <div
      className="fixed inset-x-0 bottom-0 z-40 sm:hidden"
      style={{ paddingBottom: 'env(safe-area-inset-bottom)' }}
      role="dialog"
      aria-label={`${player.name} quick stats`}
    >
      <div className="mx-2 mb-2 rounded-xl border border-slate-700 bg-slate-800/95 backdrop-blur shadow-2xl p-3">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0">
            <div className="flex items-center gap-1.5">
              <span
                className="inline-block w-2.5 h-2.5 rounded-full shrink-0"
                style={{ background: POS_COLORS[player.position] }}
              />
              <span className="font-semibold text-slate-100 truncate">{player.name}</span>
              {player.must_have && (
                <span className="text-[10px] font-medium text-amber-400 bg-amber-400/15 rounded px-1 py-px shrink-0">
                  ★
                </span>
              )}
            </div>
            <div className="text-xs text-slate-400 mt-0.5">
              {player.team.short_name} · {player.position} · £{player.price.toFixed(1)}m
            </div>
          </div>
          <button
            onClick={onClose}
            aria-label="Close"
            className="shrink-0 -mt-1 -mr-1 p-1.5 text-slate-400 hover:text-slate-200 rounded-md"
          >
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
              <path d="M4 4l8 8M12 4l-8 8" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
            </svg>
          </button>
        </div>

        <div className="grid grid-cols-5 gap-2 mt-3 text-center">
          <Stat label="Form"   value={player.form.toFixed(1)} />
          <Stat label="G.own"  value={`${player.global_ownership.toFixed(0)}%`} />
          <Stat label="T-N"    value={`${player.top_n_ownership.toFixed(0)}%`} />
          <Stat label="EO"     value={`${player.effective_ownership.toFixed(0)}%`} />
          <Stat label="FDR"    value={avgFdr.toFixed(1)} />
        </div>

        <button
          onClick={onDetails}
          className="mt-3 w-full py-2 rounded-lg bg-indigo-600 text-white text-sm font-medium active:bg-indigo-700"
        >
          View details
        </button>
      </div>
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="text-sm font-semibold text-slate-100 tabular-nums">{value}</div>
      <div className="text-[10px] uppercase tracking-wide text-slate-500 mt-0.5">{label}</div>
    </div>
  );
}
