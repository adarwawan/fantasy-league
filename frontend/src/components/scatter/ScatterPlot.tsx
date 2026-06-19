import {
  ScatterChart,
  Scatter,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  type TooltipProps,
} from 'recharts';
import type { Player } from '../../types/player';
import type { AxisKey } from './AxisSelector';
import { PlayerTooltip } from './PlayerTooltip';

const POS_COLORS: Record<string, string> = {
  GK:  '#f59e0b',
  DEF: '#3b82f6',
  MID: '#22c55e',
  FWD: '#ef4444',
};

function computeAvgFdr(player: Player): number {
  const slice = player.fixtures.slice(0, 3);
  if (!slice.length) return 3;
  return slice.reduce((s, f) => s + f.difficulty, 0) / slice.length;
}

function getValue(player: Player, axis: AxisKey, avgFdr: number): number {
  switch (axis) {
    case 'global_ownership': return player.global_ownership;
    case 'top_n_ownership':  return player.top_n_ownership;
    case 'form':             return player.form;
    case 'avg_fdr':          return avgFdr;
  }
}

function dotRadius(price: number): number {
  const min = 4, max = 15, rMin = 4, rMax = 10;
  return rMin + ((price - min) / (max - min)) * (rMax - rMin);
}

interface PlotPoint {
  x:      number;
  y:      number;
  r:      number;
  player: Player;
  avgFdr: number;
}

function CustomTooltip({ active, payload }: TooltipProps<number, string>) {
  if (!active || !payload?.length) return null;
  const { player, avgFdr } = payload[0].payload as PlotPoint;
  return <PlayerTooltip player={player} avgFdr={avgFdr} />;
}

const POSITIONS = ['GK', 'DEF', 'MID', 'FWD'] as const;

interface Props {
  players: Player[];
  xAxis:   AxisKey;
  yAxis:   AxisKey;
}

export function ScatterPlot({ players, xAxis, yAxis }: Props) {
  const byPosition = POSITIONS.reduce<Record<string, PlotPoint[]>>((acc, pos) => {
    acc[pos] = players
      .filter(p => p.position === pos)
      .map(p => {
        const avgFdr = computeAvgFdr(p);
        return {
          x: getValue(p, xAxis, avgFdr),
          y: getValue(p, yAxis, avgFdr),
          r: dotRadius(p.price),
          player: p,
          avgFdr,
        };
      });
    return acc;
  }, {} as Record<string, PlotPoint[]>);

  return (
    <div className="w-full">
      {/* Legend */}
      <div className="flex gap-4 mb-3 justify-end">
        {POSITIONS.map(pos => (
          <div key={pos} className="flex items-center gap-1.5 text-xs text-gray-600">
            <span
              className="inline-block w-2.5 h-2.5 rounded-full"
              style={{ background: POS_COLORS[pos] }}
            />
            {pos}
          </div>
        ))}
      </div>
      <ResponsiveContainer width="100%" height={520}>
        <ScatterChart margin={{ top: 10, right: 20, bottom: 20, left: 10 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
          <XAxis
            dataKey="x"
            type="number"
            name={xAxis}
            domain={['auto', 'auto']}
            tick={{ fontSize: 11 }}
          />
          <YAxis
            dataKey="y"
            type="number"
            name={yAxis}
            domain={['auto', 'auto']}
            tick={{ fontSize: 11 }}
          />
          <Tooltip content={<CustomTooltip />} cursor={{ strokeDasharray: '3 3' }} />
          {POSITIONS.map(pos => (
            <Scatter
              key={pos}
              name={pos}
              data={byPosition[pos]}
              fill={POS_COLORS[pos]}
              fillOpacity={0.75}
              shape={(props: unknown) => {
                const { cx, cy, payload } = props as { cx: number; cy: number; payload: PlotPoint };
                return (
                <circle
                  key={`${payload.player.id}`}
                  cx={cx}
                  cy={cy}
                  r={payload.r}
                  fill={POS_COLORS[pos]}
                  fillOpacity={0.75}
                  stroke="white"
                  strokeWidth={1}
                />
              );}}
            />
          ))}
        </ScatterChart>
      </ResponsiveContainer>
    </div>
  );
}
