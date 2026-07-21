import { useCallback, useEffect, useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';
import { usePlayers } from '../hooks/usePlayers';
import { useTeams } from '../hooks/useTeams';
import { fetchEntry } from '../api/entry';
import { ApiError } from '../api/client';
import type { Player } from '../types/player';
import { PlayerCard } from '../components/players/PlayerCard';
import { PlayerDrawer } from '../components/players/PlayerDrawer';
import { PositionBadge } from '../components/common/PositionBadge';
import { SkeletonRow } from '../components/common/SkeletonRow';
import { ErrorState } from '../components/common/ErrorState';

type Position = Player['position'];

// Standard FPL squad shape: 2 GK, 5 DEF, 5 MID, 3 FWD (15 players), max 3 per club.
const SQUAD_QUOTA: Record<Position, number> = { GK: 2, DEF: 5, MID: 5, FWD: 3 };
const POSITION_ORDER: Position[] = ['GK', 'DEF', 'MID', 'FWD'];
const MAX_PER_CLUB = 3;
const DEFAULT_TEAM_VALUE = 100.0;

// Games whose backend source implements team loading (fantasy.EntryLoader).
// Extend this as other games gain the capability — the API side is already
// generic, so this is the only frontend change needed.
const ENTRY_LOADER_GAMES = ['fpl'];

// A picked player carries an editable sell price — a squad member's real value
// can differ from the current market price (profit is split with the game).
interface Pick {
  player:    Player;
  sellPrice: number;
}

function storageKey(game: string) {
  return `planner:${game}`;
}

interface PersistedState {
  teamValue: number;
  picks:     { id: string; sellPrice: number }[];
}

export function PlannerPage() {
  const { game = 'fpl' } = useParams<{ game: string }>();
  const { data, isLoading, isError, refetch } = usePlayers(game);
  const { data: teamsData } = useTeams(game);

  const [teamValue, setTeamValue] = useState(DEFAULT_TEAM_VALUE);
  const [picks, setPicks] = useState<Pick[]>([]);
  const [search, setSearch] = useState('');
  const [posFilter, setPosFilter] = useState<Position | null>(null);
  const [minPrice, setMinPrice] = useState('');
  const [maxPrice, setMaxPrice] = useState('');
  const [selectedPlayer, setSelectedPlayer] = useState<Player | null>(null);
  const [hydrated, setHydrated] = useState(false);
  const [entryId, setEntryId] = useState('');
  const [loadingEntry, setLoadingEntry] = useState(false);
  const [entryError, setEntryError] = useState<string | null>(null);

  useEffect(() => {
    document.title = `${game.toUpperCase()} — Planner`;
  }, [game]);

  // Restore a saved squad once the player list is available (we need the full
  // Player objects to rebuild picks from stored ids).
  useEffect(() => {
    if (!data) return;
    setHydrated(false);
    try {
      const raw = localStorage.getItem(storageKey(game));
      if (raw) {
        const saved = JSON.parse(raw) as PersistedState;
        const byId = new Map(data.players.map(p => [p.id, p]));
        const restored: Pick[] = [];
        for (const s of saved.picks) {
          const player = byId.get(s.id);
          if (player) restored.push({ player, sellPrice: s.sellPrice });
        }
        setTeamValue(saved.teamValue ?? DEFAULT_TEAM_VALUE);
        setPicks(restored);
      } else {
        setTeamValue(DEFAULT_TEAM_VALUE);
        setPicks([]);
      }
    } catch {
      /* ignore malformed storage */
    }
    setHydrated(true);
  }, [data, game]);

  // Persist on change (after initial hydration so we don't clobber saved state).
  useEffect(() => {
    if (!hydrated) return;
    const payload: PersistedState = {
      teamValue,
      picks: picks.map(p => ({ id: p.player.id, sellPrice: p.sellPrice })),
    };
    localStorage.setItem(storageKey(game), JSON.stringify(payload));
  }, [teamValue, picks, hydrated, game]);

  const pickedIds = useMemo(() => new Set(picks.map(p => p.player.id)), [picks]);

  const posCounts = useMemo(() => {
    const c: Record<Position, number> = { GK: 0, DEF: 0, MID: 0, FWD: 0 };
    for (const p of picks) c[p.player.position]++;
    return c;
  }, [picks]);

  const clubCounts = useMemo(() => {
    const c = new Map<string, number>();
    for (const p of picks) c.set(p.player.team.id, (c.get(p.player.team.id) ?? 0) + 1);
    return c;
  }, [picks]);

  const squadCost = useMemo(() => picks.reduce((s, p) => s + p.sellPrice, 0), [picks]);
  const bank = teamValue - squadCost;

  const canAdd = useCallback(
    (player: Player): string | null => {
      if (pickedIds.has(player.id)) return 'Already picked';
      if (posCounts[player.position] >= SQUAD_QUOTA[player.position]) {
        return `${player.position} full (${SQUAD_QUOTA[player.position]})`;
      }
      if ((clubCounts.get(player.team.id) ?? 0) >= MAX_PER_CLUB) {
        return `Max ${MAX_PER_CLUB} from ${player.team.short_name}`;
      }
      return null;
    },
    [pickedIds, posCounts, clubCounts],
  );

  const addPlayer = useCallback((player: Player) => {
    setPicks(prev => {
      if (prev.some(p => p.player.id === player.id)) return prev;
      return [...prev, { player, sellPrice: player.price }];
    });
  }, []);

  const removePlayer = useCallback((id: string) => {
    setPicks(prev => prev.filter(p => p.player.id !== id));
  }, []);

  const updateSellPrice = useCallback((id: string, sellPrice: number) => {
    setPicks(prev => prev.map(p => (p.player.id === id ? { ...p, sellPrice } : p)));
  }, []);

  const clearSquad = useCallback(() => setPicks([]), []);

  // Load a manager's current squad + budget from their FPL ID as a starting
  // point. Sell prices aren't exposed by the public API, so they default to the
  // current list price — adjust any that differ using the per-player editor.
  const loadEntry = useCallback(async () => {
    const id = entryId.trim();
    if (!id || !data) return;
    setLoadingEntry(true);
    setEntryError(null);
    try {
      const entry = await fetchEntry(game, id);
      const byId = new Map(data.players.map(p => [p.id, p]));
      const loaded: Pick[] = [];
      for (const p of entry.picks) {
        const player = byId.get(p.player_id);
        if (player) loaded.push({ player, sellPrice: player.price });
      }
      if (loaded.length === 0) {
        setEntryError('Loaded team had no matching players.');
        return;
      }
      setTeamValue(entry.team_value);
      setPicks(loaded);
    } catch (err) {
      setEntryError(err instanceof ApiError ? err.message : 'Failed to load team.');
    } finally {
      setLoadingEntry(false);
    }
  }, [entryId, data, game]);

  // Available list: not yet picked, matches search + position filter.
  const available = useMemo(() => {
    if (!data) return [];
    const q = search.trim().toLowerCase();
    const min = parseFloat(minPrice);
    const max = parseFloat(maxPrice);
    return data.players
      .filter(p => !pickedIds.has(p.id))
      .filter(p => !posFilter || p.position === posFilter)
      .filter(p => (isNaN(min) || p.price >= min) && (isNaN(max) || p.price <= max))
      .filter(p =>
        !q ||
        p.name.toLowerCase().includes(q) ||
        p.team.short_name.toLowerCase().includes(q) ||
        p.team.name.toLowerCase().includes(q),
      )
      .slice(0, 50);
  }, [data, search, posFilter, minPrice, maxPrice, pickedIds]);

  if (isLoading) {
    return (
      <div>
        <h1 className="text-xl font-semibold text-slate-100 mb-4">{game.toUpperCase()} — Planner</h1>
        <div className="overflow-x-auto rounded-lg border border-slate-700/50">
          <table className="w-full text-sm" aria-label="Loading planner">
            <tbody>
              {Array.from({ length: 8 }).map((_, i) => (
                <SkeletonRow key={i} cols={['w-32', 'w-12', 'w-14', 'w-40']} />
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

  const squadFull = picks.length >= 15;
  const bankColor = bank < 0 ? 'text-red-400' : 'text-emerald-400';

  return (
    <>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-xl font-semibold text-slate-100">{game.toUpperCase()} — Planner</h1>
        {picks.length > 0 && (
          <button
            onClick={clearSquad}
            className="text-xs text-slate-400 hover:text-red-400 border border-slate-700 rounded-md px-2.5 py-1"
          >
            Clear squad
          </button>
        )}
      </div>

      {/* Load an existing team from an FPL manager ID (FPL only) */}
      {ENTRY_LOADER_GAMES.includes(game) && (
        <div className="mb-4 rounded-lg border border-slate-700/50 bg-slate-800/40 px-4 py-3">
          <div className="flex flex-wrap items-end gap-2">
            <div>
              <label htmlFor="fpl-entry" className="block text-[10px] uppercase tracking-wide text-slate-500 mb-1">
                Load from {game.toUpperCase()} ID
              </label>
              <input
                id="fpl-entry"
                type="text"
                inputMode="numeric"
                value={entryId}
                onChange={e => setEntryId(e.target.value)}
                onKeyDown={e => { if (e.key === 'Enter') loadEntry(); }}
                placeholder="e.g. 123456"
                className="w-40 px-3 py-1.5 rounded-md bg-slate-700/50 border border-slate-600 text-sm text-slate-100 placeholder-slate-500 tabular-nums focus:outline-none focus:ring-1 focus:ring-indigo-500"
              />
            </div>
            <button
              onClick={loadEntry}
              disabled={!entryId.trim() || loadingEntry}
              className="px-3 py-1.5 rounded-md bg-indigo-600 text-white text-sm font-medium hover:bg-indigo-500 disabled:opacity-40 disabled:cursor-not-allowed"
            >
              {loadingEntry ? 'Loading…' : 'Load team'}
            </button>
            {entryError && <span className="text-xs text-red-400">{entryError}</span>}
          </div>
          <p className="mt-1.5 text-[11px] text-slate-500">
            Loads your latest squad, team value and bank. Sell prices default to list price — adjust any that differ below.
          </p>
        </div>
      )}

      {/* Budget summary */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-px bg-slate-700 rounded-lg overflow-hidden border border-slate-700 mb-4">
        <div className="bg-slate-800/80 px-4 py-3">
          <label htmlFor="team-value" className="block text-[10px] uppercase tracking-wide text-slate-500 mb-1">
            Team value (£m)
          </label>
          <input
            id="team-value"
            type="number"
            step={0.1}
            min={0}
            value={teamValue}
            onChange={e => setTeamValue(parseFloat(e.target.value) || 0)}
            className="w-24 bg-transparent text-lg font-semibold text-slate-100 tabular-nums focus:outline-none focus:ring-1 focus:ring-indigo-500 rounded px-1 -ml-1"
          />
        </div>
        <div className="bg-slate-800/80 px-4 py-3">
          <p className="text-[10px] uppercase tracking-wide text-slate-500 mb-1">Squad cost</p>
          <p className="text-lg font-semibold text-slate-100 tabular-nums">£{squadCost.toFixed(1)}m</p>
        </div>
        <div className="bg-slate-800/80 px-4 py-3">
          <p className="text-[10px] uppercase tracking-wide text-slate-500 mb-1">In the bank</p>
          <p className={`text-lg font-semibold tabular-nums ${bankColor}`}>£{bank.toFixed(1)}m</p>
        </div>
        <div className="bg-slate-800/80 px-4 py-3">
          <p className="text-[10px] uppercase tracking-wide text-slate-500 mb-1">Squad</p>
          <p className="text-lg font-semibold text-slate-100 tabular-nums">{picks.length}/15</p>
          <p className="text-[10px] text-slate-500 tabular-nums mt-0.5">
            {POSITION_ORDER.map(pos => `${pos} ${posCounts[pos]}/${SQUAD_QUOTA[pos]}`).join(' · ')}
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Squad column */}
        <section>
          <h2 className="text-sm font-semibold text-slate-300 mb-3">Your squad</h2>
          {picks.length === 0 ? (
            <p className="text-sm text-slate-500 border border-dashed border-slate-700 rounded-lg px-4 py-8 text-center">
              Add players from the list to build your squad.
            </p>
          ) : (
            <div className="flex flex-col gap-4">
              {POSITION_ORDER.map(pos => {
                const group = picks.filter(p => p.player.position === pos);
                if (group.length === 0) return null;
                return (
                  <div key={pos}>
                    <div className="flex items-center gap-2 mb-1.5">
                      <PositionBadge position={pos} />
                      <span className="text-xs text-slate-500 tabular-nums">
                        {group.length}/{SQUAD_QUOTA[pos]}
                      </span>
                    </div>
                    <div className="flex flex-col gap-1.5">
                      {group.map(({ player, sellPrice }) => {
                        const priceDelta = sellPrice - player.price;
                        return (
                          <div
                            key={player.id}
                            className="rounded-lg border border-slate-700/50 bg-slate-800/50"
                          >
                            {/* Full decision info — status, form, ownership, EO,
                                differential, set-piece duties, fixtures. */}
                            <PlayerCard player={player} teams={teamsData?.teams} onClick={setSelectedPlayer} />
                            {/* Sell-price editor + remove */}
                            <div className="flex items-center gap-2 border-t border-slate-700/50 px-3 py-1.5">
                              <span className="text-[10px] uppercase tracking-wide text-slate-500">Sell price</span>
                              <span className="text-slate-400 text-sm">£</span>
                              <input
                                type="number"
                                step={0.1}
                                min={0}
                                value={sellPrice}
                                onChange={e => updateSellPrice(player.id, parseFloat(e.target.value) || 0)}
                                aria-label={`Sell price for ${player.name}`}
                                className="w-16 px-2 py-1 rounded-md bg-slate-700/50 border border-slate-600 text-sm text-slate-100 text-center tabular-nums focus:outline-none focus:ring-1 focus:ring-indigo-500"
                              />
                              <span className="text-[11px] text-slate-500 tabular-nums">
                                list £{player.price.toFixed(1)}m
                                {priceDelta !== 0 && (
                                  <span className={priceDelta > 0 ? 'text-emerald-400' : 'text-red-400'}>
                                    {' '}({priceDelta > 0 ? '+' : ''}{priceDelta.toFixed(1)})
                                  </span>
                                )}
                              </span>
                              <button
                                onClick={() => removePlayer(player.id)}
                                className="ml-auto text-slate-500 hover:text-red-400 text-lg leading-none shrink-0"
                                aria-label={`Remove ${player.name}`}
                              >
                                ✕
                              </button>
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </section>

        {/* Add players column */}
        <section>
          <h2 className="text-sm font-semibold text-slate-300 mb-3">Add players</h2>
          <div className="flex flex-wrap gap-2 items-center mb-3">
            <input
              type="search"
              value={search}
              onChange={e => setSearch(e.target.value)}
              placeholder="Search player or team…"
              className="flex-1 min-w-[10rem] px-3 py-1.5 rounded-md bg-slate-700/50 border border-slate-600 text-sm text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              aria-label="Search players to add"
            />
            <div className="flex rounded-md border border-slate-600 overflow-hidden text-sm" role="group" aria-label="Filter by position">
              <button
                aria-pressed={!posFilter}
                className={`px-2.5 py-1.5 ${!posFilter ? 'bg-indigo-600 text-white' : 'bg-slate-700/50 text-slate-300 hover:bg-slate-600'}`}
                onClick={() => setPosFilter(null)}
              >
                All
              </button>
              {POSITION_ORDER.map(p => (
                <button
                  key={p}
                  aria-pressed={posFilter === p}
                  className={`px-2.5 py-1.5 border-l border-slate-600 ${posFilter === p ? 'bg-indigo-600 text-white' : 'bg-slate-700/50 text-slate-300 hover:bg-slate-600'}`}
                  onClick={() => setPosFilter(posFilter === p ? null : p)}
                >
                  {p}
                </button>
              ))}
            </div>
            <div className="flex items-center gap-1.5" role="group" aria-label="Filter by price">
              <span className="text-xs text-slate-500">£</span>
              <input
                type="number"
                step={0.1}
                min={0}
                value={minPrice}
                onChange={e => setMinPrice(e.target.value)}
                placeholder="min"
                aria-label="Minimum price"
                className="w-16 px-2 py-1.5 rounded-md bg-slate-700/50 border border-slate-600 text-sm text-slate-100 text-center tabular-nums placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              />
              <span className="text-slate-500 text-xs">–</span>
              <input
                type="number"
                step={0.1}
                min={0}
                value={maxPrice}
                onChange={e => setMaxPrice(e.target.value)}
                placeholder="max"
                aria-label="Maximum price"
                className="w-16 px-2 py-1.5 rounded-md bg-slate-700/50 border border-slate-600 text-sm text-slate-100 text-center tabular-nums placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              />
            </div>
          </div>

          <div className="flex flex-col gap-1.5 max-h-[70vh] overflow-y-auto pr-1">
            {available.length === 0 ? (
              <p className="text-sm text-slate-500 py-4 text-center">No players match.</p>
            ) : (
              available.map(player => {
                const blockReason = canAdd(player);
                return (
                  <div key={player.id} className="flex items-center gap-2">
                    <div className="flex-1 min-w-0">
                      <PlayerCard player={player} teams={teamsData?.teams} onClick={setSelectedPlayer} />
                    </div>
                    <button
                      onClick={() => addPlayer(player)}
                      disabled={!!blockReason || squadFull}
                      title={blockReason ?? undefined}
                      className="shrink-0 w-9 h-9 rounded-md bg-indigo-600 text-white text-lg font-semibold hover:bg-indigo-500 disabled:opacity-30 disabled:cursor-not-allowed"
                      aria-label={`Add ${player.name}`}
                    >
                      +
                    </button>
                  </div>
                );
              })
            )}
          </div>
        </section>
      </div>

      <PlayerDrawer player={selectedPlayer} teams={teamsData?.teams} onClose={() => setSelectedPlayer(null)} />
    </>
  );
}
