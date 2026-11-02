---
name: add-page
description: Add or substantially change a frontend page in frontend/src - route, nav entry, api/hooks/types, shared UI primitives, loading/empty/error states, mobile check. Use for any new screen or major view.
---

# Add a frontend page

Newest reference: `pages/ShotZonePage.tsx` with `api/shotzone.ts`, `hooks/useShotZone.ts`, `types/shotzone.ts`, `components/shotzone/`. Conventions are in `frontend/CLAUDE.md`.

## Steps

1. **Types** `src/types/<name>.ts` mirroring the backend JSON exactly (snake_case field names as returned).
2. **API** `src/api/<name>.ts`: thin functions over `apiFetch` from `api/client.ts`. Per-game data takes the `game` param (`/api/${game}/...`); isolated modules use game-less routes (`/api/shot-zones/...`). Encode path params with `encodeURIComponent`.
3. **Hook** `src/hooks/use<Name>.ts`: TanStack Query, key `['area', 'thing', ...args]`, `enabled` when args are required, deliberate `staleTime` (30 min for Understat-derived data).
4. **Page** `src/pages/<Name>Page.tsx`:
   - `PageHeader`, `usePageTitle`, then the body in `QueryBoundary` with a skeleton from `components/ui/Skeletons` and `ErrorState`/`EmptyState`.
   - Compose from `components/ui` (`Card`, `FilterBar`, `Field`, `Select`, `SearchInput`, `SegmentedControl`, `TeamBadge`...). Add a new primitive only if two pages need it.
   - Feature components go in `components/<area>/`. Derived labels/chips go in `utils/`, not the backend.
   - State the window on screen ("last 5 GW", "last 6 matches"). No season-total views unless asked.
5. **Route** in `src/App.tsx`: `<Route path="/:game/<slug>" element={<AppShell><NamePage /></AppShell>} />`.
6. **Nav** in `NAV_LINKS` in `components/layout/AppShell.tsx`. Add `games: ['fpl']` for FPL-only pages; keep the workflow ordering (research, read-outs, act).
7. **Verify**:
   - `make check-frontend` (tsc).
   - Start the `frontend` launch config (`.claude/launch.json`) and the backend, then use the page for real: every state (loading, empty, error, populated), filters, and a ~375px viewport where wide tables scroll inside their container instead of widening the page.
   - Check the console for errors. Type-checking alone does not prove the UI works.

## Don't

- Don't fetch inside components; go through hooks.
- Don't add formatters, test frameworks, or dependencies as part of a page change.
