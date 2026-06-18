import { useParams, useNavigate } from 'react-router-dom';

export type GameID = 'fpl' | 'wcf' | 'uclf';

export function useGame() {
  const { game } = useParams<{ game: GameID }>();
  const navigate  = useNavigate();
  const gameID    = (game ?? 'fpl') as GameID;

  function setGame(next: GameID) {
    navigate(`/${next}/players`);
  }

  return { game: gameID, setGame };
}
