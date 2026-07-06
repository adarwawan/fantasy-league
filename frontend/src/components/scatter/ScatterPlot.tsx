import {
  ScatterChart,
  Scatter,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceArea,
  type TooltipProps,
} from 'recharts';
import { useState } from 'react';
import type { Player } from '../../types/player';
import type { AxisKey } from './AxisSelector';
import { PlayerTooltip } from './PlayerTooltip';
import { PlayerPeekCard } from './PlayerPeekCard';
import { useMediaQuery } from '../../hooks/useMediaQuery';

const POS_COLORS: Record<string, string> = {
  GK:  '#34d399', // emerald — matches PositionBadge
  DEF: '#60a5fa', // blue
  MID: '#a78bfa', // purple
  FWD: '#f87171', // red
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

function dotRadius(ownership: number): number {
  // scale by global_ownership: 0%→4px, 60%+→12px
  const rMin = 4, rMax = 12, clamp = 60;
  return rMin + Math.min(ownership / clamp, 1) * (rMax - rMin);
}

const MUST_HAVE_COLOR = '#fbbf24'; // amber

// 5-point star path centered on (cx, cy), sized to visually match a dot of radius r
function starPath(cx: number, cy: number, r: number): string {
  const outer = r * 1.4;
  const inner = outer * 0.45;
  const points: string[] = [];
  for (let i = 0; i < 10; i++) {
    const radius = i % 2 === 0 ? outer : inner;
    const angle = -Math.PI / 2 + (i * Math.PI) / 5;
    points.push(`${cx + radius * Math.cos(angle)},${cy + radius * Math.sin(angle)}`);
  }
  return `M${points.join('L')}Z`;
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

// Show differential zone overlay when one axis is form and the other is an ownership axis
function isDifferentialView(xAxis: AxisKey, yAxis: AxisKey): boolean {
  const ownershipAxes = new Set<AxisKey>(['global_ownership', 'top_n_ownership']);
  const hasForm = xAxis === 'form' || yAxis === 'form';
  const hasOwnership = ownershipAxes.has(xAxis) || ownershipAxes.has(yAxis);
  return hasForm && hasOwnership;
}

interface Props {
  players:        Player[];
  xAxis:          AxisKey;
  yAxis:          AxisKey;
  onPlayerClick?: (p: Player) => void;
}

export function ScatterPlot({ players, xAxis, yAxis, onPlayerClick }: Props) {
  // On touch/small screens a floating tooltip covers the very dot being
  // inspected, so we dock a peek card to the bottom instead (tap-to-select).
  const isMobile = useMediaQuery('(max-width: 639px)');
  const [peek, setPeek] = useState<PlotPoint | null>(null);

  // On mobile, a dot tap selects (docked card); on desktop it opens the drawer.
  const handleDotTap = (point: PlotPoint) => {
    if (isMobile) setPeek(point);
    else onPlayerClick?.(point.player);
  };

  const byPosition = POSITIONS.reduce<Record<string, PlotPoint[]>>((acc, pos) => {
    acc[pos] = players
      .filter(p => p.position === pos)
      .map(p => {
        const avgFdr = computeAvgFdr(p);
        return {
          x: getValue(p, xAxis, avgFdr),
          y: getValue(p, yAxis, avgFdr),
          r: dotRadius(p.global_ownership),
          player: p,
          avgFdr,
        };
      });
    return acc;
  }, {} as Record<string, PlotPoint[]>);

  const showZone = isDifferentialView(xAxis, yAxis);

  // Differential zone: high form + low ownership quadrant
  // Always compute so recharts gets real numbers (never undefined)
  const formIsX = xAxis === 'form';
  const allX = players.map(p => getValue(p, xAxis, computeAvgFdr(p)));
  const allY = players.map(p => getValue(p, yAxis, computeAvgFdr(p)));
  const minX = allX.length ? Math.min(...allX) : 0;
  const maxX = allX.length ? Math.max(...allX) : 10;
  const minY = allY.length ? Math.min(...allY) : 0;
  const maxY = allY.length ? Math.max(...allY) : 10;
  const midX = (minX + maxX) / 2;
  const midY = (minY + maxY) / 2;

  // When form is X: high form = right half (x>mid), low ownership = bottom half (y<mid)
  // When form is Y: high form = top half (y>mid), low ownership = left half (x<mid)
  const zoneX1 = formIsX ? midX : minX;
  const zoneX2 = formIsX ? maxX  : midX;
  const zoneY1 = formIsX ? minY  : midY;
  const zoneY2 = formIsX ? midY  : maxY;

  return (
    <div className="w-full">
      {/* Legend */}
      <div className="flex gap-4 mb-3 justify-end flex-wrap">
        {POSITIONS.map(pos => (
          <div key={pos} className="flex items-center gap-1.5 text-xs text-slate-400">
            <span className="inline-block w-2.5 h-2.5 rounded-full" style={{ background: POS_COLORS[pos] }} />
            {pos}
          </div>
        ))}
        {showZone && (
          <div className="flex items-center gap-1.5 text-xs text-emerald-400 ml-2 pl-2 border-l border-slate-700">
            <span className="inline-block w-3 h-3 rounded bg-emerald-400/20 border border-emerald-400/40" />
            Differential zone
          </div>
        )}
        <div className="flex items-center gap-1.5 text-xs text-amber-400 ml-2 pl-2 border-l border-slate-700">
          <svg width="12" height="12" viewBox="-12 -12 24 24" aria-hidden="true">
            <path d={starPath(0, 0, 7)} fill="none" stroke={MUST_HAVE_COLOR} strokeWidth={2} />
          </svg>
          Must-have
        </div>
        <div className="flex items-center gap-1.5 text-xs text-slate-500 ml-2 pl-2 border-l border-slate-700">
          dot size = ownership
        </div>
      </div>

      <ResponsiveContainer width="100%" height={520}>
        <ScatterChart margin={{ top: 10, right: 20, bottom: 20, left: 10 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
          <XAxis
            dataKey="x"
            type="number"
            name={xAxis}
            domain={['auto', 'auto']}
            tick={{ fontSize: 11, fill: '#94a3b8' }}
            stroke="#475569"
          />
          <YAxis
            dataKey="y"
            type="number"
            name={yAxis}
            domain={['auto', 'auto']}
            tick={{ fontSize: 11, fill: '#94a3b8' }}
            stroke="#475569"
          />
          {/* Floating tooltip on hover-capable devices only; mobile uses the docked peek card */}
          {!isMobile && (
            <Tooltip content={<CustomTooltip />} cursor={{ strokeDasharray: '3 3', stroke: '#64748b' }} />
          )}

          {/* Differential zone overlay — high form, low ownership quadrant */}
          {showZone && (
            <ReferenceArea
              x1={zoneX1}
              x2={zoneX2}
              y1={zoneY1}
              y2={zoneY2}
              fill="#34d399"
              fillOpacity={0.07}
              stroke="#34d399"
              strokeOpacity={0.3}
              strokeDasharray="4 4"
              label={{ value: 'Differential zone', position: 'insideTopRight', fontSize: 10, fill: '#34d399', opacity: 0.7 }}
            />
          )}

          {POSITIONS.map(pos => (
            <Scatter
              key={pos}
              name={pos}
              data={byPosition[pos]}
              fill={POS_COLORS[pos]}
              fillOpacity={0.8}
              shape={(props: unknown) => {
                const { cx, cy, payload } = props as { cx: number; cy: number; payload: PlotPoint };
                const interactive = {
                  style: onPlayerClick ? { cursor: 'pointer' } : undefined,
                  onClick: onPlayerClick ? () => handleDotTap(payload) : undefined,
                };
                if (payload.player.must_have) {
                  return (
                    <path
                      key={payload.player.id}
                      d={starPath(cx, cy, payload.r)}
                      fill={POS_COLORS[pos]}
                      fillOpacity={0.95}
                      stroke={MUST_HAVE_COLOR}
                      strokeWidth={1.5}
                      {...interactive}
                    />
                  );
                }
                return (
                  <circle
                    key={payload.player.id}
                    cx={cx}
                    cy={cy}
                    r={payload.r}
                    fill={POS_COLORS[pos]}
                    fillOpacity={0.8}
                    stroke="#1e293b"
                    strokeWidth={1.5}
                    {...interactive}
                  />
                );
              }}
            />
          ))}
        </ScatterChart>
      </ResponsiveContainer>

      {peek && (
        <PlayerPeekCard
          player={peek.player}
          avgFdr={peek.avgFdr}
          onDetails={() => {
            onPlayerClick?.(peek.player);
            setPeek(null);
          }}
          onClose={() => setPeek(null)}
        />
      )}
    </div>
  );
}
