import { useEffect, useState, type RefObject } from 'react';
import type { PlayerQueryParams } from '../../api/players';

const PRICE_MIN = 4.0;
const PRICE_MAX = 15.0;

type Position = 'GK' | 'DEF' | 'MID' | 'FWD';
const POSITIONS: Position[] = ['GK', 'DEF', 'MID', 'FWD'];

const TOP_N_OPTIONS: Record<string, readonly number[]> = {
  wcf: [100, 1000],
  fpl: [1000, 10000, 100000],
};

function topNLabel(n: number): string {
  if (n >= 100000) return '100k';
  if (n >= 10000)  return '10k';
  if (n >= 1000)   return '1k';
  return String(n);
}

interface Props {
  game:         string;
  params:       PlayerQueryParams;
  onChange:     (next: PlayerQueryParams) => void;
  search:       string;
  onSearch:     (v: string) => void;
  searchRef?:   RefObject<HTMLInputElement>;
}

export function PlayerFilters({ game, params, onChange, search, onSearch, searchRef }: Props) {
  const topNOptions = TOP_N_OPTIONS[game] ?? TOP_N_OPTIONS['fpl'];
  const topN     = params.top_n     ?? topNOptions[topNOptions.length - 1];
  const minPrice = params.min_price ?? PRICE_MIN;
  const maxPrice = params.max_price ?? PRICE_MAX;

  // Local draft state so typing doesn't apply the filter on every keystroke;
  // committed on blur / Enter. Re-sync when the applied params change externally.
  const [minDraft, setMinDraft] = useState(String(minPrice));
  const [maxDraft, setMaxDraft] = useState(String(maxPrice));
  useEffect(() => { setMinDraft(String(minPrice)); }, [minPrice]);
  useEffect(() => { setMaxDraft(String(maxPrice)); }, [maxPrice]);

  function commitMin(raw: string) {
    const v = parseFloat(raw);
    if (isNaN(v)) { setMinDraft(String(minPrice)); return; }
    const clamped = Math.min(Math.max(v, PRICE_MIN), maxPrice - 0.5);
    setMinDraft(String(clamped));
    onChange({ ...params, min_price: clamped });
  }

  function commitMax(raw: string) {
    const v = parseFloat(raw);
    if (isNaN(v)) { setMaxDraft(String(maxPrice)); return; }
    const clamped = Math.max(Math.min(v, PRICE_MAX), minPrice + 0.5);
    setMaxDraft(String(clamped));
    onChange({ ...params, max_price: clamped });
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
            min={PRICE_MIN}
            max={PRICE_MAX - 0.5}
            step={0.5}
            value={minDraft}
            onChange={e => setMinDraft(e.target.value)}
            onBlur={e => commitMin(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') (e.target as HTMLInputElement).blur(); }}
            aria-label="Minimum price"
            className="w-16 px-2 py-1.5 rounded-md bg-slate-700/50 border border-slate-600 text-sm text-slate-100 text-center tabular-nums focus:outline-none focus:ring-1 focus:ring-indigo-500"
          />
          <span className="text-slate-500 text-xs">–</span>
          <input
            type="number"
            min={PRICE_MIN + 0.5}
            max={PRICE_MAX}
            step={0.5}
            value={maxDraft}
            onChange={e => setMaxDraft(e.target.value)}
            onBlur={e => commitMax(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter') (e.target as HTMLInputElement).blur(); }}
            aria-label="Maximum price"
            className="w-16 px-2 py-1.5 rounded-md bg-slate-700/50 border border-slate-600 text-sm text-slate-100 text-center tabular-nums focus:outline-none focus:ring-1 focus:ring-indigo-500"
          />
        </div>
      </div>

      {/* Top-N */}
      <div>
        <label className="block text-xs text-slate-400 mb-1">Top-N</label>
        <div className="flex rounded-md border border-slate-600 overflow-hidden text-sm" role="group" aria-label="Top N managers">
          {topNOptions.map(n => (
            <button
              key={n}
              aria-pressed={topN === n}
              className={`px-3 py-1.5 border-l first:border-l-0 border-slate-600 ${topN === n ? 'bg-indigo-600 text-white' : 'bg-slate-700/50 text-slate-300 hover:bg-slate-600'}`}
              onClick={() => onChange({ ...params, top_n: n })}
            >
              Top-{topNLabel(n)}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
