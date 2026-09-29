import type { Ref } from 'react';
import { CONTROL_CLASS } from './controlStyles';

interface Props {
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
  /** Accessible name; defaults to the placeholder. */
  label?: string;
  id?: string;
  inputRef?: Ref<HTMLInputElement>;
  className?: string;
}

export function SearchInput({ value, onChange, placeholder, label, id, inputRef, className = 'w-full sm:w-48' }: Props) {
  return (
    <input
      id={id}
      ref={inputRef}
      type="search"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      onKeyDown={(e) => {
        if (e.key === 'Escape') { onChange(''); (e.target as HTMLInputElement).blur(); }
      }}
      placeholder={placeholder}
      aria-label={label ?? placeholder}
      className={`${CONTROL_CLASS} ${className}`}
    />
  );
}
