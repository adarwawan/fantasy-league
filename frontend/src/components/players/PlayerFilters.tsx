import { type RefObject } from 'react';
import type { PlayerQueryParams } from '../../api/players';

type Position = 'GK' | 'DEF' | 'MID' | 'FWD';
const POSITIONS: Position[] = ['GK', 'DEF', 'MID', 'FWD'];

interface Props {
  params:       PlayerQueryParams;
  onChange:     (next: PlayerQueryParams) => void;
  search:       string;
  onSearch:     (v: string) => void;
  searchRef?:   RefObject<HTMLInputElement>;
}

export function PlayerFilters({ params, onChange, search, onSearch, searchRef }: Props) {
  const topN     = params.top_n     ?? 10000;
  const minPrice = params.min_price ?? 4.0;
  const maxPrice = params.max_price ?? 15.0;

  function handleMin(raw: string) {
    const v = parseFloat(raw);
    if (!isNaN(v)) onChange({ ...params, min_price: Math.min(v, maxPrice - 0.5) });
  }

  function handleMax(raw: string) {
    const v = parseFloat(raw);
    if (!isNaN(v)) onChange({ ...params, max_price: Math.max(v, minPrice + 0.5) });
  }

  return (
    <div className="flex flex-wrap gap-4 items-end mb-4">
      {/* Search */}
      <div>
        <label htmlFor="player-search" className="block text-xs text-slate-400 mb-1">
          Search player{' '}
          <kbd className="text-[10px] bg-slate-700 border border-slate-600 rounded px-1 py-0.5 font-mono text-slate-400">/</kbd>
        </label>
        <input
          id="player-search"
          ref={searchRef}
          type="search"
          value={search}
          onChange={e => onSearch(e.target.value)}
          onKeyDown={e => { if (e.key === 'Escape') { onSearch(''); (e.target as HTMLInputElement).blur(); } }}
          placeholder="e.g. Salah"
          className="px-3 py-1.5 rounded-md bg-slate-700/50 border border-slate-600 text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 w-40"
          aria-label="Search players by name or team"
        />
      </div>

      {/* Position toggle */}
      <div>
        <label className="block text-xs text-slate-400 mb-1">Position</label>
        <div className="flex rounded-md border border-slate-600 overflow-hidden text-sm" role="group" aria-label="Filter by position">
          <button
            aria-pressed={!params.pos}
            className={`px-3 py-1.5 ${!params.pos ? 'bg-indigo-600 text-white' : 'bg-slate-700/50 text-slate-300 hover:bg-slate-600'}`}
            onClick={() => onChange({ ...params, pos: undefined })}
          >
            All
          </button>
          {POSITIONS.map(p => (
            <button
              key={p}
              aria-pressed={params.pos === p}
              className={`px-3 py-1.5 border-l border-slate-600 ${params.pos === p ? 'bg-indigo-600 text-white' : 'bg-slate-700/50 text-slate-300 hover:bg-slate-600'}`}
              onClick={() => onChange({ ...params, pos: params.pos === p ? undefined : p })}
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
            value={minPrice}
            onChange={e => handleMin(e.target.value)}
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
            onChange={e => handleMax(e.target.value)}
            aria-label="Maximum price"
            className="w-16 px-2 py-1.5 rounded-md bg-slate-700/50 border border-slate-600 text-sm text-slate-100 text-center tabular-nums focus:outline-none focus:ring-1 focus:ring-indigo-500"
          />
        </div>
      </div>

      {/* Top-N */}
      <div>
        <label className="block text-xs text-slate-400 mb-1">Top-N</label>
        <div className="flex rounded-md border border-slate-600 overflow-hidden text-sm" role="group" aria-label="Top N managers">
          {([1000, 10000, 100000] as const).map(n => (
            <button
              key={n}
              aria-pressed={topN === n}
              className={`px-3 py-1.5 border-l first:border-l-0 border-slate-600 ${topN === n ? 'bg-indigo-600 text-white' : 'bg-slate-700/50 text-slate-300 hover:bg-slate-600'}`}
              onClick={() => onChange({ ...params, top_n: n })}
            >
              Top-{n >= 100000 ? '100k' : n >= 10000 ? '10k' : '1k'}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
