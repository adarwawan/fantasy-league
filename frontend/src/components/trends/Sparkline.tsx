interface SparklineProps {
  values: number[];
  width?: number;
  height?: number;
  className?: string;
}

// Sparkline draws a normalised polyline of the given values. Used for a row's
// net-transfer trajectory over the session.
export function Sparkline({ values, width = 96, height = 24, className }: SparklineProps) {
  if (values.length < 2) {
    return <span className="text-xs text-slate-600">—</span>;
  }
  const min = Math.min(...values);
  const max = Math.max(...values);
  const span = max - min || 1;
  const stepX = width / (values.length - 1);
  const points = values
    .map((v, i) => {
      const x = i * stepX;
      const y = height - ((v - min) / span) * height;
      return `${x.toFixed(1)},${y.toFixed(1)}`;
    })
    .join(' ');
  const rising = values[values.length - 1] >= values[0];

  return (
    <svg width={width} height={height} className={className} aria-hidden>
      <polyline
        points={points}
        fill="none"
        stroke={rising ? 'rgb(52 211 153)' : 'rgb(248 113 113)'}
        strokeWidth={1.5}
        strokeLinejoin="round"
        strokeLinecap="round"
      />
    </svg>
  );
}
