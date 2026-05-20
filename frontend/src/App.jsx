import { useState, useEffect } from 'react'
import { fetchTeams, simulate } from './api'
import TeamPicker from './components/TeamPicker'
import Scoreboard from './components/Scoreboard'
import PlayByPlay from './components/PlayByPlay'
import SeasonStandings from './components/SeasonStandings'
import styles from './App.module.css'

export default function App() {
  const [mode, setMode]     = useState('game')   // 'game' | 'season'
  const [teams, setTeams]   = useState([])
  const [homeId, setHomeId] = useState('LAL')
  const [awayId, setAwayId] = useState('BOS')
  const [result, setResult] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError]   = useState(null)

  useEffect(() => {
    fetchTeams()
      .then(setTeams)
      .catch(() => setError('Cannot reach the API at localhost:8080. Start the backend with: cd backend && go run ./cmd/server'))
  }, [])

  async function handleSimulate() {
    if (homeId === awayId) { setError('Please select two different teams.'); return }
    setError(null)
    setResult(null)
    setLoading(true)
    try {
      setResult(await simulate(homeId, awayId))
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={styles.app}>
      {/* ── Navbar ── */}
      <nav className={styles.navbar}>
        <div className={styles.logo}>
          <div className={styles.logoIcon}>
            <BasketballIcon />
          </div>
          <span className={styles.wordmark}>Sim1.0</span>
        </div>

        <div className={styles.modeTabs}>
          <button
            className={`${styles.modeTab} ${mode === 'game' ? styles.modeTabActive : ''}`}
            onClick={() => setMode('game')}
          >
            <DiceIcon />
            Game Sim
          </button>
          <button
            className={`${styles.modeTab} ${mode === 'season' ? styles.modeTabActive : ''}`}
            onClick={() => setMode('season')}
          >
            <CalendarIcon />
            Season Sim
          </button>
        </div>
      </nav>

      {/* ── Game Sim tab ── */}
      {mode === 'game' && (
        <>
          <TeamPicker
            teams={teams}
            homeId={homeId} onHomeChange={setHomeId}
            awayId={awayId} onAwayChange={setAwayId}
          />

          <div className={styles.simWrap}>
            <button className={styles.btnSim} onClick={handleSimulate} disabled={loading || teams.length === 0}>
              {loading ? <span className={styles.spinner} /> : <PlayIcon />}
              <span>{loading ? 'Simulating…' : 'Simulate Game'}</span>
            </button>
          </div>

          {error && <div className={styles.errorBox}>{error}</div>}

          {result && (
            <>
              <Scoreboard state={result.state} seed={result.seed} homeId={homeId} awayId={awayId} />
              <PlayByPlay events={result.events} />
            </>
          )}
        </>
      )}

      {/* ── Season Sim tab ── */}
      {mode === 'season' && (
        <SeasonStandings teams={teams} />
      )}
    </div>
  )
}

function BasketballIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="white" width="24" height="24">
      <path d="M12 2C6.477 2 2 6.477 2 12s4.477 10 10 10 10-4.477 10-10S17.523 2 12 2zm0 2c.74 0 1.46.1 2.15.28A9.97 9.97 0 0 0 12.07 9H12a9.97 9.97 0 0 0-2.15-4.72A7.98 7.98 0 0 1 12 4zm-3.5.93A11.97 11.97 0 0 1 10.06 9H4.26A8.02 8.02 0 0 1 8.5 4.93zM4 12c0-.34.02-.67.06-1h6.01c-.04.33-.07.66-.07 1s.03.67.07 1H4.06A8.1 8.1 0 0 1 4 12zm.26 3h5.8a11.97 11.97 0 0 1-1.56 4.07A8.02 8.02 0 0 1 4.26 15zm7.81 5c-.74 0-1.46-.1-2.15-.28A9.97 9.97 0 0 0 11.93 15H12a9.97 9.97 0 0 0 2.15 4.72A7.98 7.98 0 0 1 12 20zm3.5-.93A11.97 11.97 0 0 1 13.94 15h5.8a8.02 8.02 0 0 1-4.24 4.07zM20 12c0 .34-.02.67-.06 1h-6.01c.04-.33.07-.66.07-1s-.03-.67-.07-1h6.01c.04.33.06.66.06 1zm-.26-3h-5.8a11.97 11.97 0 0 1 1.56-4.07A8.02 8.02 0 0 1 19.74 9z" />
    </svg>
  )
}

function PlayIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="white" width="16" height="16">
      <path d="M8 5v14l11-7z" />
    </svg>
  )
}

function DiceIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" width="14" height="14">
      <path d="M5 3h14a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2zm2.5 4a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3zm9 0a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3zm-4.5 4a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3zm-4.5 4a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3zm9 0a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3z" />
    </svg>
  )
}

function CalendarIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" width="14" height="14">
      <path d="M19 3h-1V1h-2v2H8V1H6v2H5C3.9 3 3 3.9 3 5v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H5V8h14v11z" />
    </svg>
  )
}
