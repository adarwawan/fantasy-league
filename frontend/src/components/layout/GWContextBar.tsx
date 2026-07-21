import { useState, useEffect } from 'react';
import { useGWContext } from '../../hooks/useGWContext';

function useCountdown(target: Date | null): string {
  const [display, setDisplay] = useState('');

  useEffect(() => {
    if (!target) { setDisplay(''); return; }

    function update() {
      const diff = target!.getTime() - Date.now();
      if (diff <= 0) { setDisplay('Deadline passed'); return; }

      const h = Math.floor(diff / 3_600_000);
      const m = Math.floor((diff % 3_600_000) / 60_000);
      const s = Math.floor((diff % 60_000) / 1_000);

      if (h >= 48) {
        const d = Math.floor(h / 24);
        setDisplay(`${d}d ${h % 24}h`);
      } else if (h >= 1) {
        setDisplay(`${h}h ${m}m`);
      } else {
        setDisplay(`${m}m ${s}s`);
      }
    }

    update();
    const id = setInterval(update, 1_000);
    return () => clearInterval(id);
  }, [target]);

  return display;
}

function formatCachedAt(iso: string): string {
  return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

// Data-freshness thresholds. The sync cadence is well under an hour near
// deadlines, so anything past an hour is worth flagging and past three hours is
// a likely stale-sync incident — colour it loudly rather than trusting the number.
const STALE_AMBER_MS = 60 * 60 * 1000;
const STALE_RED_MS = 3 * 60 * 60 * 1000;

function freshness(cachedAt: string, now: number): { label: string; className: string } {
  const ageMs = now - new Date(cachedAt).getTime();
  const ageMin = Math.floor(ageMs / 60_000);

  let label: string;
  if (ageMin < 1) label = 'just now';
  else if (ageMin < 60) label = `${ageMin}m ago`;
  else label = `${Math.floor(ageMin / 60)}h ${ageMin % 60}m ago`;

  let className = 'text-slate-400';
  if (ageMs >= STALE_RED_MS) className = 'text-rose-400 font-semibold';
  else if (ageMs >= STALE_AMBER_MS) className = 'text-amber-400';

  return { label, className };
}

// useNow re-renders on an interval so the freshness label/colour advances even
// when no deadline countdown is driving updates.
function useNow(intervalMs: number): number {
  const [now, setNow] = useState(Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), intervalMs);
    return () => clearInterval(id);
  }, [intervalMs]);
  return now;
}

export function GWContextBar({ game }: { game: string }) {
  const { gw, deadline, cachedAt, isLoading } = useGWContext(game);
  const countdown = useCountdown(deadline);
  const now = useNow(30_000);

  if (isLoading || gw === null) return null;

  return (
    <div className="bg-slate-900/60 border-b border-slate-700/50 px-6 py-1.5 flex items-center gap-6 text-xs">
      <div className="flex items-center gap-1.5">
        <span className="text-slate-500 uppercase tracking-wide font-medium">GW</span>
        <span className="text-white font-bold text-sm">{gw}</span>
      </div>

      {countdown && (
        <div className="flex items-center gap-1.5">
          <span className="text-slate-500 uppercase tracking-wide font-medium">Deadline</span>
          <span className={`font-mono font-medium ${countdown.includes('m') && !countdown.includes('h') && !countdown.includes('d') ? 'text-amber-400' : 'text-emerald-400'}`}>
            {countdown}
          </span>
        </div>
      )}

      {cachedAt && (() => {
        const { label, className } = freshness(cachedAt, now);
        return (
          <div className="flex items-center gap-1.5 ml-auto">
            <span className="text-slate-500 uppercase tracking-wide font-medium">Updated</span>
            <span className={`font-mono ${className}`} title={`Last sync ${formatCachedAt(cachedAt)}`}>
              {label}
            </span>
          </div>
        );
      })()}
    </div>
  );
}
