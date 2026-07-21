import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AppShell } from './components/layout/AppShell';
import { PlayersPage } from './pages/PlayersPage';
import { TeamsPage }   from './pages/TeamsPage';
import { ScatterPage } from './pages/ScatterPage';
import { StatsPage }   from './pages/StatsPage';
import { PlannerPage } from './pages/PlannerPage';

const queryClient = new QueryClient();

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Navigate to="/fpl/players" replace />} />
          <Route path="/:game" element={<Navigate to="players" replace />} />
          <Route path="/:game/players" element={<AppShell><PlayersPage /></AppShell>} />
          <Route path="/:game/teams"   element={<AppShell><TeamsPage /></AppShell>} />
          <Route path="/:game/scatter" element={<AppShell><ScatterPage /></AppShell>} />
          <Route path="/:game/stats"   element={<AppShell><StatsPage /></AppShell>} />
          <Route path="/:game/planner" element={<AppShell><PlannerPage /></AppShell>} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
