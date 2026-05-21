import { useState, useMemo } from 'react'
import { simulateSeason, simulatePlayIn } from '../api'
import { espnLogo } from '../utils/espnLogo'
import CupBracket from './CupBracket'
import PlayInBracket from './PlayInBracket'
import styles from './SeasonStandings.module.css'

function zeroRow(team) {
  return {
    team_id:     team.id,
    team_name:   team.name,
    conference:  team.conference,
    division:    team.division,
    w:           0,
    l:           0,
    conf_record: '0-0',
    div_record:  '0-0',
    streak:      '—',
    last_10:     '0-0',
    home_record: '0-0',
    away_record: '0-0',
    ppg:         0,
    oppg:        0,
    diff:        0,
  }
}

export default function SeasonStandings({ teams }) {
  const [conf, setConf]                   = useState('east')
  const [eastStandings, setEastStandings] = useState(null)
  const [westStandings, setWestStandings] = useState(null)
  const [cup, setCup]                     = useState(null)
  const [playin, setPlayin]               = useState(null)
  const [loading, setLoading]             = useState(false)
  const [playinLoading, setPlayinLoading] = useState(false)
  const [error, setError]                 = useState(null)

  // Default rows derived from /teams (conference/division from API)
  const defaultEast = useMemo(
    () => teams.filter(t => t.conference === 'east').map(zeroRow).sort((a, b) => a.team_id.localeCompare(b.team_id)),
    [teams]
  )
  const defaultWest = useMemo(
    () => teams.filter(t => t.conference === 'west').map(zeroRow).sort((a, b) => a.team_id.localeCompare(b.team_id)),
    [teams]
  )

  const simulated = eastStandings !== null
  const eastRows  = eastStandings ?? defaultEast
  const westRows  = westStandings ?? defaultWest
  const confRows  = conf === 'east' ? eastRows : westRows

  // Seeds 7-10 per conference for play-in
  const eastSeeds = { s7: eastRows[6]?.team_id, s8: eastRows[7]?.team_id, s9: eastRows[8]?.team_id, s10: eastRows[9]?.team_id }
  const westSeeds = { s7: westRows[6]?.team_id, s8: westRows[7]?.team_id, s9: westRows[8]?.team_id, s10: westRows[9]?.team_id }

  async function handleSimulate() {
    setError(null)
    setLoading(true)
    setPlayin(null)
    try {
      const data = await simulateSeason()
      setEastStandings(data.east)
      setWestStandings(data.west)
      setCup(data.cup)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  async function handleSimulatePlayIn() {
    setError(null)
    setPlayinLoading(true)
    try {
      const data = await simulatePlayIn(
        { seed7: eastSeeds.s7, seed8: eastSeeds.s8, seed9: eastSeeds.s9, seed10: eastSeeds.s10 },
        { seed7: westSeeds.s7, seed8: westSeeds.s8, seed9: westSeeds.s9, seed10: westSeeds.s10 },
      )
      setPlayin(data)
    } catch (err) {
      setError(err.message)
    } finally {
      setPlayinLoading(false)
    }
  }

  return (
    <div className={styles.wrap}>

      {error && <div className={styles.errorBox}>{error}</div>}

      {/* Toolbar: conference tabs left, simulate button right */}
      <div className={styles.toolbar}>
        <div className={styles.confTabs}>
          <button
            className={`${styles.confTab} ${conf === 'east' ? styles.confTabActive : ''}`}
            onClick={() => setConf('east')}
          >
            East
          </button>
          <button
            className={`${styles.confTab} ${conf === 'west' ? styles.confTabActive : ''}`}
            onClick={() => setConf('west')}
          >
            West
          </button>
          <button
            className={`${styles.confTab} ${conf === 'cup' ? styles.confTabActive : ''}`}
            onClick={() => setConf('cup')}
          >
            NBA Cup
          </button>
          <button
            className={`${styles.confTab} ${conf === 'playin' ? styles.confTabActive : ''}`}
            onClick={() => setConf('playin')}
          >
            Play-In
          </button>
        </div>

        <button
          className={styles.btnSim}
          onClick={handleSimulate}
          disabled={loading || teams.length === 0}
        >
          {loading ? <span className={styles.spinner} /> : <CalendarIcon />}
          <span>{loading ? 'Simulating…' : 'Simulate Season'}</span>
        </button>
      </div>

      {/* NBA Cup bracket */}
      {conf === 'cup' && <CupBracket cup={cup} />}

      {/* Play-In bracket */}
      {conf === 'playin' && (
        <PlayInBracket
          eastSeeds={eastSeeds}
          westSeeds={westSeeds}
          playin={playin}
          onSimulate={handleSimulatePlayIn}
          loading={playinLoading}
          seasonSimulated={simulated}
        />
      )}

      {/* Standings table */}
      {conf !== 'cup' && conf !== 'playin' && (
        <div className={styles.tableWrap}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th className={styles.thRank}>#</th>
                <th className={styles.thTeam}>Team</th>
                <th className={styles.thNum}>W</th>
                <th className={styles.thNum}>L</th>
                <th className={styles.thNum}>CONF</th>
                <th className={styles.thNum}>DIV</th>
                <th className={styles.thNum}>Streak</th>
                <th className={styles.thNum}>L10</th>
                <th className={styles.thNum}>Home</th>
                <th className={styles.thNum}>Away</th>
                <th className={styles.thNum}>PPG</th>
                <th className={styles.thNum}>OPPG</th>
                <th className={styles.thNum}>DIFF</th>
              </tr>
            </thead>
            <tbody>
              {confRows.map((row, i) => (
                <StandingsRow key={row.team_id} row={row} rank={i + 1} simulated={simulated} />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

function StandingsRow({ row, rank, simulated }) {
  const streakWin  = simulated && row.streak.startsWith('W')
  const streakLoss = simulated && row.streak.startsWith('L')
  const diffPos    = simulated && row.diff > 0
  const diffNeg    = simulated && row.diff < 0

  return (
    <tr className={styles.row}>
      <td className={styles.tdRank}>{rank}</td>

      <td className={styles.tdTeam}>
        <div className={styles.teamCell}>
          <img
            src={espnLogo(row.team_id)}
            alt={row.team_id}
            className={styles.teamLogo}
            onError={e => { e.currentTarget.style.opacity = '0' }}
          />
          <span className={styles.teamAbbr}>{row.team_id}</span>
          <span className={styles.teamName}>{row.team_name}</span>
        </div>
      </td>

      <td className={styles.tdNum}>{row.w}</td>
      <td className={styles.tdNum}>{row.l}</td>

      <td className={styles.tdNum}>{simulated ? row.conf_record : '—'}</td>
      <td className={styles.tdNum}>{simulated ? row.div_record : '—'}</td>

      <td className={`${styles.tdNum} ${streakWin ? styles.win : ''} ${streakLoss ? styles.loss : ''}`}>
        {row.streak}
      </td>

      <td className={styles.tdNum}>{row.last_10}</td>
      <td className={styles.tdNum}>{row.home_record}</td>
      <td className={styles.tdNum}>{row.away_record}</td>

      <td className={styles.tdNum}>{simulated ? row.ppg.toFixed(1) : '0.0'}</td>
      <td className={styles.tdNum}>{simulated ? row.oppg.toFixed(1) : '0.0'}</td>

      <td className={`${styles.tdNum} ${diffPos ? styles.win : ''} ${diffNeg ? styles.loss : ''}`}>
        {simulated ? (row.diff > 0 ? '+' : '') + row.diff.toFixed(1) : '0.0'}
      </td>
    </tr>
  )
}

function CalendarIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" width="15" height="15">
      <path d="M19 3h-1V1h-2v2H8V1H6v2H5C3.9 3 3 3.9 3 5v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H5V8h14v11z" />
    </svg>
  )
}
