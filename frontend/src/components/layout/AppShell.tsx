import { Link, useParams, useLocation } from 'react-router-dom';
import { GameSwitcher } from './GameSwitcher';
import { GWContextBar } from './GWContextBar';

const NAV_LINKS = [
  { to: 'players', label: 'Players' },
  { to: 'teams',   label: 'Teams'   },
  { to: 'scatter', label: 'Scatter' },
];

export function AppShell({ children }: { children: React.ReactNode }) {
  const { game = 'fpl' } = useParams<{ game: string }>();
  const { pathname } = useLocation();

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100">
      <header className="bg-slate-900 border-b border-slate-700/60 px-6 py-3 flex items-center gap-6 sticky top-0 z-30">
        <span className="font-bold text-base text-white tracking-tight select-none">
          ⚽ Fantasy
        </span>
        <GameSwitcher />
        <nav className="flex gap-1 ml-2">
          {NAV_LINKS.map(({ to, label }) => {
            const href = `/${game}/${to}`;
            const active = pathname === href;
            return (
              <Link
                key={to}
                to={href}
                className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                  active
                    ? 'text-white bg-slate-700/70'
                    : 'text-slate-400 hover:text-white hover:bg-slate-800'
                }`}
              >
                {label}
              </Link>
            );
          })}
        </nav>
      </header>

      <GWContextBar game={game} />

      <main className="p-6">{children}</main>
    </div>
  );
}
