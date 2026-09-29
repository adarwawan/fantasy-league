interface Option<T extends string> {
  value: T;
  label: React.ReactNode;
}

interface Props<T extends string> {
  options: readonly Option<T>[];
  value: T | null;
  onChange: (value: T) => void;
  /** Accessible name for the group. */
  label: string;
  size?: 'sm' | 'md';
}

const SIZE = { sm: 'px-2.5 py-1 text-xs', md: 'px-3 py-1.5 text-sm' } as const;

/** The one toggle-group used for views, filters and modes across the app. */
export function SegmentedControl<T extends string>({ options, value, onChange, label, size = 'md' }: Props<T>) {
  return (
    <div
      role="group"
      aria-label={label}
      className="inline-flex overflow-hidden rounded-md border border-line bg-surface"
    >
      {options.map((o, i) => {
        const active = o.value === value;
        return (
          <button
            key={o.value}
            type="button"
            aria-pressed={active}
            onClick={() => onChange(o.value)}
            className={`${SIZE[size]} font-medium transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-accent-ring ${
              i > 0 ? 'border-l border-line' : ''
            } ${active ? 'bg-accent text-white' : 'text-slate-300 hover:bg-surface-raised hover:text-white'}`}
          >
            {o.label}
          </button>
        );
      })}
    </div>
  );
}
