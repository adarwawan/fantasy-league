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
    <div className="flex gap-1 rounded-lg bg-slate-800 p-1">
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
                ? 'bg-slate-600 shadow text-white'
                : enabled
                ? 'text-slate-400 hover:text-white hover:bg-slate-700'
                : 'text-slate-600 cursor-not-allowed'
            }`}
          >
            {label}
          </button>
        );
      })}
    </div>
  );
}
