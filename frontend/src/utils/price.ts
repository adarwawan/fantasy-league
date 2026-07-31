import type { Player } from '../types/player';

// FPL prices rise and fall during the season, so the price-filter ceiling is
// derived from the data rather than hardcoded: take the most expensive player
// and round up to the next £0.5. A floor keeps the control from shrinking below
// the expected top price when data is sparse (e.g. preseason).
export const PRICE_CEILING_FLOOR = 15.5;
// Cheapest a player is expected to be; the live floor drops below this only if
// the data actually contains a cheaper player (prices fall during the season).
export const PRICE_FLOOR_CAP = 4.0;

export function priceCeiling(players: Player[] | undefined): number {
  const max = (players ?? []).reduce((m, p) => Math.max(m, p.price), 0);
  const rounded = Math.ceil(max * 2) / 2;
  return Math.max(rounded, PRICE_CEILING_FLOOR);
}

export function priceFloor(players: Player[] | undefined): number {
  const list = players ?? [];
  if (list.length === 0) return PRICE_FLOOR_CAP;
  const min = list.reduce((m, p) => Math.min(m, p.price), Infinity);
  const rounded = Math.floor(min * 2) / 2;
  return Math.min(rounded, PRICE_FLOOR_CAP);
}
