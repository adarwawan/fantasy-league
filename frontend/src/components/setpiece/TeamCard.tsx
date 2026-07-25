import type { SetPieceTeam, TakerRow, TargetRow } from '../../types/setpiece';
import { teamMeta, readableText } from './teamMeta';

interface Props {
  team:          SetPieceTeam;
  windowMatches: number;
  updatedAt?:    string;
}

export function TeamCard({ team, windowMatches, updatedAt }: Props) {
  const meta = teamMeta(team.team);

  return (
    <div className="rounded-2xl border border-slate-700/60 bg-slate-900 p-5 sm:p-6 w-full h-full">
      {/* Header */}
      <div className="flex items-center gap-3 mb-5">
        <span
          className="flex items-center justify-center h-11 w-11 rounded-xl text-sm font-bold tracking-wide shrink-0"
          style={{ backgroundColor: meta.color, color: readableText(meta.color) }}
        >
          {meta.code}
        </span>
        <div className="min-w-0">
          <h2 className="text-lg font-bold text-slate-100 leading-tight">{team.team}</h2>
          <p className="text-xs text-slate-500">
            observed · last {windowMatches} matches
            {updatedAt && <> · updated {timeAgo(updatedAt)}</>}
          </p>
        </div>
      </div>

      {/* Takers */}
      <SectionHeading title="Takers" />
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-6">
        <TakerSubCard title="Penalties" rows={team.takers.penalty} emptyLabel="No observed penalties" />
        <TakerSubCard title="Direct free kicks" rows={team.takers.dfk} emptyLabel="No observed free kicks" />
      </div>

      {/* Target men */}
      <SectionHeading
        title="Set-piece target men"
        note="— shots off corners & set-piece free kicks"
      />
      <TargetTable rows={team.targets} team={team} />

      <p className="mt-3 text-[11px] leading-relaxed text-slate-500">
        target man = who attacks the ball, not who delivers it
      </p>
    </div>
  );
}

function SectionHeading({ title, note }: { title: string; note?: string }) {
  return (
    <div className="flex items-baseline gap-2 mb-3">
      <span className="inline-block h-3.5 w-3.5 rounded-sm ring-1 ring-slate-500 shrink-0 translate-y-0.5" />
      <h3 className="text-base font-semibold text-slate-100">{title}</h3>
      {note && <span className="text-xs text-slate-500">{note}</span>}
    </div>
  );
}

// --- Takers ---------------------------------------------------------------

