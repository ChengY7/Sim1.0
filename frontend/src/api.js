export async function fetchTeams() {
  const res = await fetch('/teams')
  if (!res.ok) throw new Error(`/teams ${res.status}`)
  const { teams } = await res.json()
  return teams
}

export async function simulate(homeTeamId, awayTeamId) {
  const res = await fetch('/simulate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ home_team_id: homeTeamId, away_team_id: awayTeamId }),
  })
  const body = await res.json()
  if (!res.ok) throw new Error(body.error ?? `Server error ${res.status}`)
  return body
}

export async function simulateSeason() {
  const res = await fetch('/simulate/season', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({}),
  })
  const body = await res.json()
  if (!res.ok) throw new Error(body.error ?? `Server error ${res.status}`)
  return body
}
