import { useGame, type GameID } from '../../hooks/useGame';
import { useGameAvailability } from '../../hooks/useGameAvailability';

const GAMES: { id: GameID; label: string }[] = [
  { id: 'fpl',  label: 'FPL'  },
  { id: 'wcf',  label: 'WCF'  },
  { id: 'uclf', label: 'UCLF' },
];

export function GameSwitcher() {
  const { game, setGame } = useGame();
  const availability = useGameAvailability();

  return (
    <div className="flex gap-1 rounded-lg bg-gray-100 p-1">
      {GAMES.map(({ id, label }) => {
        const enabled = availability[id];
        return (
          <button
            key={id}
            onClick={() => enabled && setGame(id)}
            disabled={!enabled}
            title={enabled ? undefined : `${label} — no data available`}
            className={`px-3 py-1 rounded-md text-sm font-medium transition-colors ${
              game === id
                ? 'bg-white shadow text-gray-900'
                : enabled
                ? 'text-gray-500 hover:text-gray-700'
                : 'text-gray-300 cursor-not-allowed'
            }`}
          >
            {label}
          </button>
        );
      })}
    </div>
  );
}