function TakerSubCard({ title, rows, emptyLabel }: { title: string; rows: TakerRow[]; emptyLabel: string }) {
  const primary = rows[0];
  const backups = rows.slice(1);
  const total   = rows.reduce((sum, r) => sum + r.attempts, 0);

  return (
    <div className="rounded-xl border border-slate-700/50 bg-slate-800/30 p-3.5">
      <div className="text-xs text-slate-500 mb-1.5">{title}</div>

      {!primary && <div className="text-sm text-slate-500">{emptyLabel}</div>}

      {primary && (
        <>
          <div className="font-bold text-slate-100">{primary.player_name}</div>
          <div className="text-xs text-slate-400 mt-1">
            {primary.attempts} of {total} · {primary.goals} scored
            {primary.last_taken && <> · last {fmtDate(primary.last_taken)}</>}
          </div>
          {backups.length > 0 && (
            <div className="mt-2 space-y-0.5">
              {backups.map((b) => (
                <div key={b.player_id} className="text-xs text-slate-500">
                  {b.player_name} · {b.attempts}
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  );
}

// --- Target-men table -----------------------------------------------------

// dutySplit is a player's shot breakdown by set-piece source.
interface DutySplit { ck: number; fk: number }

// buildSplits maps player_id -> {corner shots, fk shots} from the per-duty rows.
function buildSplits(team: SetPieceTeam): Map<string, DutySplit> {
  const m = new Map<string, DutySplit>();
  const bump = (id: string, key: keyof DutySplit, n: number) => {
    const cur = m.get(id) ?? { ck: 0, fk: 0 };
    cur[key] = n;
    m.set(id, cur);
  };
  for (const r of team.targets_by_duty.corner ?? [])   bump(r.player_id, 'ck', r.shots);
  for (const r of team.targets_by_duty.setpiece ?? []) bump(r.player_id, 'fk', r.shots);
  return m;
}

function TargetTable({ rows, team }: { rows: TargetRow[]; team: SetPieceTeam }) {
  if (rows.length === 0) {
    return <div className="text-sm text-slate-500 py-2">No observed set-piece shots in the window.</div>;
  }
  const maxXg  = Math.max(...rows.map((r) => r.xg), 0.01);
  const splits = buildSplits(team);

  return (
    <div className="rounded-xl border border-slate-700/50 overflow-hidden">
      <table className="w-full text-sm">
        <thead>
          <tr className="text-xs text-slate-500 border-b border-slate-700/60">
            <th className="text-left font-medium px-3 py-2 w-8">#</th>
            <th className="text-left font-medium px-3 py-2">Player</th>
            <th className="text-left font-medium px-3 py-2">Source <span className="text-slate-600">· xG</span></th>
            <th className="text-right font-medium px-3 py-2">Shots</th>
            <th className="text-right font-medium px-3 py-2">G</th>
            <th className="text-right font-medium px-3 py-2">Head%</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r, i) => {
            const split = splits.get(r.player_id) ?? { ck: 0, fk: 0 };
            return (
              <tr key={r.player_id} className="border-b border-slate-700/40 last:border-0">
                <td className="px-3 py-2.5 text-slate-500 tabular-nums">{i + 1}</td>
                <td className="px-3 py-2.5">
                  <span className="font-semibold text-slate-100">{r.player_name}</span>
                </td>
                <td className="px-3 py-2.5">
                  <SplitVolumeBar xg={r.xg} max={maxXg} split={split} />
                </td>
                <td className="px-3 py-2.5 text-right tabular-nums text-slate-200">{r.shots}</td>
                <td className={`px-3 py-2.5 text-right tabular-nums ${r.goals > 0 ? 'text-emerald-400 font-semibold' : 'text-slate-500'}`}>
                  {r.goals}
                </td>
                <td className="px-3 py-2.5 text-right tabular-nums text-slate-300">
                  {r.header_pct != null ? `${Math.round(r.header_pct)}%` : '—'}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>

      {/* Legend for the source split. */}
      <div className="flex items-center gap-3 px-3 py-2 text-[11px] text-slate-500 border-t border-slate-700/40">
        <span className="flex items-center gap-1.5"><span className="h-2.5 w-2.5 rounded-sm bg-sky-500" /> corner</span>
        <span className="flex items-center gap-1.5"><span className="h-2.5 w-2.5 rounded-sm bg-violet-400" /> set-piece FK</span>
      </div>
    </div>
  );
}

// SplitVolumeBar renders a bar whose total length encodes xG (volume, relative
// to the table max) and whose internal split shows corner vs set-piece-FK shots.
function SplitVolumeBar({ xg, max, split }: { xg: number; max: number; split: DutySplit }) {
  const total = Math.max(4, Math.round((xg / max) * 100)); // bar length as % of column
  const shots = split.ck + split.fk;
  const ckPct = shots > 0 ? (split.ck / shots) * total : total;
  const fkPct = total - ckPct;
  const title = `${split.ck} CK · ${split.fk} FK · xG ${xg.toFixed(2)}`;

  return (
    <div className="flex items-center gap-2 min-w-[110px]">
      <div className="flex-1 h-2.5 rounded-full bg-slate-700/50 overflow-hidden flex" title={title}>
        {shots > 0 ? (
          <>
            <div className="h-full bg-sky-500" style={{ width: `${ckPct}%` }} />
            <div className="h-full bg-violet-400" style={{ width: `${fkPct}%` }} />
          </>
        ) : (
          <div className="h-full bg-slate-500" style={{ width: `${total}%` }} />
        )}
      </div>
      <span className="tabular-nums text-xs text-slate-400 w-9 text-right">{xg.toFixed(2)}</span>
    </div>
  );
}

// --- utils ----------------------------------------------------------------

function fmtDate(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}

function timeAgo(iso: string): string {
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return '';
  const mins = Math.max(0, Math.round((Date.now() - then) / 60000));
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.round(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  const days = Math.round(hrs / 24);
  return `${days}d ago`;
}
