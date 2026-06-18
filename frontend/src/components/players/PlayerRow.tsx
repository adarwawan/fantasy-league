import type { Row } from '@tanstack/react-table';
import type { Player } from '../../types/player';
import { FixtureChip } from './FixtureChip';

const STATUS_DOT: Record<Player['status'], string> = {
  available: 'bg-green-500',
  doubt:     'bg-amber-400',
  injured:   'bg-red-500',
};

export function PlayerRow({ row }: { row: Row<Player> }) {
  const player = row.original;
  const isInjured = player.status === 'injured';

  return (
    <tr className={`border-b border-gray-100 hover:bg-gray-50 ${isInjured ? 'opacity-40' : ''}`}>
      {/* Name + status */}
      <td className="px-3 py-2 text-sm font-medium text-gray-900 whitespace-nowrap">
        <div className="flex items-center gap-1.5">
          <span
            className={`shrink-0 w-2 h-2 rounded-full ${STATUS_DOT[player.status]}`}
            title={player.news || player.status}
          />
          <span
            className="cursor-default"
            title={player.news || undefined}
          >
            {player.name}
          </span>
        </div>
      </td>

      {/* Team */}
      <td className="px-3 py-2 text-sm text-gray-600">{player.team.short_name}</td>

      {/* Position */}
      <td className="px-3 py-2 text-sm text-gray-600">{player.position}</td>

      {/* Price */}
      <td className="px-3 py-2 text-sm text-gray-600 text-right">
        £{player.price.toFixed(1)}m
      </td>

      {/* Form */}
      <td className="px-3 py-2 text-sm text-gray-600 text-right">
        {player.form.toFixed(1)}
      </td>

      {/* Global ownership */}
      <td className="px-3 py-2 text-sm text-gray-600 text-right">
        {player.global_ownership.toFixed(1)}%
      </td>

      {/* Top-N ownership */}
      <td className="px-3 py-2 text-sm text-gray-600 text-right">
        {player.top_n_ownership.toFixed(1)}%
      </td>

      {/* Next 5 fixture chips */}
      <td className="px-3 py-2">
        <div className="flex gap-1 flex-wrap">
          {player.fixtures.slice(0, 5).map((f, i) => (
            <FixtureChip key={i} fixture={f} />
          ))}
        </div>
      </td>
    </tr>
  );
}
