export type AxisKey = 'global_ownership' | 'top_n_ownership' | 'form' | 'avg_fdr';

export const AXIS_OPTIONS: { value: AxisKey; label: string }[] = [
  { value: 'global_ownership', label: 'Global Own %' },
  { value: 'top_n_ownership', label: 'Top-N Own %' },
  { value: 'form',            label: 'Form' },
  { value: 'avg_fdr',         label: 'Avg FDR (next 3)' },
];

interface Props {
  label:    string;
  value:    AxisKey;
  onChange: (v: AxisKey) => void;
}

export function AxisSelector({ label, value, onChange }: Props) {
  return (
    <div className="flex items-center gap-2">
      <span className="text-xs text-gray-500 font-medium w-6">{label}</span>
      <select
        value={value}
        onChange={e => onChange(e.target.value as AxisKey)}
        className="text-sm border border-gray-200 rounded-md px-2 py-1.5 bg-white text-gray-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"
      >
        {AXIS_OPTIONS.map(opt => (
          <option key={opt.value} value={opt.value}>{opt.label}</option>
        ))}
      </select>
    </div>
  );
}
