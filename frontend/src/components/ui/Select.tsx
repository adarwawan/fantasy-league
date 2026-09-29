import { CONTROL_CLASS } from './controlStyles';

interface Props<T extends string | number> {
  value: T;
  onChange: (value: string) => void;
  options: readonly { value: T; label: string }[];
  /** Accessible name (or use <Field> to show a visible label). */
  label?: string;
  id?: string;
}

export function Select<T extends string | number>({ value, onChange, options, label, id }: Props<T>) {
  return (
    <select
      id={id}
      value={value}
      aria-label={label}
      onChange={(e) => onChange(e.target.value)}
      className={CONTROL_CLASS}
    >
      {options.map((o) => (
        <option key={o.value} value={o.value}>{o.label}</option>
      ))}
    </select>
  );
}
