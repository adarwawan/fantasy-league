import { useCallback, useEffect, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { useScatter } from '../hooks/useScatter';
import { ScatterPlot } from '../components/scatter/ScatterPlot';
import { AxisSelector, type AxisKey } from '../components/scatter/AxisSelector';
import { PlayerDrawer } from '../components/players/PlayerDrawer';
import { ErrorState } from '../components/common/ErrorState';
import type { Player } from '../types/player';

type Position = 'GK' | 'DEF' | 'MID' | 'FWD';
const POSITIONS: Position[] = ['GK', 'DEF', 'MID', 'FWD'];

function getAxisParam(sp: URLSearchParams, key: string, fallback: AxisKey): AxisKey {
  const v = sp.get(key);
  const valid: AxisKey[] = ['global_ownership', 'top_n_ownership', 'form', 'avg_fdr'];
  return valid.includes(v as AxisKey) ? (v as AxisKey) : fallback;
}

export function ScatterPage() {
  const { game = 'fpl' } = useParams<{ game: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const [selectedPlayer, setSelectedPlayer] = useState<Player | null>(null);

  useEffect(() => {
    document.title = `${game.toUpperCase()} — Scatter Plot`;
  }, [game]);

  const xAxis    = getAxisParam(searchParams, 'x', 'global_ownership');
  const yAxis    = getAxisParam(searchParams, 'y', 'form');
  const pos      = searchParams.get('pos') as Position | null;
  const maxPrice = searchParams.has('max_price')
    ? parseFloat(searchParams.get('max_price')!)
    : 15.0;

  const { data, isLoading, isError, refetch } = useScatter(game);

  function set(updates: Record<string, string | undefined>) {
    const next = new URLSearchParams(searchParams);
    for (const [k, v] of Object.entries(updates)) {
      if (v === undefined) next.delete(k);
      else next.set(k, v);
    }
    setSearchParams(next, { replace: true });
  }

  const minPrice = searchParams.has('min_price') ? parseFloat(searchParams.get('min_price')!) : 4;

  const filtered: Player[] = (data?.players ?? []).filter(p => {
    if (pos && p.position !== pos) return false;
    if (p.price < minPrice) return false;
    if (p.price > maxPrice) return false;
    return true;
  });

  const handlePlayerClick = useCallback((p: Player) => setSelectedPlayer(p), []);
  const handleDrawerClose = useCallback(() => setSelectedPlayer(null), []);

  if (isLoading) return (
    <div>
      <h1 className="text-xl font-semibold text-slate-100 mb-4">{game.toUpperCase()} — Scatter Plot</h1>
      <div className="h-[480px] rounded-lg border border-slate-700/50 bg-slate-800/40 animate-pulse" />
    </div>
  );
  if (isError || !data) return (
    <ErrorState
      message="Failed to load scatter data. Check your connection and try again."
      onRetry={() => refetch()}
    />
  );

  return (
    <>
      <div>
        <h1 className="text-xl font-semibold text-slate-100 mb-4">
          {game.toUpperCase()} — Scatter Plot
        </h1>

        {/* Controls bar */}
        <div className="flex flex-wrap gap-4 items-end mb-6">
          {/* Axis selectors */}
          <div className="flex flex-col gap-2">
            <AxisSelector label="X" value={xAxis} onChange={v => set({ x: v })} />
            <AxisSelector label="Y" value={yAxis} onChange={v => set({ y: v })} />
          </div>

          <div className="h-10 border-l border-slate-700" />

          {/* Position filter */}
          <div>
            <label className="block text-xs text-slate-400 mb-1">Position</label>
            <div className="flex rounded-md border border-slate-600 overflow-hidden text-sm">
              <button
                className={`px-3 py-1.5 ${!pos ? 'bg-indigo-600 text-white' : 'bg-slate-700/50 text-slate-300 hover:bg-slate-600'}`}
                onClick={() => set({ pos: undefined })}
              >
                All
              </button>
              {POSITIONS.map(p => (
                <button
                  key={p}
                  className={`px-3 py-1.5 border-l border-slate-600 ${pos === p ? 'bg-indigo-600 text-white' : 'bg-slate-700/50 text-slate-300 hover:bg-slate-600'}`}
                  onClick={() => set({ pos: pos === p ? undefined : p })}
                >
                  {p}
                </button>
              ))}
            </div>
          </div>

          {/* Price range */}
          <div>
            <label className="block text-xs text-slate-400 mb-1">Price range (£m)</label>
            <div className="flex items-center gap-1.5">
              <input
                type="number"
                min={4}
                max={14.5}
                step={0.5}
                value={searchParams.has('min_price') ? parseFloat(searchParams.get('min_price')!) : 4}
                onChange={e => set({ min_price: e.target.value === '4' ? undefined : e.target.value })}
                aria-label="Minimum price"
                className="w-16 px-2 py-1.5 rounded-md bg-slate-700/50 border border-slate-600 text-sm text-slate-100 text-center tabular-nums focus:outline-none focus:ring-1 focus:ring-indigo-500"
              />
              <span className="text-slate-500 text-xs">–</span>
              <input
                type="number"
                min={4.5}
                max={15}
                step={0.5}
                value={maxPrice}
                onChange={e => set({ max_price: e.target.value === '15' ? undefined : e.target.value })}
                aria-label="Maximum price"
                className="w-16 px-2 py-1.5 rounded-md bg-slate-700/50 border border-slate-600 text-sm text-slate-100 text-center tabular-nums focus:outline-none focus:ring-1 focus:ring-indigo-500"
              />
            </div>
          </div>

          <div className="text-xs text-slate-500 self-end pb-1">
            {filtered.length} players · click dot to inspect
          </div>
        </div>

        <ScatterPlot
          players={filtered}
          xAxis={xAxis}
          yAxis={yAxis}
          onPlayerClick={handlePlayerClick}
        />
      </div>

      <PlayerDrawer player={selectedPlayer} onClose={handleDrawerClose} />
    </>
  );
}
