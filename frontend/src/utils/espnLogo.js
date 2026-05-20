const ESPN_SLUG = {
  GSW: 'gs',
  NOP: 'no',
  NYK: 'ny',
  SAS: 'sa',
  UTA: 'utah',
  WAS: 'wsh',
}

export function espnLogo(teamId) {
  const slug = ESPN_SLUG[teamId] ?? teamId.toLowerCase()
  return `https://a.espncdn.com/i/teamlogos/nba/500/${slug}.png`
}
