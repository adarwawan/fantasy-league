/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        // Semantic tokens — use these in components instead of raw slate/indigo
        // so surfaces, borders and the accent stay consistent app-wide.
        surface: {
          DEFAULT: '#0f172a', // slate-900 — cards, panels, inputs
          raised:  '#1e293b', // slate-800 — nested blocks, hover fills
        },
        line: 'rgb(51 65 85 / 0.6)', // slate-700 @ 60% — every border/divider
        accent: {
          DEFAULT: '#4f46e5', // indigo-600 — selected, primary actions
          hover:   '#6366f1', // indigo-500
          ring:    '#6366f1', // focus ring
        },
        pos: {
          gk:  '#34d399', // emerald-400
          def: '#60a5fa', // blue-400
          mid: '#c084fc', // purple-400
          fwd: '#f87171', // red-400
        },
      },
    },
  },
  plugins: [],
}
