import { Link, useParams } from 'react-router-dom';
import { GameSwitcher } from './GameSwitcher';

export function AppShell({ children }: { children: React.ReactNode }) {
  const { game = 'fpl' } = useParams<{ game: string }>();
  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-6 py-3 flex items-center gap-6">
        <span className="font-bold text-lg text-gray-900">Fantasy Dashboard</span>
        <GameSwitcher />
        <nav className="flex gap-4 ml-4">
          <Link to={`/${game}/players`} className="text-sm text-gray-600 hover:text-gray-900">Players</Link>
          <Link to={`/${game}/teams`}   className="text-sm text-gray-600 hover:text-gray-900">Teams</Link>
          <Link to={`/${game}/scatter`} className="text-sm text-gray-600 hover:text-gray-900">Scatter</Link>
        </nav>
      </header>
      <main className="p-6">{children}</main>
    </div>
  );
}
