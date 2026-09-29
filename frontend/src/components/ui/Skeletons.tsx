/** Shimmering placeholder bar. Compose skeletons from this. */
export function SkeletonBar({ className = '' }: { className?: string }) {
  return <div className={`animate-pulse rounded bg-slate-700 ${className}`} />;
}

/** Table loading state; pass one width class per column. */
export function TableSkeleton({ cols, rows = 8, label = 'Loading' }: { cols: string[]; rows?: number; label?: string }) {
  return (
    <div className="overflow-x-auto rounded-lg border border-line">
      <table className="w-full text-sm" aria-label={label} aria-busy="true">
        <tbody>
          {Array.from({ length: rows }).map((_, i) => (
            <tr key={i} className="border-b border-line">
              {cols.map((w, j) => (
                <td key={j} className="px-4 py-3"><SkeletonBar className={`h-3.5 ${w}`} /></td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

/** Card-grid loading state. `className` sets the grid columns to match the real grid. */
export function CardGridSkeleton({
  count = 6,
  height = 'h-40',
  className = 'grid gap-3 sm:grid-cols-2 lg:grid-cols-3',
}: { count?: number; height?: string; className?: string }) {
  return (
    <div className={className} aria-busy="true">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className={`animate-pulse rounded-xl border border-line bg-surface-raised/40 ${height}`} />
      ))}
    </div>
  );
}

/** Single large block (charts, detail panels). */
export function PanelSkeleton({ height = 'h-72' }: { height?: string }) {
  return <div aria-busy="true" className={`animate-pulse rounded-2xl border border-line bg-surface ${height}`} />;
}
