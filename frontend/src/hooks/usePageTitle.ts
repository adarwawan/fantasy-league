import { useEffect } from 'react';
import { useParams } from 'react-router-dom';

/** Sets document.title to "<Page> · <GAME>" — the one title format for every page. */
export function usePageTitle(page: string) {
  const { game = 'fpl' } = useParams<{ game: string }>();
  useEffect(() => {
    document.title = `${page} · ${game.toUpperCase()}`;
  }, [page, game]);
}
