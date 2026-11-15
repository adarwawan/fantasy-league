import {
  BarChart,
  Bar,
  Cell,
  LabelList,
  ReferenceLine,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  type TooltipProps,
} from 'recharts';
import { teamNameByCode } from '../setpiece/teamMeta';
import { Card } from '../ui/Card';
import type { TeamBonusEntry } from '../../types/stats';
import { NEUTRAL_BONUS_PER_MATCH, bonusPerMatch, sortByBonusRate } from '../../utils/bonus';

const ABOVE = '#38bdf8'; // sky-400: above the neutral line
const BELOW = '#64748b'; // slate-500: at or below it
const ROW_PX = 26;

interface Row {
  team: string;
  name: string;
  rate: number;
  bonus: number;
  matches: number;
}

function ChartTooltip({ active, payload }: TooltipProps<number, string>) {
  const row = payload?.[0]?.payload as Row | undefined;
  if (!active || !row) return null;
  return (
    <div className="rounded-lg border border-line bg-surface px-2.5 py-1.5 text-xs text-slate-200 shadow-lg">
      <div className="font-semibold">{row.name}</div>
      <div className="tabular-nums text-slate-400">
        {row.rate.toFixed(2)} per match · {row.bonus} bonus in {row.matches} matches
      </div>
    </div>
  );
}

/**
 * Bonus per match for every club, sorted high to low, against the neutral
 * 3.0 line (6 points are handed out per match, shared by both sides). Rate,
 * not volume, is the headline: clubs can have played different match counts.
 */
export function TeamBonusChart({ teams }: { teams: TeamBonusEntry[] }) {
  const rows: Row[] = sortByBonusRate(teams).map((t) => ({
    team: t.team,
    name: teamNameByCode(t.team),
    rate: Math.round(bonusPerMatch(t) * 100) / 100,
    bonus: t.bonus,
    matches: t.matches,
  }));
  const max = Math.max(6, Math.ceil(Math.max(0, ...rows.map((r) => r.rate))));

  return (
    <Card>
      <div className="mb-2 flex items-baseline justify-between gap-2">
        <h3 className="text-sm font-semibold text-slate-100">Bonus per match</h3>
        <span className="text-[11px] text-slate-500">neutral = {NEUTRAL_BONUS_PER_MATCH.toFixed(1)}</span>
      </div>
      <div
        style={{ height: rows.length * ROW_PX + 28 }}
        role="img"
        aria-label={`Bonus points per match for ${rows.length} clubs, highest first`}
      >
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={rows} layout="vertical" margin={{ top: 0, right: 36, bottom: 0, left: 0 }} barCategoryGap={4}>
            <XAxis
              type="number"
              domain={[0, max]}
              allowDecimals={false}
              tickCount={max + 1}
              tick={{ fill: '#64748b', fontSize: 10 }}
              axisLine={false}
              tickLine={false}
            />
            <YAxis
              type="category"
              dataKey="name"
              width={112}
              tick={{ fill: '#cbd5e1', fontSize: 11 }}
              axisLine={false}
              tickLine={false}
              interval={0}
            />
            <Tooltip content={<ChartTooltip />} cursor={{ fill: 'rgba(148,163,184,0.08)' }} />
            <ReferenceLine x={NEUTRAL_BONUS_PER_MATCH} stroke="#94a3b8" strokeDasharray="4 3" />
            <Bar dataKey="rate" radius={[0, 4, 4, 0]} maxBarSize={14} isAnimationActive={false}>
              {rows.map((r) => (
                <Cell key={r.team} fill={r.rate > NEUTRAL_BONUS_PER_MATCH ? ABOVE : BELOW} />
              ))}
              <LabelList
                dataKey="rate"
                position="right"
                formatter={(v: number) => v.toFixed(1)}
                style={{ fill: '#e2e8f0', fontSize: 11 }}
              />
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>
    </Card>
  );
}
