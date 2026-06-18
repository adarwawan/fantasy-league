import type { Player } from '../../types/player';

interface Props {
  player:  Player;
  avgFdr:  number;
}

export function PlayerTooltip({ player, avgFdr }: Props) {
  return (
    <div className="bg-white border border-gray-200 rounded-lg shadow-lg p-3 text-sm min-w-[180px]">
      <div className="font-semibold text-gray-900 mb-1">{player.name}</div>
      <div className="text-gray-500 text-xs mb-2">
        {player.team.short_name} · {player.position} · £{player.price.toFixed(1)}m
      </div>
      <div className="space-y-0.5 text-xs text-gray-700">
        <div className="flex justify-between gap-4">
          <span>Form</span>
          <span className="font-medium">{player.form.toFixed(1)}</span>
        </div>
        <div className="flex justify-between gap-4">
          <span>Global own</span>
          <span className="font-medium">{player.global_ownership.toFixed(1)}%</span>
        </div>
        <div className="flex justify-between gap-4">
          <span>Top-N own</span>
          <span className="font-medium">{player.top_n_ownership.toFixed(1)}%</span>
        </div>
        <div className="flex justify-between gap-4">
          <span>Avg FDR</span>
          <span className="font-medium">{avgFdr.toFixed(2)}</span>
        </div>
      </div>
    </div>
  );
}
