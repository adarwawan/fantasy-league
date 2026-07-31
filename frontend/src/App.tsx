import { BrowserRouter, Routes, Route, Navigate, useParams } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AppShell } from './components/layout/AppShell';
import { PlayersPage } from './pages/PlayersPage';
import { TeamsPage }   from './pages/TeamsPage';
import { StatsPage }   from './pages/StatsPage';
import { PlannerPage } from './pages/PlannerPage';
import { SetPiecesPage } from './pages/SetPiecesPage';

const queryClient = new QueryClient();

// Scatter was merged into Players as its "Plot" view — redirect old links.
function ScatterRedirect() {
  const { game = 'fpl' } = useParams<{ game: string }>();
  return <Navigate to={`/${game}/players?view=plot`} replace />;
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Navigate to="/fpl/players" replace />} />
          <Route path="/:game" element={<Navigate to="players" replace />} />
          <Route path="/:game/players" element={<AppShell><PlayersPage /></AppShell>} />
          <Route path="/:game/teams"   element={<AppShell><TeamsPage /></AppShell>} />
          {/* Scatter merged into Players as the "Plot" view — redirect old links. */}
          <Route path="/:game/scatter" element={<ScatterRedirect />} />
          <Route path="/:game/stats"   element={<AppShell><StatsPage /></AppShell>} />
          <Route path="/:game/planner" element={<AppShell><PlannerPage /></AppShell>} />
          <Route path="/:game/set-pieces" element={<AppShell><SetPiecesPage /></AppShell>} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
