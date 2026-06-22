export const fdrColours: Record<number, { bg: string; text: string }> = {
  1: { bg: 'bg-green-200',  text: 'text-green-900'  },
  2: { bg: 'bg-teal-100',   text: 'text-teal-900'   },
  3: { bg: 'bg-amber-100',  text: 'text-amber-900'  },
  4: { bg: 'bg-orange-200', text: 'text-orange-900' },
  5: { bg: 'bg-red-200',    text: 'text-red-900'    },
};

export function fdrLabel(difficulty: number): string {
  const labels: Record<number, string> = { 1: 'Very Easy', 2: 'Easy', 3: 'Medium', 4: 'Hard', 5: 'Very Hard' };
  return labels[difficulty] ?? 'Unknown';
}

/** Map opponent ovr_form to chip colours: high form = tough opponent = red */
export function ovrFormColour(ovr: number): { bg: string; text: string } {
  if (ovr >= 2.5) return { bg: 'bg-red-200',    text: 'text-red-900'    };
  if (ovr >= 1.5) return { bg: 'bg-amber-100',  text: 'text-amber-900'  };
  return               { bg: 'bg-green-200',  text: 'text-green-900'  };
}
