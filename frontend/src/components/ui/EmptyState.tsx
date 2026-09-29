import { Button } from './Button';

interface Props {
  /** What is empty, as a plain noun phrase: "players", "set-piece data". */
  what: string;
  /** Why it's empty / when it will fill in. Ignored when `filtered`. */
  hint?: string;
  /** Empty because filters/search excluded everything (vs. no data at all). */
  filtered?: boolean;
  /** Filtered only: resets the filters. */
  onClear?: () => void;
}

/**
 * Full-area empty state. Two wordings only:
 *   no data:  "No {what} yet." + hint
 *   filtered: "No {what} match your filters." + Clear filters
 */
export function EmptyState({ what, hint, filtered = false, onClear }: Props) {
  return (
    <div className="flex flex-col items-center gap-2 py-16 text-center">
      <span className="text-3xl text-slate-600" aria-hidden="true">∅</span>
      <p className="text-sm text-slate-300">
        {filtered ? `No ${what} match your filters.` : `No ${what} yet.`}
      </p>
      {!filtered && hint && <p className="text-xs text-slate-500">{hint}</p>}
      {filtered && onClear && (
        <Button variant="secondary" onClick={onClear} className="mt-1">Clear filters</Button>
      )}
    </div>
  );
}

/** Empty inside a card or table cell — same "No {what}." phrasing, no chrome. */
export function InlineEmpty({ children }: { children: React.ReactNode }) {
  return <p className="py-2 text-xs text-slate-500">{children}</p>;
}
