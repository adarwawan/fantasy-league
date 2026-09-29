import type { Zone, ZoneTab } from '../../types/shotzone';

interface ZoneTotals { shots: number; goals: number; xg: number }

interface Props {
  totals:     Record<Zone, ZoneTotals>;
  activeTab:  ZoneTab;
  onSelect:   (tab: ZoneTab) => void;
}

// The pitch diagram itself is the zone filter: it shows the selected team's
// shots/goals/xG per zone at all times, and clicking a zone both highlights
// it and filters the panel below — no separate tab list.
const ZONE_DEFS: Array<{ zone: Zone; x: number; width: number; labelColor: string; fill: string }> = [
  { zone: 'left',   x: 50,  width: 73, labelColor: '#bae6fd', fill: '#38bdf8' },
  { zone: 'center', x: 123, width: 74, labelColor: '#ede9fe', fill: '#a78bfa' },
  { zone: 'right',  x: 197, width: 73, labelColor: '#ffe4e6', fill: '#fb7185' },
];

const ZONE_LABEL: Record<Zone, string> = {
  center:  'center inside box',
  left:    'left flank',
  right:   'right flank',
  outside: 'outside box',
};

export function ZoneDiagram({ totals, activeTab, onSelect }: Props) {
  return (
    <div className="rounded-xl border border-slate-700/60 bg-slate-900/60 p-3">
      <div className="flex justify-end mb-2">
        <button
          onClick={() => onSelect('overall')}
          className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-colors ${
            activeTab === 'overall' ? 'bg-accent text-white' : 'bg-slate-800 text-slate-300 hover:bg-slate-700'
          }`}
        >
          Overall
        </button>
      </div>

      <svg viewBox="0 0 320 232" className="w-full max-w-md mx-auto block">
        <rect x="10" y="4" width="300" height="222" fill="none" stroke="#334155" strokeWidth="1" />
        <line x1="130" y1="4" x2="190" y2="4" stroke="#94a3b8" strokeWidth="3" />

        {ZONE_DEFS.map(({ zone, x, width, labelColor, fill }) => {
          const active = activeTab === zone;
          const opacity = active ? 0.75 : activeTab === 'overall' ? 0.4 : 0.12;
          const t = totals[zone];
          return (
            <g
              key={zone}
              onClick={() => onSelect(zone)}
              className="cursor-pointer"
              role="button"
              aria-pressed={active}
            >
              <rect
                x={x} y={14} width={width} height={72}
                fill={fill} fillOpacity={opacity} stroke={fill} strokeWidth={active ? 2 : 1}
              />
              <text x={x + width / 2} y={28} textAnchor="middle" fontSize="9" fill={labelColor}>
                {ZONE_LABEL[zone]}
              </text>
              <text x={x + width / 2} y={52} textAnchor="middle" fontSize="20" fontWeight={500} fill="#f8fafc">
                {t.shots}
              </text>
              <text x={x + width / 2} y={68} textAnchor="middle" fontSize="9" fill={labelColor}>
                {t.goals}G &middot; {t.xg.toFixed(1)}xG
              </text>
            </g>
          );
        })}

        <g
          onClick={() => onSelect('outside')}
          className="cursor-pointer"
          role="button"
          aria-pressed={activeTab === 'outside'}
        >
          <rect
            x={10} y={94} width={300} height={132}
            fill="#475569"
            fillOpacity={activeTab === 'outside' ? 0.75 : activeTab === 'overall' ? 0.22 : 0.12}
            stroke="#475569"
            strokeWidth={activeTab === 'outside' ? 2 : 1}
          />
          <text x={160} y={118} textAnchor="middle" fontSize="10" fill="#cbd5e1">
            outside box &middot; any width, deeper
          </text>
          <text x={160} y={158} textAnchor="middle" fontSize="26" fontWeight={500} fill="#f1f5f9">
            {totals.outside.shots}
          </text>
          <text x={160} y={180} textAnchor="middle" fontSize="11" fill="#cbd5e1">
            {totals.outside.goals}G &middot; {totals.outside.xg.toFixed(1)}xG
          </text>
        </g>
      </svg>
    </div>
  );
}
