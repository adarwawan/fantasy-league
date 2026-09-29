import { Button } from './Button';

interface Props {
  /** What failed to load, as a plain noun phrase: "players", "set-piece data". */
  what: string;
  onRetry?: () => void;
}

/** Full-area load failure. One wording everywhere: "Couldn't load {what}." */
export function ErrorState({ what, onRetry }: Props) {
  return (
    <div role="alert" className="flex flex-col items-center gap-3 py-16 text-center">
      <span className="text-3xl" aria-hidden="true">⚠</span>
      <div>
        <p className="text-sm text-slate-300">Couldn’t load {what}.</p>
        <p className="mt-1 text-xs text-slate-500">Check your connection and try again.</p>
      </div>
      {onRetry && <Button onClick={onRetry}>Try again</Button>}
    </div>
  );
}
