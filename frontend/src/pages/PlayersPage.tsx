import { useEffect } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { usePlayers } from '../hooks/usePlayers';
import type { PlayerQueryParams } from '../api/players';
import { PlayerFilters } from '../components/players/PlayerFilters';
import { PlayerTable } from '../components/players/PlayerTable';

function paramsFromSearch(sp: URLSearchParams): PlayerQueryParams {
  const p: PlayerQueryParams = {};
  const sort = sp.get('sort');
  if (sort) p.sort = sort as PlayerQueryParams['sort'];
  const pos = sp.get('pos');
  if (pos) p.pos = pos as PlayerQueryParams['pos'];
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
  if (p.max_price) out.max_price = String(p.max_price);
  if (p.top_n)     out.top_n     = String(p.top_n);
  return out;
}

export function PlayersPage() {
  const { game = 'fpl' } = useParams<{ game: string }>();
  const [searchParams, setSearchParams] = useSearchParams();

  useEffect(() => {
    document.title = `${game.toUpperCase()} — Players`;
  }, [game]);

  const params = paramsFromSearch(searchParams);
  const { data, isLoading, isError } = usePlayers(game, params);

  function handleChange(next: PlayerQueryParams) {
    setSearchParams(paramsToSearch(next), { replace: true });
  }

  if (isLoading) {
    return <div className="text-gray-500 py-8 text-center">Loading players…</div>;
  }

  if (isError || !data) {
    return <div className="text-red-500 py-8 text-center">Failed to load players.</div>;
  }

  return (
    <div>
      <h1 className="text-xl font-semibold text-gray-900 mb-4">
        {game.toUpperCase()} — Players
      </h1>
      <PlayerFilters params={params} onChange={handleChange} />
      <PlayerTable players={data.players} topNSize={data.meta.top_n_size} />
    </div>
  );
}
