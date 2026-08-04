import { ApiError } from './client';
import type { LeadersResponse, SeriesResponse, SessionResponse } from '../types/trends';

// The Trends service runs as its own binary on a separate port, so it has its
// own base URL rather than sharing VITE_API_URL.
const TRENDS_BASE = import.meta.env.VITE_TRENDS_API_URL ?? 'http://localhost:8081';

async function trendsFetch<T>(path: string): Promise<T> {
  const res = await fetch(`${TRENDS_BASE}${path}`);
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, body.error ?? res.statusText);
  }
  return res.json() as Promise<T>;
}

export function fetchTrendsSession(): Promise<SessionResponse> {
  return trendsFetch<SessionResponse>('/api/trends/session');
}

export function fetchTrendsLeaders(
  window: string,
  direction: 'in' | 'out' = 'in',
  limit = 25,
): Promise<LeadersResponse> {
  return trendsFetch<LeadersResponse>(
    `/api/trends/leaders?window=${encodeURIComponent(window)}&direction=${direction}&limit=${limit}`,
  );
}

export function fetchTrendsSeries(playerExtId: number): Promise<SeriesResponse> {
  return trendsFetch<SeriesResponse>(`/api/trends/player/${playerExtId}/series`);
}
