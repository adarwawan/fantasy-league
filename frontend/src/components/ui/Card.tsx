type Variant = 'default' | 'panel' | 'nested';

interface Props extends React.HTMLAttributes<HTMLDivElement> {
  /** default: list/grid item · panel: large page section · nested: block inside a card */
  variant?: Variant;
}

const VARIANT: Record<Variant, string> = {
  default: 'rounded-xl border border-line bg-surface-raised/60 p-3 sm:p-4',
  panel:   'rounded-2xl border border-line bg-surface p-5 sm:p-6',
  nested:  'rounded-lg border border-line bg-surface/60 p-3',
};

export function Card({ variant = 'default', className = '', ...rest }: Props) {
  return <div className={`${VARIANT[variant]} ${className}`} {...rest} />;
}
