export async function fetchFIFATeams() {
  const res = await fetch('/fifa/teams')
  if (!res.ok) throw new Error(`/fifa/teams ${res.status}`)
  const { teams } = await res.json()
  return teams
}

export async function simulateFIFAMatch(team1Id, team2Id, extraTime = false, seed = null) {
  const body = { team1_id: team1Id, team2_id: team2Id, extra_time: extraTime }
  if (seed != null) body.seed = seed
  const res = await fetch('/fifa/simulate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error ?? `Server error ${res.status}`)
  return data
}

export async function simulateGroupStage(seed = null) {
  const body = {}
  if (seed != null) body.seed = seed
  const res = await fetch('/fifa/simulate/group-stage', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error ?? `Server error ${res.status}`)
  return data
}
