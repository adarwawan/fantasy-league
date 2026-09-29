interface Props {
  children: React.ReactNode;
  /** Right-aligned result count / hint, e.g. "42 players". */
  summary?: React.ReactNode;
}

/**
 * The one filter row under every PageHeader. Children are <Field>s in a fixed
 * order: search → position → price → page-specific controls.
 */
export function FilterBar({ children, summary }: Props) {
  return (
    <div className="mb-4 flex flex-wrap items-end gap-x-4 gap-y-3">
      {children}
      {summary && <div className="ml-auto pb-1.5 text-xs text-slate-500">{summary}</div>}
    </div>
  );
}
