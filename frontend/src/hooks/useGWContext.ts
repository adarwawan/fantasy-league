import { useQuery } from '@tanstack/react-query';
import { fetchDeadline } from '../api/fixtures';

export function useGWContext(game: string) {
  const { data, isLoading } = useQuery({
    queryKey: ['deadline', game],
    queryFn: () => fetchDeadline(game),
    retry: false,
  });

  const deadline = data?.next_deadline ? new Date(data.next_deadline) : null;
  const isPast = deadline ? deadline.getTime() <= Date.now() : false;

  return {
    gw: data?.current_gw ?? null,
    deadline: isPast ? null : deadline,
    cachedAt: data?.cached_at ?? null,
    isLoading,
  };
}
