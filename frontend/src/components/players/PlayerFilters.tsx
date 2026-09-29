import { useEffect, useState, type RefObject } from 'react';
import type { PlayerQueryParams } from '../../api/players';
import { FilterBar } from '../ui/FilterBar';
import { Field } from '../ui/Field';
import { SearchInput } from '../ui/SearchInput';
import { PositionFilter } from '../ui/PositionFilter';
import { PriceRange } from '../ui/PriceRange';
import { SegmentedControl } from '../ui/SegmentedControl';

// Fallback bounds used before player data is available. The live floor/cap are
// derived from the cheapest/most expensive loaded player (see PlayersPage) and
// passed in via `priceMin`/`priceMax`, so in-season price moves are tracked
// automatically.
const PRICE_MIN_FALLBACK = 4.0;
const PRICE_MAX_FALLBACK = 15.5;

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
  /** Live price floor/ceiling derived from the data; fall back when omitted. */
  priceMin?:    number;
  priceMax?:    number;
}

export function PlayerFilters({ game, params, onChange, search, onSearch, searchRef, priceMin, priceMax }: Props) {
  const PRICE_MIN = priceMin ?? PRICE_MIN_FALLBACK;
  const PRICE_MAX = priceMax ?? PRICE_MAX_FALLBACK;
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
    <FilterBar>
      <Field
        label="Search player"
        htmlFor="player-search"
      >
        <SearchInput
          id="player-search"
          inputRef={searchRef}
          value={search}
          onChange={onSearch}
          placeholder="e.g. Salah  ( / )"
          label="Search players by name or team"
        />
      </Field>

      <Field label="Position">
        <PositionFilter value={params.pos} onChange={pos => onChange({ ...params, pos })} />
      </Field>

      <Field label="Price (£m)">
        <PriceRange
          min={minDraft}
          max={maxDraft}
          onMinChange={setMinDraft}
          onMaxChange={setMaxDraft}
          onMinCommit={commitMin}
          onMaxCommit={commitMax}
          minBound={PRICE_MIN}
          maxBound={PRICE_MAX}
        />
      </Field>

      <Field label="Top-N managers">
        <SegmentedControl
          label="Top N managers"
          value={String(topN)}
          onChange={v => onChange({ ...params, top_n: Number(v) })}
          options={topNOptions.map(n => ({ value: String(n), label: `Top-${topNLabel(n)}` }))}
        />
      </Field>
    </FilterBar>
  );
}
