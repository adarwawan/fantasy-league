import { Select } from '../ui/Select';

export type AxisKey = 'global_ownership' | 'top_n_ownership' | 'effective_ownership' | 'form' | 'avg_fdr';

export const AXIS_OPTIONS: { value: AxisKey; label: string }[] = [
  { value: 'global_ownership', label: 'Global Own %' },
  { value: 'top_n_ownership', label: 'Top-N Own %' },
  { value: 'effective_ownership', label: 'Effective Own %' },
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
    <Select
      id={`axis-${label.toLowerCase()}`}
      label={`${label} axis`}
      value={value}
      onChange={v => onChange(v as AxisKey)}
      options={AXIS_OPTIONS}
    />
  );
}
