// Display metadata for the canonical PL team names the set-piece API emits.
// Kept isolated in the frontend (no dependency on the players/teams pipeline):
// just a short code + primary colour for the card badge.
interface TeamMeta {
  code:  string;
  color: string; // hex primary colour
}

const TEAM_META: Record<string, TeamMeta> = {
  'Arsenal':        { code: 'ARS', color: '#EF0107' },
  'Aston Villa':    { code: 'AVL', color: '#670E36' },
  'Bournemouth':    { code: 'BOU', color: '#DA291C' },
  'Brentford':      { code: 'BRE', color: '#E30613' },
  'Brighton':       { code: 'BHA', color: '#0057B8' },
  'Burnley':        { code: 'BUR', color: '#6C1D45' },
  'Chelsea':        { code: 'CHE', color: '#034694' },
  'Crystal Palace': { code: 'CRY', color: '#1B458F' },
  'Everton':        { code: 'EVE', color: '#003399' },
  'Fulham':         { code: 'FUL', color: '#1B1B1B' },
  'Leeds':          { code: 'LEE', color: '#FFCD00' },
  'Liverpool':      { code: 'LIV', color: '#C8102E' },
  'Man City':       { code: 'MCI', color: '#6CABDD' },
  'Man Utd':        { code: 'MUN', color: '#DA291C' },
  'Newcastle':      { code: 'NEW', color: '#241F20' },
  "Nott'm Forest":  { code: 'NFO', color: '#DD0000' },
  'Spurs':          { code: 'TOT', color: '#132257' },
  'Sunderland':     { code: 'SUN', color: '#EB172B' },
  'West Ham':       { code: 'WHU', color: '#7A263A' },
  'Wolves':         { code: 'WOL', color: '#FDB913' },
};

export function teamMeta(name: string): TeamMeta {
  return TEAM_META[name] ?? { code: name.slice(0, 3).toUpperCase(), color: '#64748b' };
}

/** Pick a readable text colour (black/white) for a solid hex background. */
export function readableText(hex: string): string {
  const h = hex.replace('#', '');
  const r = parseInt(h.slice(0, 2), 16);
  const g = parseInt(h.slice(2, 4), 16);
  const b = parseInt(h.slice(4, 6), 16);
  // Relative luminance (sRGB approximation).
  const lum = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
  return lum > 0.6 ? '#0f172a' : '#ffffff';
}
