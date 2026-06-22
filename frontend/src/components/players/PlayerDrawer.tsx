import { useEffect } from 'react';
import type { Player } from '../../types/player';
import type { Team } from '../../types/team';
import { PositionBadge } from '../common/PositionBadge';
import { FixtureChip } from './FixtureChip';

const STATUS_DOT: Record<Player['status'], string> = {
  available: 'bg-green-500',
  doubt:     'bg-amber-400',
  injured:   'bg-red-500',
};

interface Props {
  player: Player | null;
  teams?: Team[];
  onClose: () => void;
}

export function PlayerDrawer({ player, teams, onClose }: Props) {
  useEffect(() => {
    if (!player) return;
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [player, onClose]);

  const diff = player ? player.top_n_ownership - player.global_ownership : 0;
  const diffColor = diff > 0 ? 'text-emerald-400' : diff < 0 ? 'text-red-400' : 'text-slate-400';
  const diffSign  = diff > 0 ? '+' : '';

  return (
    <>
      {/* Backdrop */}
      <div
        className={`fixed inset-0 z-40 bg-black/50 transition-opacity duration-200 ${player ? 'opacity-100' : 'opacity-0 pointer-events-none'}`}
        onClick={onClose}
      />

      {/* Drawer */}
      <aside
        className={`fixed top-0 right-0 z-50 h-full w-80 bg-slate-900 border-l border-slate-700 shadow-2xl flex flex-col transition-transform duration-250 ease-out ${player ? 'translate-x-0' : 'translate-x-full'}`}
        aria-label="Player details"
      >
        {player && (
          <>
            {/* Header */}
            <div className="flex items-start justify-between p-4 border-b border-slate-700">
              <div className="flex items-center gap-2">
                <span className={`w-2 h-2 rounded-full shrink-0 mt-1 ${STATUS_DOT[player.status]}`} title={player.news || player.status} />
                <div>
                  <p className="text-base font-semibold text-slate-100">{player.name}</p>
                  <p className="text-xs text-slate-400">{player.team.name}</p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <PositionBadge position={player.position} />
                <button
                  onClick={onClose}
                  className="text-slate-400 hover:text-slate-100 text-lg leading-none ml-1"
                  aria-label="Close"
                >
                  ✕
                </button>
              </div>
            </div>

            {/* Stats grid */}
            <div className="grid grid-cols-2 gap-px bg-slate-700 border-b border-slate-700">
              {[
                { label: 'Price',    value: `£${player.price.toFixed(1)}m` },
                { label: 'Form',     value: player.form.toFixed(1) },
                { label: 'Global %', value: `${player.global_ownership.toFixed(1)}%` },
                { label: 'Top-N %',  value: `${player.top_n_ownership.toFixed(1)}%` },
              ].map(({ label, value }) => (
                <div key={label} className="bg-slate-800/80 px-4 py-3">
                  <p className="text-[10px] uppercase tracking-wide text-slate-500">{label}</p>
                  <p className="text-sm font-semibold text-slate-100 tabular-nums">{value}</p>
                </div>
              ))}
            </div>

            {/* Differential */}
            <div className="px-4 py-3 border-b border-slate-700">
              <p className="text-[10px] uppercase tracking-wide text-slate-500 mb-1">Differential (Top-N − Global)</p>
              <p className={`text-lg font-bold tabular-nums ${diffColor}`}>{diffSign}{diff.toFixed(1)}%</p>
              {player.news && (
                <p className="text-xs text-amber-400 mt-2">{player.news}</p>
              )}
            </div>

            {/* Fixture run */}
            <div className="flex-1 overflow-y-auto px-4 py-3">
              <p className="text-[10px] uppercase tracking-wide text-slate-500 mb-3">Fixture run ({player.fixtures.length} GWs)</p>
              <div className="flex flex-col gap-2">
                {player.fixtures.map((f, i) => (
                  <div key={i} className="flex items-center justify-between">
                    <FixtureChip fixture={f} oppOvrForm={teams?.find(t => t.short_name === f.opp)?.ovr_form} />
                    <span className="text-xs text-slate-500">{new Date(f.kickoff).toLocaleDateString('en-GB', { day: 'numeric', month: 'short' })}</span>
                  </div>
                ))}
              </div>
            </div>
          </>
        )}
      </aside>
    </>
  );
}
