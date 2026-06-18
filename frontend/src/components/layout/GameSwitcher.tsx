import { useGame, type GameID } from '../../hooks/useGame';

const GAMES: { id: GameID; label: string }[] = [
  { id: 'fpl',  label: 'FPL'  },
  { id: 'wcf',  label: 'WCF'  },
  { id: 'uclf', label: 'UCLF' },
];

export function GameSwitcher() {
  const { game, setGame } = useGame();
  return (
    <div className="flex gap-1 rounded-lg bg-gray-100 p-1">
      {GAMES.map(({ id, label }) => (
        <button
          key={id}
          onClick={() => setGame(id)}
          className={`px-3 py-1 rounded-md text-sm font-medium transition-colors ${
            game === id
              ? 'bg-white shadow text-gray-900'
              : 'text-gray-500 hover:text-gray-700'
          }`}
        >
          {label}
        </button>
      ))}
    </div>
  );
}
