# Frontend (React + TypeScript)

Vite 6, React 18, TypeScript, Tailwind 3, TanStack Query/Table/Virtual, Recharts, react-router v6. No test runner or formatter is configured; the gate is `tsc`.

## Layout (`src/`)

- `api/` — one file per backend area; all calls go through `apiFetch` in `api/client.ts`. Isolated modules (set-pieces, shot-zones) use game-less routes (`/api/shot-zones/...`); per-game routes take the `game` param.
- `hooks/` — TanStack Query wrappers (`useX`), one per api file. Query keys `['area', 'thing', ...args]`; set `staleTime` deliberately (30 min for Understat-derived data).
- `types/` — API response types mirroring backend JSON.
- `utils/` — pure derivations (`minutes.ts`, `fdr.ts`, `ownership.ts`, `price.ts`). Derived labels/chips live here, not on the backend.
- `components/ui/` — shared primitives (`Button`, `Card`, `FilterBar`, `SegmentedControl`, `Select`, `QueryBoundary`, `EmptyState`, `ErrorState`, `Skeletons`...). **Reuse these before writing new markup.**
- `components/layout/` — `AppShell`, `PageHeader`, `GWContextBar`, `GameSwitcher`.
- `components/<area>/` and `pages/<Area>Page.tsx` — feature code. Routes are declared in `App.tsx` under `/:game/<page>` wrapped in `AppShell`.

## Conventions

- New page: `pages/XPage.tsx` + route in `App.tsx` + nav entry in `AppShell` + `api/x.ts` + `hooks/useX.ts` + `types/x.ts`. FPL-only pages are gated by game.
- Wrap data views in `QueryBoundary`; provide loading (skeleton), empty and error states.
- Mobile-first: Players/Teams/Shot Zones were all reworked for mobile. Check a ~375px viewport; wide tables scroll horizontally instead of overflowing the page.
- Show rolling-window numbers and say which window (e.g. "last 5 GW"), per the form-not-season invariant.
- Fixture chips share one data shape across player and team views (`FixtureChip`); keep them consistent.
- API base URL from `VITE_API_URL`.
- `dist/` and `node_modules/` are gitignored; CI deploys with `npm ci && npm run build`.

## Verify

```bash
cd frontend && npx tsc --noEmit        # or: make check-frontend (repo root)
npm run dev                            # :5173, needs the backend on :8080 for real data
```

Type-checks don't prove a UI change works: load the page in the dev server (launch config `frontend`) and exercise it, including the mobile viewport.
