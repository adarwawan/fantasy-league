/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
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
