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

export function GWContextBar({ game }: { game: string }) {
  const { gw, deadline, cachedAt, isLoading } = useGWContext(game);
  const countdown = useCountdown(deadline);

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

      {cachedAt && (
        <div className="flex items-center gap-1.5 ml-auto">
          <span className="text-slate-500 uppercase tracking-wide font-medium">Updated</span>
          <span className="text-slate-400 font-mono">{formatCachedAt(cachedAt)}</span>
        </div>
      )}
    </div>
  );
}
