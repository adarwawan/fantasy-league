import { useCallback, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { useScatter } from '../../hooks/useScatter';
import { ScatterPlot } from './ScatterPlot';
import { AxisSelector, type AxisKey } from './AxisSelector';
import { PlayerDrawer } from '../players/PlayerDrawer';
import { FilterBar } from '../ui/FilterBar';
import { Field } from '../ui/Field';
import { PositionFilter, type Position } from '../ui/PositionFilter';
import { PriceRange } from '../ui/PriceRange';
import { ErrorState } from '../ui/ErrorState';
import { PanelSkeleton } from '../ui/Skeletons';
import { priceCeiling, priceFloor } from '../../utils/price';
import type { Player } from '../../types/player';


// Thresholds below which a player is "fringe" (only hidden when BOTH are met).
const FORM_FLOOR = 2.5;
const OWNERSHIP_FLOOR = 2; // percent global ownership

function getAxisParam(sp: URLSearchParams, key: string, fallback: AxisKey): AxisKey {
  const v = sp.get(key);
  const valid: AxisKey[] = ['global_ownership', 'top_n_ownership', 'effective_ownership', 'form', 'avg_fdr'];
  return valid.includes(v as AxisKey) ? (v as AxisKey) : fallback;
}

/**
 * The scatter-plot view of the player pool. Lives inside the Players page as
 * an alternate "Plot" view (toggled via the `view` search param), so it owns
 * its own data fetch (useScatter), loading/error states, and inspect drawer.
 */
export function ScatterView() {
  const { game = 'fpl' } = useParams<{ game: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const [selectedPlayer, setSelectedPlayer] = useState<Player | null>(null);

  const { data, isLoading, isError, refetch } = useScatter(game);

  // Bounds track the cheapest/most expensive loaded player, so in-season price
  // moves stay reachable; they are also the default (unfiltered) min/max.
  const priceMin = priceFloor(data?.players);
  const priceMax = priceCeiling(data?.players);

  const xAxis    = getAxisParam(searchParams, 'x', 'global_ownership');
  const yAxis    = getAxisParam(searchParams, 'y', 'form');
  const pos      = searchParams.get('pos') as Position | null;
  const maxPrice = searchParams.has('max_price')
    ? parseFloat(searchParams.get('max_price')!)
    : priceMax;

  function set(updates: Record<string, string | undefined>) {
    const next = new URLSearchParams(searchParams);
    for (const [k, v] of Object.entries(updates)) {
      if (v === undefined) next.delete(k);
      else next.set(k, v);
    }
    setSearchParams(next, { replace: true });
  }

  const minPrice = searchParams.has('min_price') ? parseFloat(searchParams.get('min_price')!) : priceMin;

  // "Fringe" = barely owned AND out of form — noise on the plot. Hidden by
  // default (must-haves are always kept). Toggle with the `all` param.
  const showAll = searchParams.get('all') === '1';

  const priceScoped: Player[] = (data?.players ?? []).filter(p => {
    if (pos && p.position !== pos) return false;
    if (p.price < minPrice) return false;
    if (p.price > maxPrice) return false;
    return true;
  });

  const isFringe = (p: Player): boolean =>
    !p.must_have && p.form < FORM_FLOOR && p.global_ownership < OWNERSHIP_FLOOR;

  const filtered: Player[] = showAll ? priceScoped : priceScoped.filter(p => !isFringe(p));
  const hiddenCount = priceScoped.length - filtered.length;

  const handlePlayerClick = useCallback((p: Player) => setSelectedPlayer(p), []);
  const handleDrawerClose = useCallback(() => setSelectedPlayer(null), []);

  if (isLoading) return (
    <PanelSkeleton height="h-[480px]" />
  );
  if (isError || !data) return (
    <ErrorState what="player plot data" onRetry={() => refetch()} />
  );

  return (
    <>
      <FilterBar summary={`${filtered.length} players · click a dot to inspect`}>
        <Field label="X axis" htmlFor="axis-x">
          <AxisSelector label="X" value={xAxis} onChange={v => set({ x: v })} />
        </Field>
        <Field label="Y axis" htmlFor="axis-y">
          <AxisSelector label="Y" value={yAxis} onChange={v => set({ y: v })} />
        </Field>

        <Field label="Position">
          <PositionFilter value={pos} onChange={p => set({ pos: p })} />
        </Field>

        <Field label="Price (£m)">
          <PriceRange
            min={String(minPrice)}
            max={String(maxPrice)}
            onMinChange={v => set({ min_price: parseFloat(v) <= priceMin ? undefined : v })}
            onMaxChange={v => set({ max_price: parseFloat(v) >= priceMax ? undefined : v })}
            minBound={priceMin}
            maxBound={priceMax}
          />
        </Field>

        {/* Hide fringe players (low form + low ownership) */}
        <label className="flex cursor-pointer select-none items-center gap-2 pb-2 text-xs text-slate-300">
          <input
            type="checkbox"
            checked={!showAll}
            onChange={e => set({ all: e.target.checked ? undefined : '1' })}
            className="h-3.5 w-3.5 accent-indigo-600"
          />
          Hide fringe
          {!showAll && hiddenCount > 0 && (
            <span className="text-slate-500">({hiddenCount} hidden)</span>
          )}
        </label>
      </FilterBar>

      <ScatterPlot
        players={filtered}
        xAxis={xAxis}
        yAxis={yAxis}
        onPlayerClick={handlePlayerClick}
      />

      <PlayerDrawer player={selectedPlayer} onClose={handleDrawerClose} />
    </>
  );
}
