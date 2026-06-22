interface Props {
  message?: string;
  onRetry?: () => void;
}

export function ErrorState({ message = 'Something went wrong.', onRetry }: Props) {
  return (
    <div
      role="alert"
      className="flex flex-col items-center gap-3 py-16 text-center"
    >
      <span className="text-3xl" aria-hidden="true">⚠</span>
      <p className="text-slate-300 text-sm">{message}</p>
      {onRetry && (
        <button
          onClick={onRetry}
          className="mt-1 px-4 py-2 rounded-md bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-sm font-medium text-white transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-400"
        >
          Try again
        </button>
      )}
    </div>
  );
}
