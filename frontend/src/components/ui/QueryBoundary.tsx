import { EmptyState } from './EmptyState';
import { ErrorState } from './ErrorState';

interface Props {
  isLoading: boolean;
  isError: boolean;
  /** True when the query succeeded but there is nothing to show. */
  isEmpty?: boolean;
  onRetry: () => void;
  /** Noun phrase used in the error/empty copy: "players", "shot-zone data". */
  what: string;
  /** Empty-state explanation (no-data case). */
  emptyHint?: string;
  skeleton: React.ReactNode;
  children: React.ReactNode;
}

/**
 * The single loading → error → empty → content switch. Pages keep their
 * PageHeader outside this so the title never disappears while loading.
 */
export function QueryBoundary({ isLoading, isError, isEmpty = false, onRetry, what, emptyHint, skeleton, children }: Props) {
  if (isLoading) return <>{skeleton}</>;
  if (isError) return <ErrorState what={what} onRetry={onRetry} />;
  if (isEmpty) return <EmptyState what={what} hint={emptyHint} />;
  return <>{children}</>;
}
