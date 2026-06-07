import { flagEmoji } from '../utils/flagEmoji'
import styles from './GroupStandings.module.css'

const GROUP_ORDER = ['A','B','C','D','E','F','G','H','I','J','K','L']

const EMPTY_ROW = { mp: 0, w: 0, d: 0, l: 0, gf: 0, ga: 0, gd: 0, pts: 0, advance: false }

export default function GroupStandings({ teams, result }) {
  const groups = buildGroups(teams, result)

  return (
    <div className={styles.grid}>
      {GROUP_ORDER.map(g => (
        <GroupCard key={g} group={g} teams={groups[g] ?? []} />
      ))}
    </div>
  )
}

function GroupCard({ group, teams }) {
  return (
    <div className={styles.card}>
      <div className={styles.groupHeader}>Group {group}</div>
      <table className={styles.table}>
        <thead>
          <tr>
            <th className={styles.thTeam}>Team</th>
            <th>MP</th>
            <th>W</th>
            <th>D</th>
            <th>L</th>
            <th>GF</th>
            <th>GA</th>
            <th>GD</th>
            <th>Pts</th>
          </tr>
        </thead>
        <tbody>
          {teams.map(t => (
            <tr key={t.team_id} className={t.advance ? styles.rowAdvance : ''}>
              <td className={styles.tdTeam}>
                <span className={styles.flag}>{flagEmoji(t.team_id)}</span>
                <span className={styles.abbr}>{t.team_id}</span>
                <span className={styles.name}>{t.name}</span>
              </td>
              <td>{t.mp}</td>
              <td>{t.w}</td>
              <td>{t.d}</td>
              <td>{t.l}</td>
              <td>{t.gf}</td>
              <td>{t.ga}</td>
              <td className={gdClass(t.gd)}>{t.mp > 0 ? formatGD(t.gd) : '—'}</td>
              <td className={styles.pts}>{t.pts}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function buildGroups(teams, result) {
  if (result) {
    const map = {}
    for (const g of result.groups) {
      map[g.group] = g.teams
    }
    return map
  }

  // Initial state: all zeros, sorted alphabetically within each group
  const map = {}
  for (const t of teams) {
    if (!map[t.group]) map[t.group] = []
    map[t.group].push({ team_id: t.id, name: t.name, ...EMPTY_ROW })
  }
  for (const g of Object.keys(map)) {
    map[g].sort((a, b) => a.name.localeCompare(b.name))
  }
  return map
}

function formatGD(gd) {
  if (gd > 0) return `+${gd}`
  return `${gd}`
}

function gdClass(gd) {
  if (gd > 0) return styles.gdPos
  if (gd < 0) return styles.gdNeg
  return ''
}
