import { CONTROL_CLASS } from './controlStyles';

interface Props {
  min: string;
  max: string;
  onMinChange: (v: string) => void;
  onMaxChange: (v: string) => void;
  /** Commit hooks for draft-style inputs (fire on blur / Enter). */
  onMinCommit?: (v: string) => void;
  onMaxCommit?: (v: string) => void;
  minBound?: number;
  maxBound?: number;
  step?: number;
  placeholders?: { min: string; max: string };
}

const INPUT = `${CONTROL_CLASS} w-16 px-2 text-center tabular-nums`;

/** "£m min – max" pair. Pair with <Field label="Price (£m)">. */
export function PriceRange({
  min, max, onMinChange, onMaxChange, onMinCommit, onMaxCommit,
  minBound, maxBound, step = 0.5, placeholders,
}: Props) {
  const commitKey = (fn?: (v: string) => void) => (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && fn) (e.target as HTMLInputElement).blur();
  };
  return (
    <div className="flex items-center gap-1.5" role="group" aria-label="Price range in £m">
      <input
        type="number" step={step} min={minBound} max={maxBound}
        value={min} placeholder={placeholders?.min}
        onChange={(e) => onMinChange(e.target.value)}
        onBlur={onMinCommit ? (e) => onMinCommit(e.target.value) : undefined}
        onKeyDown={commitKey(onMinCommit)}
        aria-label="Minimum price" className={INPUT}
      />
      <span className="text-xs text-slate-500">–</span>
      <input
        type="number" step={step} min={minBound} max={maxBound}
        value={max} placeholder={placeholders?.max}
        onChange={(e) => onMaxChange(e.target.value)}
        onBlur={onMaxCommit ? (e) => onMaxCommit(e.target.value) : undefined}
        onKeyDown={commitKey(onMaxCommit)}
        aria-label="Maximum price" className={INPUT}
      />
    </div>
  );
}
