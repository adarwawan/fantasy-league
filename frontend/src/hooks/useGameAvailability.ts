import type { GameID } from './useGame';

export function useGameAvailability(): Record<GameID, boolean> {
  return {
    fpl: true,
    wcf: false,
    uclf: false,
  };
}
