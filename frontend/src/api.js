export async function fetchSeasons() {
  const res = await fetch('/seasons')
  if (!res.ok) throw new Error(`/seasons ${res.status}`)
  return res.json() // { seasons, default_season }
}

export async function fetchTeams() {
  const res = await fetch('/teams')
  if (!res.ok) throw new Error(`/teams ${res.status}`)
  const { teams } = await res.json()
  return teams
}

export async function simulate(homeTeamId, awayTeamId, season) {
  const res = await fetch('/simulate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ home_team_id: homeTeamId, away_team_id: awayTeamId, season }),
  })
  const body = await res.json()
  if (!res.ok) throw new Error(body.error ?? `Server error ${res.status}`)
  return body
}

export async function simulateSeason(season) {
  const res = await fetch('/simulate/season', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ season }),
  })
  const body = await res.json()
  if (!res.ok) throw new Error(body.error ?? `Server error ${res.status}`)
  return body
}

export async function simulateDraftLottery(teams) {
  const res = await fetch('/simulate/draft-lottery', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ teams }),
  })
  const body = await res.json()
  if (!res.ok) throw new Error(body.error ?? `Server error ${res.status}`)
  return body
}

export async function simulatePlayIn(east, west, season) {
  const res = await fetch('/simulate/playin', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ east, west, season }),
  })
  const body = await res.json()
  if (!res.ok) throw new Error(body.error ?? `Server error ${res.status}`)
  return body
}
