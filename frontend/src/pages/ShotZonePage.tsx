import { useEffect, useMemo, useState } from 'react';
import { useShotZoneTeam, useShotZoneTeams } from '../hooks/useShotZone';
import { ErrorState } from '../components/common/ErrorState';
import { ZoneDiagram } from '../components/shotzone/ZoneDiagram';
import { TakenTable } from '../components/shotzone/TakenTable';
import { ConcededStats } from '../components/shotzone/ConcededStats';
import { teamMeta, readableText } from '../components/setpiece/teamMeta';
import type { Zone, ZoneTab } from '../types/shotzone';

type Mode = 'taken' | 'conceded';

const ZONES: Zone[] = ['center', 'left', 'right', 'outside'];

export function ShotZonePage() {
  const { data: teamsData, isLoading: teamsLoading, isError: teamsError, refetch: refetchTeams } = useShotZoneTeams();

  const [mode, setMode] = useState<Mode>('taken');
  const [team, setTeam] = useState<string | undefined>(undefined);
  const [zoneTab, setZoneTab] = useState<ZoneTab>('overall');
  const [search, setSearch] = useState('');
  const [minShots, setMinShots] = useState(0);

  const teams = teamsData?.teams ?? [];

  useEffect(() => {
    document.title = 'Shot Zone Rankings — Understat';
  }, []);

  // Default to the first team once the list loads.
  useEffect(() => {
    if (!team && teams.length > 0) setTeam(teams[0]);
  }, [team, teams]);

  const { data: detail, isLoading: detailLoading, isError: detailError, refetch: refetchDetail } = useShotZoneTeam(team);

  const zoneTotals = useMemo(() => {
    const empty = { shots: 0, goals: 0, xg: 0 };
    const totals: Record<Zone, { shots: number; goals: number; xg: number }> = {
      center: { ...empty }, left: { ...empty }, right: { ...empty }, outside: { ...empty },
    };
    if (!detail) return totals;
    if (mode === 'taken') {
      for (const zone of ZONES) {
        const rows = detail.taken[zone];
        totals[zone] = rows.reduce(
          (acc, r) => ({ shots: acc.shots + r.shots, goals: acc.goals + r.goals, xg: acc.xg + r.xg }),
          { shots: 0, goals: 0, xg: 0 },
        );
      }
    } else {
      for (const zone of ZONES) totals[zone] = detail.conceded[zone];
    }
    return totals;
  }, [detail, mode]);

  if (teamsLoading) return <LoadingState />;
  if (teamsError) {
    return (
      <ErrorState
        message="Failed to load shot-zone data. The source may be temporarily unavailable."
        onRetry={() => refetchTeams()}
      />
    );
  }
  if (teams.length === 0) return <EmptyState />;

  const meta = team ? teamMeta(team) : undefined;

  return (
    <div>
      <div className="mb-4">
        <h1 className="text-xl font-semibold text-slate-100">Shot Zone Rankings</h1>
        <p className="text-sm text-slate-400 mt-1 max-w-2xl">
          Open play only, observed from Understat shot data — who shoots (and scores) from where, and which
          teams concede the most from each zone. Last {teamsData?.window_matches ?? 6} matches.
        </p>
      </div>

      <div className="flex flex-wrap items-center gap-3 mb-4">
        <div className="inline-flex rounded-lg border border-slate-700/60 bg-slate-900 p-0.5">
          {(['taken', 'conceded'] as Mode[]).map((m) => (
            <button
              key={m}
              onClick={() => setMode(m)}
              className={`px-3 py-1.5 text-sm font-medium rounded-md transition-colors ${
                mode === m ? 'bg-violet-600 text-white' : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              {m === 'taken' ? 'Shots taken' : 'Shots conceded'}
            </button>
          ))}
        </div>

        <select
          value={team ?? ''}
          onChange={(e) => setTeam(e.target.value)}
          className="px-3 py-1.5 text-sm rounded-md bg-slate-900 border border-slate-700/60 text-slate-100 focus:outline-none focus:ring-2 focus:ring-violet-500/50"
        >
          {teams.map((t) => (
            <option key={t} value={t}>{t}</option>
          ))}
        </select>
      </div>

      {detailLoading && <LoadingState />}
      {detailError && (
        <ErrorState message="Failed to load this team's shot-zone data." onRetry={() => refetchDetail()} />
      )}

      {detail && (
        <>
          <div className="mb-4">
            <ZoneDiagram totals={zoneTotals} activeTab={zoneTab} onSelect={setZoneTab} />
          </div>

          {mode === 'taken' && (
            <div className="flex flex-wrap gap-2 mb-4">
              <input
                type="search"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search player…"
                className="w-full sm:w-48 px-3 py-1.5 text-sm rounded-md bg-slate-900 border border-slate-700/60 text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-violet-500/50"
              />
              <select
                value={minShots}
                onChange={(e) => setMinShots(Number(e.target.value))}
                className="px-3 py-1.5 text-sm rounded-md bg-slate-900 border border-slate-700/60 text-slate-100 focus:outline-none focus:ring-2 focus:ring-violet-500/50"
              >
                <option value={0}>Min shots: any</option>
                <option value={3}>Min shots: 3+</option>
                <option value={5}>Min shots: 5+</option>
              </select>
            </div>
          )}

          <div className="rounded-2xl border border-slate-700/60 bg-slate-900 p-5 sm:p-6">
            {meta && (
              <div className="flex items-center gap-3 mb-5">
                <span
                  className="flex items-center justify-center h-11 w-11 rounded-xl text-sm font-bold tracking-wide shrink-0"
                  style={{ backgroundColor: meta.color, color: readableText(meta.color) }}
                >
                  {meta.code}
                </span>
                <div className="min-w-0">
                  <h2 className="text-lg font-bold text-slate-100 leading-tight">{team}</h2>
                  <p className="text-xs text-slate-500">
                    {mode === 'conceded' ? 'conceded · ' : ''}{zoneLabel(zoneTab)} · observed · last {detail.window_matches} matches
                  </p>
                </div>
              </div>
            )}

            {mode === 'taken' ? (
              <TakenTable rows={detail.taken[zoneTab]} search={search} minShots={minShots} />
            ) : (
              <ConcededStats totals={detail.conceded[zoneTab]} />
            )}
          </div>
        </>
      )}
    </div>
  );
}

function zoneLabel(tab: ZoneTab): string {
  switch (tab) {
    case 'overall': return 'overall';
    case 'center':  return 'center inside box';
    case 'left':    return 'left flank';
    case 'right':   return 'right flank';
    case 'outside': return 'outside box';
  }
}

function LoadingState() {
  return (
    <div className="rounded-2xl border border-slate-700/60 bg-slate-900 p-6 h-72 animate-pulse">
      <div className="flex items-center gap-3 mb-6">
        <div className="h-11 w-11 rounded-xl bg-slate-700" />
        <div className="h-4 w-32 rounded bg-slate-700" />
      </div>
      <div className="h-48 rounded-xl bg-slate-800" />
    </div>
  );
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center gap-2 py-16 text-center">
      <span className="text-3xl" aria-hidden="true">🎯</span>
      <p className="text-slate-300 text-sm">No shot-zone data yet.</p>
      <p className="text-slate-500 text-xs">Signals appear once matches have been played and synced.</p>
    </div>
  );
}
