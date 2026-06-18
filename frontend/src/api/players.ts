import { apiFetch } from './client';
import type { PlayersResponse } from '../types/player';

export interface PlayerQueryParams {
  sort?:      'global_ownership' | 'top_n_ownership' | 'form' | 'price';
  pos?:       'GK' | 'DEF' | 'MID' | 'FWD';
  max_price?: number;
  top_n?:     number;
}

export function fetchPlayers(game: string, params: PlayerQueryParams = {}): Promise<PlayersResponse> {
  const qs = new URLSearchParams();
  if (params.sort)      qs.set('sort', params.sort);
  if (params.pos)       qs.set('pos', params.pos);
  if (params.max_price) qs.set('max_price', String(params.max_price));
  if (params.top_n)     qs.set('top_n', String(params.top_n));
  const query = qs.toString();
  return apiFetch<PlayersResponse>(`/api/${game}/players${query ? `?${query}` : ''}`);
}
