import { Link, useParams, useLocation } from 'react-router-dom';
import { GameSwitcher } from './GameSwitcher';
import { GWContextBar } from './GWContextBar';

const NAV_LINKS = [
  { to: 'players', label: 'Players' },
  { to: 'teams',   label: 'Teams'   },
  { to: 'scatter', label: 'Scatter' },
  // Stats is FPL-specific (built around the FPL scoring rules).
  { to: 'stats',   label: 'Stats', games: ['fpl'] },
];

export function AppShell({ children }: { children: React.ReactNode }) {
  const { game = 'fpl' } = useParams<{ game: string }>();
  const { pathname } = useLocation();

  const navLinks = NAV_LINKS
    .filter(({ games }) => !games || games.includes(game))
    .map(({ to, label }) => {
      const href = `/${game}/${to}`;
      const active = pathname === href;
      return { to, label, href, active };
    });

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100">
      <header className="bg-slate-900 border-b border-slate-700/60 sticky top-0 z-30">
        {/* Top bar: logo + game switcher (+ nav on desktop) */}
        <div className="px-4 py-3 flex items-center gap-4">
          <span className="font-bold text-base text-white tracking-tight select-none">
            ⚽ Fantasy
          </span>
          <GameSwitcher />
          {/* Desktop nav — hidden on mobile */}
          <nav className="hidden md:flex gap-1 ml-2">
            {navLinks.map(({ to, label, href, active }) => (
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
            ))}
          </nav>
        </div>

        {/* Mobile tab strip — hidden on desktop */}
        <nav className="md:hidden flex border-t border-slate-700/60">
          {navLinks.map(({ to, label, href, active }) => (
            <Link
              key={to}
              to={href}
              className={`flex-1 text-center py-2.5 text-sm font-medium transition-colors border-b-2 ${
                active
                  ? 'text-white border-violet-500'
                  : 'text-slate-400 border-transparent hover:text-white hover:border-slate-500'
              }`}
            >
              {label}
            </Link>
          ))}
        </nav>
      </header>

      <GWContextBar game={game} />

      <main className="px-4 py-6 md:p-6">{children}</main>
    </div>
  );
}
