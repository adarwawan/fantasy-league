/** A control with a small visible label above it — the filter-bar convention. */
export function Field({ label, htmlFor, children }: { label: string; htmlFor?: string; children: React.ReactNode }) {
  return (
    <div>
      <label htmlFor={htmlFor} className="mb-1 block text-xs text-slate-400">{label}</label>
      {children}
    </div>
  );
}
