interface GoalsCellProps {
  value:     number;
  highlight: boolean;
}

export function GoalsCell({ value, highlight }: GoalsCellProps) {
  return (
    <span
      className={`inline-block min-w-[3rem] px-2 py-0.5 rounded text-sm font-semibold tabular-nums text-center ${
        highlight ? 'bg-blue-200 text-blue-900' : 'text-slate-300'
      }`}
    >
      {value.toFixed(2)}
    </span>
  );
}
