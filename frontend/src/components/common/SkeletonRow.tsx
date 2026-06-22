/** Animated shimmer row for table loading states. Pass one width class per column cell. */
export function SkeletonRow({ cols }: { cols: string[] }) {
  return (
    <tr className="border-b border-slate-700/40">
      {cols.map((w, i) => (
        <td key={i} className="px-4 py-3">
          <div className={`h-3.5 ${w} rounded bg-slate-700 animate-pulse`} />
        </td>
      ))}
    </tr>
  );
}
