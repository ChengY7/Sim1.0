export async function fetchSeasons() {
  const res = await fetch('/nba/seasons')
  if (!res.ok) throw new Error(`/nba/seasons ${res.status}`)
  return res.json() // { seasons, default_season }
}

export async function fetchTeams() {
  const res = await fetch('/nba/teams')
  if (!res.ok) throw new Error(`/nba/teams ${res.status}`)
  const { teams } = await res.json()
  return teams
}

export async function simulate(homeTeamId, awayTeamId, season) {
  const res = await fetch('/nba/simulate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ home_team_id: homeTeamId, away_team_id: awayTeamId, season }),
  })
  const body = await res.json()
  if (!res.ok) throw new Error(body.error ?? `Server error ${res.status}`)
  return body
}

export async function simulateSeason(season) {
  const res = await fetch('/nba/simulate/season', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ season }),
  })
  const body = await res.json()
  if (!res.ok) throw new Error(body.error ?? `Server error ${res.status}`)
  return body
}

export async function simulateDraftLottery(teams) {
  const res = await fetch('/nba/simulate/draft-lottery', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ teams }),
  })
  const body = await res.json()
  if (!res.ok) throw new Error(body.error ?? `Server error ${res.status}`)
  return body
}

export async function simulatePlayIn(east, west, season) {
  const res = await fetch('/nba/simulate/playin', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ east, west, season }),
  })
  const body = await res.json()
  if (!res.ok) throw new Error(body.error ?? `Server error ${res.status}`)
  return body
}
