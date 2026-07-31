import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { usePlayers } from '../hooks/usePlayers';
import { useTeams } from '../hooks/useTeams';
import type { PlayerQueryParams } from '../api/players';
import type { Player } from '../types/player';
import { PlayerFilters } from '../components/players/PlayerFilters';
import { PlayerTable } from '../components/players/PlayerTable';
import { PlayerDrawer } from '../components/players/PlayerDrawer';
import { ScatterView } from '../components/scatter/ScatterView';
import { SkeletonRow } from '../components/common/SkeletonRow';
import { ErrorState } from '../components/common/ErrorState';
import { priceCeiling, priceFloor } from '../utils/price';

type PlayerView = 'table' | 'plot';

function ViewToggle({ view, onChange }: { view: PlayerView; onChange: (v: PlayerView) => void }) {
  const opts: { id: PlayerView; label: string }[] = [
    { id: 'table', label: 'Table' },
    { id: 'plot',  label: 'Plot'  },
  ];
  return (
    <div className="flex rounded-md border border-slate-600 overflow-hidden text-sm">
      {opts.map(({ id, label }) => (
        <button
          key={id}
          onClick={() => onChange(id)}
          className={`px-3 py-1 ${id !== 'table' ? 'border-l border-slate-600' : ''} ${
            view === id ? 'bg-indigo-600 text-white' : 'bg-slate-700/50 text-slate-300 hover:bg-slate-600'
          }`}
        >
          {label}
        </button>
      ))}
    </div>
  );
}

const PLAYER_SKELETON_COLS = [
  'w-32', 'w-12', 'w-10', 'w-14', 'w-10', 'w-14', 'w-14', 'w-14', 'w-40',
];

function paramsFromSearch(sp: URLSearchParams): PlayerQueryParams {
  const p: PlayerQueryParams = {};
  const sort = sp.get('sort');
  if (sort) p.sort = sort as PlayerQueryParams['sort'];
  const pos = sp.get('pos');
  if (pos) p.pos = pos as PlayerQueryParams['pos'];
  const minPrice = sp.get('min_price');
  if (minPrice) p.min_price = parseFloat(minPrice);
  const maxPrice = sp.get('max_price');
  if (maxPrice) p.max_price = parseFloat(maxPrice);
  const topN = sp.get('top_n');
  if (topN) p.top_n = parseInt(topN, 10);
  return p;
}

function paramsToSearch(p: PlayerQueryParams, priceMin: number, priceMax: number): Record<string, string> {
  const out: Record<string, string> = {};
  if (p.sort)      out.sort      = p.sort;
  if (p.pos)       out.pos       = p.pos;
  if (p.min_price && p.min_price > priceMin)   out.min_price = String(p.min_price);
  if (p.max_price && p.max_price < priceMax)   out.max_price = String(p.max_price);
  if (p.top_n)     out.top_n     = String(p.top_n);
  return out;
}

function useDebounced<T>(value: T, delay = 200): T {
  const [debounced, setDebounced] = useState(value);
  const timerRef = useRef<ReturnType<typeof setTimeout>>();
  useEffect(() => {
    clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(timerRef.current);
  }, [value, delay]);
  return debounced;
}

export function PlayersPage() {
  const { game = 'fpl' } = useParams<{ game: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const [searchInput, setSearchInput] = useState('');
  const [selectedPlayer, setSelectedPlayer] = useState<Player | null>(null);
  const debouncedSearch = useDebounced(searchInput);
  const searchRef = useRef<HTMLInputElement>(null);

  const view: PlayerView = searchParams.get('view') === 'plot' ? 'plot' : 'table';

  const setView = useCallback((next: PlayerView) => {
    const sp = new URLSearchParams(searchParams);
    if (next === 'plot') sp.set('view', 'plot');
    else sp.delete('view');
    setSearchParams(sp, { replace: true });
  }, [searchParams, setSearchParams]);

  useEffect(() => {
    document.title = `${game.toUpperCase()} — Players`;
  }, [game]);

  const header = (
    <div className="flex items-center justify-between gap-3 mb-4">
      <h1 className="text-xl font-semibold text-slate-100">
        {game.toUpperCase()} — Players
      </h1>
      <ViewToggle view={view} onChange={setView} />
    </div>
  );

  // `/` global shortcut focuses search
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === '/' && document.activeElement?.tagName !== 'INPUT' && document.activeElement?.tagName !== 'TEXTAREA') {
        e.preventDefault();
        searchRef.current?.focus();
      }
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);

  const params = paramsFromSearch(searchParams);
  const { data, isLoading, isError, refetch } = usePlayers(game, params);
  const { data: teamsData } = useTeams(game);

  // Price bounds track the cheapest/most expensive loaded player (rounded to
  // £0.5), so the filter reaches players whose price has drifted during the
  // season, below £4.0 or above £15.5.
  const priceMin = useMemo(() => priceFloor(data?.players), [data]);
  const priceMax = useMemo(() => priceCeiling(data?.players), [data]);

  const filteredPlayers = useMemo(() => {
    if (!data) return [];
    let players = data.players;
    if (params.min_price && params.min_price > priceMin) {
      players = players.filter(p => p.price >= (params.min_price ?? priceMin));
    }
    if (!debouncedSearch.trim()) return players;
    const q = debouncedSearch.toLowerCase();
    return players.filter(p =>
      p.name.toLowerCase().includes(q) ||
      p.team.short_name.toLowerCase().includes(q) ||
      p.team.name.toLowerCase().includes(q)
    );
  }, [data, debouncedSearch, params.min_price, priceMin]);

  function handleChange(next: PlayerQueryParams) {
    setSearchParams(paramsToSearch(next, priceMin, priceMax), { replace: true });
  }

  const handlePlayerClick = useCallback((p: Player) => setSelectedPlayer(p), []);
  const handleDrawerClose = useCallback(() => setSelectedPlayer(null), []);

  // Plot view uses its own data source (useScatter) and owns its loading/
  // error/drawer states, so it short-circuits the table-data gating below.
  if (view === 'plot') {
    return (
      <div>
        {header}
        <ScatterView />
      </div>
    );
  }

  if (isLoading) {
    return (
      <div>
        {header}
        <div className="overflow-x-auto rounded-lg border border-slate-700/50">
          <table className="w-full text-sm" aria-label="Loading players">
            <tbody>
              {Array.from({ length: 10 }).map((_, i) => (
                <SkeletonRow key={i} cols={PLAYER_SKELETON_COLS} />
              ))}
            </tbody>
          </table>
        </div>
      </div>
    );
  }

  if (isError || !data) {
    return (
      <ErrorState
        message="Failed to load players. Check your connection and try again."
        onRetry={() => refetch()}
      />
    );
  }

  return (
    <>
      <div>
        {header}
        <PlayerFilters
          game={game}
          params={params}
          onChange={handleChange}
          search={searchInput}
          onSearch={setSearchInput}
          searchRef={searchRef}
          priceMin={priceMin}
          priceMax={priceMax}
        />
        <PlayerTable
          players={filteredPlayers}
          topNSize={data.meta.top_n_size}
          teams={teamsData?.teams}
          currentGw={data.meta.gw}
          onPlayerClick={handlePlayerClick}
        />
      </div>
      <PlayerDrawer player={selectedPlayer} teams={teamsData?.teams} currentGw={data.meta.gw} onClose={handleDrawerClose} />
    </>
  );
}
