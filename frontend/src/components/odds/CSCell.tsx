interface CSCellProps {
  value:     number;
  highlight: boolean;
}

export function CSCell({ value, highlight }: CSCellProps) {
  return (
    <span
      className={`inline-block min-w-[3rem] px-2 py-0.5 rounded text-sm font-semibold tabular-nums text-center ${
        highlight ? 'bg-rose-100 text-rose-900' : 'text-slate-300'
      }`}
    >
      {value.toFixed(1)}%
    </span>
  );
}
