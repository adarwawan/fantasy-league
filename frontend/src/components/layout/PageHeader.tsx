interface Props {
  /** Page name only — the game is already shown by the game switcher. */
  title: string;
  /** One sentence on what the page shows and where the data comes from. */
  description?: string;
  /** Data scope, e.g. "Last 6 matches". Right-aligned next to the actions. */
  meta?: React.ReactNode;
  /** View toggles / primary page actions. */
  actions?: React.ReactNode;
}

export function PageHeader({ title, description, meta, actions }: Props) {
  return (
    <div className="mb-4 flex flex-wrap items-start justify-between gap-x-4 gap-y-2">
      <div className="min-w-0">
        <h1 className="text-xl font-semibold text-slate-100">{title}</h1>
        {description && <p className="mt-1 max-w-2xl text-sm text-slate-400">{description}</p>}
      </div>
      {(meta || actions) && (
        <div className="flex items-center gap-3">
          {meta && <span className="text-xs text-slate-500">{meta}</span>}
          {actions}
        </div>
      )}
    </div>
  );
}
