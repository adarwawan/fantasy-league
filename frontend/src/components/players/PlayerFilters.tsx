import type { PlayerQueryParams } from '../../api/players';

type Position = 'GK' | 'DEF' | 'MID' | 'FWD';
const POSITIONS: Position[] = ['GK', 'DEF', 'MID', 'FWD'];

interface Props {
  params:    PlayerQueryParams;
  onChange:  (next: PlayerQueryParams) => void;
  search:    string;
  onSearch:  (v: string) => void;
}

export function PlayerFilters({ params, onChange, search, onSearch }: Props) {
  const topN     = params.top_n     ?? 10000;
  const maxPrice = params.max_price ?? 15.0;

  return (
    <div className="flex flex-wrap gap-4 items-end mb-4">
      {/* Search */}
      <div>
        <label className="block text-xs text-slate-400 mb-1">Search player</label>
        <input
          type="search"
          value={search}
          onChange={e => onSearch(e.target.value)}
          placeholder="e.g. Salah"
          className="px-3 py-1.5 rounded-md bg-slate-700/50 border border-slate-600 text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 w-40"
        />
      </div>

      {/* Position toggle */}
      <div>
        <label className="block text-xs text-slate-400 mb-1">Position</label>
        <div className="flex rounded-md border border-slate-600 overflow-hidden text-sm">
          <button
            className={`px-3 py-1.5 ${!params.pos ? 'bg-indigo-600 text-white' : 'bg-slate-700/50 text-slate-300 hover:bg-slate-600'}`}
            onClick={() => onChange({ ...params, pos: undefined })}
          >
            All
          </button>
          {POSITIONS.map(p => (
            <button
              key={p}
              className={`px-3 py-1.5 border-l border-slate-600 ${params.pos === p ? 'bg-indigo-600 text-white' : 'bg-slate-700/50 text-slate-300 hover:bg-slate-600'}`}
              onClick={() => onChange({ ...params, pos: params.pos === p ? undefined : p })}
            >
              {p}
            </button>
          ))}
        </div>
      </div>

      {/* Max price slider */}
      <div>
        <label className="block text-xs text-slate-400 mb-1">
          Max price: £{maxPrice.toFixed(1)}m
        </label>
        <input
          type="range"
          min={4}
          max={15}
          step={0.5}
          value={maxPrice}
          onChange={e => onChange({ ...params, max_price: parseFloat(e.target.value) })}
          className="w-36 accent-indigo-600"
        />
      </div>

      {/* Top-N */}
      <div>
        <label className="block text-xs text-slate-400 mb-1">Top-N</label>
        <div className="flex rounded-md border border-slate-600 overflow-hidden text-sm">
          {([1000, 10000, 100000] as const).map(n => (
            <button
              key={n}
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
