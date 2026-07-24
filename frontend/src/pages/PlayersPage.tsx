import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { usePlayers } from '../hooks/usePlayers';
import { useTeams } from '../hooks/useTeams';
import type { PlayerQueryParams } from '../api/players';
import type { Player } from '../types/player';
import { PlayerFilters } from '../components/players/PlayerFilters';
import { PlayerTable } from '../components/players/PlayerTable';
import { PlayerDrawer } from '../components/players/PlayerDrawer';
import { SkeletonRow } from '../components/common/SkeletonRow';
import { ErrorState } from '../components/common/ErrorState';

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

function paramsToSearch(p: PlayerQueryParams): Record<string, string> {
  const out: Record<string, string> = {};
  if (p.sort)      out.sort      = p.sort;
  if (p.pos)       out.pos       = p.pos;
  if (p.min_price && p.min_price > 4)   out.min_price = String(p.min_price);
  if (p.max_price && p.max_price < 15)  out.max_price = String(p.max_price);
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

  useEffect(() => {
    document.title = `${game.toUpperCase()} — Players`;
  }, [game]);

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

  const filteredPlayers = useMemo(() => {
    if (!data) return [];
    let players = data.players;
    if (params.min_price && params.min_price > 4) {
      players = players.filter(p => p.price >= (params.min_price ?? 4));
    }
    if (!debouncedSearch.trim()) return players;
    const q = debouncedSearch.toLowerCase();
    return players.filter(p =>
      p.name.toLowerCase().includes(q) ||
      p.team.short_name.toLowerCase().includes(q) ||
      p.team.name.toLowerCase().includes(q)
    );
  }, [data, debouncedSearch, params.min_price]);

  function handleChange(next: PlayerQueryParams) {
    setSearchParams(paramsToSearch(next), { replace: true });
  }

  const handlePlayerClick = useCallback((p: Player) => setSelectedPlayer(p), []);
  const handleDrawerClose = useCallback(() => setSelectedPlayer(null), []);

  if (isLoading) {
    return (
      <div>
        <h1 className="text-xl font-semibold text-slate-100 mb-4">{game.toUpperCase()} — Players</h1>
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
        <h1 className="text-xl font-semibold text-slate-100 mb-4">
          {game.toUpperCase()} — Players
        </h1>
        <PlayerFilters
          game={game}
          params={params}
          onChange={handleChange}
          search={searchInput}
          onSearch={setSearchInput}
          searchRef={searchRef}
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
