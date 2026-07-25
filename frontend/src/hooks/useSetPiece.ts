import { useQuery } from '@tanstack/react-query';
import { fetchSetPieceTeams } from '../api/setpiece';

export function useSetPieceTeams() {
  return useQuery({
    queryKey:  ['setpiece', 'teams'],
    queryFn:   fetchSetPieceTeams,
    staleTime: 30 * 60 * 1000,
  });
}
