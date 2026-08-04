import { useEffect, useState } from 'react';
import { useTrendsLeaders, useTrendsSession } from '../hooks/useTrends';
import { LeaderBoard } from '../components/trends/LeaderBoard';
import { ErrorState } from '../components/common/ErrorState';
import type { TrendsSession } from '../types/trends';

const WINDOWS = [
  { label: '30 min', value: '30m' },
  { label: '1 hour', value: '1h' },
  { label: '3 hours', value: '3h' },
];

const TOP_N = 15;

function isActive(s: unknown): s is TrendsSession {
  return !!s && (s as TrendsSession).active === true;
}

export function TrendsPage() {
  const [window, setWindow] = useState('30m');
  const [expanded, setExpanded] = useState<number | null>(null);

  const { data: session } = useTrendsSession();
  const inflows = useTrendsLeaders(window, 'in', TOP_N);
  const outflows = useTrendsLeaders(window, 'out', TOP_N);

  useEffect(() => {
    document.title = 'Trends — Transfer Velocity';
  }, []);

  const active = isActive(session);
  const isLoading = inflows.isLoading || outflows.isLoading;
  const isError = inflows.isError || outflows.isError;
  const metric = inflows.data?.metric ?? 'transfers';
  const toggle = (id: number) => setExpanded((cur) => (cur === id ? null : id));

  return (
    <div>
      <div className="mb-4">
        <h1 className="text-xl font-semibold text-slate-100">Trends</h1>
        <p className="text-sm text-slate-400 mt-1 max-w-2xl">
          Live transfer velocity in the hours before the deadline. Movers are ranked by the change
          in net transfers over the selected window — a surge that accelerates near the deadline is
          the classic "news is leaking" pattern, in either direction.
        </p>
      </div>

      {active && (
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div className="text-sm text-slate-400">
            <span className="text-slate-200 font-medium">Gameweek {session.gameweek}</span>
            {' · '}
            <span className="text-slate-500">{session.poll_count} polls captured</span>
            {' · '}
            <span
              className="text-slate-500"
              title={
                metric === 'ownership'
                  ? 'GW1 has no transfers before the deadline, so movement is ranked by ownership change.'
                  : 'Ranked by net transfer change over the window.'
              }
            >
              by {metric === 'ownership' ? 'ownership Δ' : 'net transfers'}
            </span>
          </div>
          <div className="inline-flex rounded-md bg-slate-900 border border-slate-700/60 p-0.5">
            {WINDOWS.map((w) => (
              <button
                key={w.value}
                onClick={() => setWindow(w.value)}
                className={`px-3 py-1 text-sm rounded transition-colors ${
                  window === w.value
                    ? 'bg-slate-700/70 text-white'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                {w.label}
              </button>
            ))}
          </div>
        </div>
      )}

      {!active && (
        <div className="rounded-lg border border-slate-700/60 bg-slate-900/40 px-6 py-10 text-center">
          <div className="text-2xl mb-2">⏱️</div>
          <div className="text-slate-200 font-medium">No recording in progress</div>
          <p className="text-sm text-slate-400 mt-1">
            Transfer recording starts in the ~24 hours before the next deadline.
          </p>
        </div>
      )}

      {active && isError && (
        <ErrorState
          message="Failed to load transfer trends."
          onRetry={() => {
            inflows.refetch();
            outflows.refetch();
          }}
        />
      )}

      {active && isLoading && (
        <div className="text-sm text-slate-500 py-8 text-center">Loading fastest movers…</div>
      )}

      {active && !isLoading && !isError && (
        <div className="grid gap-4 lg:grid-cols-2 items-start">
          <LeaderBoard
            title="Top 15 Inflows"
            accent="in"
            metric={metric}
            rows={inflows.data?.leaders ?? []}
            expanded={expanded}
            onToggle={toggle}
          />
          <LeaderBoard
            title="Top 15 Outflows"
            accent="out"
            metric={metric}
            rows={outflows.data?.leaders ?? []}
            expanded={expanded}
            onToggle={toggle}
          />
        </div>
      )}
    </div>
  );
}
