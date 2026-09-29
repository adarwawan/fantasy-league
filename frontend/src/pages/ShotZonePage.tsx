import { useEffect, useMemo, useState } from 'react';
import { useShotZoneTeam, useShotZoneTeams } from '../hooks/useShotZone';
import { QueryBoundary } from '../components/ui/QueryBoundary';
import { ErrorState } from '../components/ui/ErrorState';
import { PanelSkeleton } from '../components/ui/Skeletons';
import { PageHeader } from '../components/layout/PageHeader';
import { usePageTitle } from '../hooks/usePageTitle';
import { Card } from '../components/ui/Card';
import { FilterBar } from '../components/ui/FilterBar';
import { Field } from '../components/ui/Field';
import { Select } from '../components/ui/Select';
import { SearchInput } from '../components/ui/SearchInput';
import { SegmentedControl } from '../components/ui/SegmentedControl';
import { TeamBadge } from '../components/ui/TeamBadge';
import { ZoneDiagram } from '../components/shotzone/ZoneDiagram';
import { TakenTable } from '../components/shotzone/TakenTable';
import { ConcededStats } from '../components/shotzone/ConcededStats';
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

  usePageTitle('Shot Zones');

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

  const header = (
    <PageHeader
      title="Shot Zones"
      description="Open play only, observed from Understat shot data — who shoots (and scores) from where, and which teams concede the most from each zone."
      meta={teamsData ? `Last ${teamsData.window_matches} matches` : undefined}
    />
  );

  if (teamsLoading || teamsError || teams.length === 0) {
    return (
      <div>
        {header}
        <QueryBoundary
          isLoading={teamsLoading}
          isError={teamsError}
          isEmpty={teams.length === 0}
          onRetry={() => refetchTeams()}
          what="shot-zone data"
          emptyHint="Signals appear once matches have been played and synced."
          skeleton={<PanelSkeleton />}
        >
          {null}
        </QueryBoundary>
      </div>
    );
  }

  return (
    <div>
      {header}

      <FilterBar>
        <Field label="Show">
          <SegmentedControl
            label="Shot direction"
            value={mode}
            onChange={setMode}
            options={[
              { value: 'taken', label: 'Shots taken' },
              { value: 'conceded', label: 'Shots conceded' },
            ]}
          />
        </Field>
        <Field label="Team" htmlFor="shotzone-team">
          <Select
            id="shotzone-team"
            value={team ?? ''}
            onChange={setTeam}
            options={teams.map((t) => ({ value: t, label: t }))}
          />
        </Field>
      </FilterBar>

      {detailLoading && <PanelSkeleton />}
      {detailError && (
        <ErrorState what="this team's shot-zone data" onRetry={() => refetchDetail()} />
      )}

      {detail && (
        <>
          <div className="mb-4">
            <ZoneDiagram totals={zoneTotals} activeTab={zoneTab} onSelect={setZoneTab} />
          </div>

          {mode === 'taken' && (
            <FilterBar>
              <Field label="Search player" htmlFor="shotzone-search">
                <SearchInput id="shotzone-search" value={search} onChange={setSearch} placeholder="e.g. Salah" />
              </Field>
              <Field label="Min shots" htmlFor="shotzone-min">
                <Select
                  id="shotzone-min"
                  value={minShots}
                  onChange={(v) => setMinShots(Number(v))}
                  options={[
                    { value: 0, label: 'Any' },
                    { value: 3, label: '3+' },
                    { value: 5, label: '5+' },
                  ]}
                />
              </Field>
            </FilterBar>
          )}

          <Card variant="panel">
            {team && (
              <div className="flex items-center gap-3 mb-5">
                <TeamBadge name={team} size="lg" />
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
          </Card>
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
