import { SegmentedControl } from './SegmentedControl';

export type Position = 'GK' | 'DEF' | 'MID' | 'FWD';

const OPTIONS = [
  { value: 'ALL', label: 'All' },
  { value: 'GK', label: 'GK' },
  { value: 'DEF', label: 'DEF' },
  { value: 'MID', label: 'MID' },
  { value: 'FWD', label: 'FWD' },
] as const;

/** All / GK / DEF / MID / FWD. `value` undefined or null means All. */
export function PositionFilter({
  value,
  onChange,
  size = 'md',
}: {
  value: Position | null | undefined;
  onChange: (next: Position | undefined) => void;
  size?: 'sm' | 'md';
}) {
  return (
    <SegmentedControl
      label="Filter by position"
      size={size}
      options={OPTIONS}
      value={value ?? 'ALL'}
      onChange={(v) => onChange(v === 'ALL' || v === value ? undefined : (v as Position))}
    />
  );
}
