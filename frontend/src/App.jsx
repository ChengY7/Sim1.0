import { useState, useEffect } from 'react'
import { fetchSeasons, fetchTeams, simulate } from './nba/api'
import TeamPicker from './nba/components/TeamPicker'
import Scoreboard from './nba/components/Scoreboard'
import PlayByPlay from './nba/components/PlayByPlay'
import SeasonStandings from './nba/components/SeasonStandings'
import styles from './App.module.css'

export default function App() {
  const [sport, setSport]     = useState('nba')  // 'nba' | 'fifa'
  const [mode, setMode]       = useState('game')  // 'game' | 'season'
  const [teams, setTeams]     = useState([])
  const [seasons, setSeasons] = useState([])
  const [season, setSeason]   = useState(null)
  const [homeId, setHomeId]   = useState('LAL')
  const [awayId, setAwayId]   = useState('BOS')
  const [result, setResult]   = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError]     = useState(null)

  useEffect(() => {
    if (sport !== 'nba') return
    fetchTeams()
      .then(setTeams)
      .catch(() => setError('Cannot reach the API at localhost:8080. Start the backend with: cd backend && go run ./cmd/server'))

    fetchSeasons()
      .then(({ seasons, default_season }) => {
        setSeasons(seasons)
        setSeason(default_season)
      })
      .catch(() => {})
  }, [sport])

  function handleSportChange(newSport) {
    if (newSport === sport) return
    setSport(newSport)
    setResult(null)
    setError(null)
  }

  function handleSeasonChange(newSeason) {
    if (newSeason === season) return
    setSeason(newSeason)
    setResult(null)
  }

  async function handleSimulate() {
    if (homeId === awayId) { setError('Please select two different teams.'); return }
    setError(null)
    setResult(null)
    setLoading(true)
    try {
      setResult(await simulate(homeId, awayId, season))
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

        <div className={styles.sportTabs}>
          <button
            className={`${styles.sportTab} ${sport === 'nba' ? styles.sportTabActive : ''}`}
            onClick={() => handleSportChange('nba')}
          >
            <BasketballSmallIcon />
            NBA
          </button>
          <button
            className={`${styles.sportTab} ${sport === 'fifa' ? styles.sportTabActive : ''}`}
            onClick={() => handleSportChange('fifa')}
          >
            <SoccerIcon />
            World Cup
          </button>
        </div>

        {sport === 'nba' && (
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
        )}
      </nav>

      {/* ── NBA ── */}
      {sport === 'nba' && (
        <>
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

          {mode === 'season' && (
            <div className={styles.seasonContent}>
              <div className={styles.seasonBar}>
                <span className={styles.seasonLabel}>Season</span>
                <div className={styles.seasonTabs}>
                  {seasons.length > 0
                    ? seasons.map(s => (
                        <button
                          key={s.season}
                          className={`${styles.seasonTab} ${season === s.season ? styles.seasonTabActive : ''}`}
                          onClick={() => handleSeasonChange(s.season)}
                        >
                          {s.season}
                        </button>
                      ))
                    : season
                      ? <button className={`${styles.seasonTab} ${styles.seasonTabActive}`}>{season}</button>
                      : <span className={styles.seasonLoading}>Loading…</span>
                  }
                </div>
              </div>
              <SeasonStandings
                key={season ?? 'default'}
                teams={teams}
                season={season}
                actualStandings={seasons.find(s => s.season === season)?.standings ?? []}
              />
            </div>
          )}
        </>
      )}

      {/* ── FIFA / World Cup ── */}
      {sport === 'fifa' && (
        <div className={styles.comingSoon}>
          <SoccerIcon size={48} />
          <h2>World Cup 2026</h2>
          <p>Group stage simulation coming soon.</p>
        </div>
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

function BasketballSmallIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" width="14" height="14">
      <path d="M12 2C6.477 2 2 6.477 2 12s4.477 10 10 10 10-4.477 10-10S17.523 2 12 2zm0 2c.74 0 1.46.1 2.15.28A9.97 9.97 0 0 0 12.07 9H12a9.97 9.97 0 0 0-2.15-4.72A7.98 7.98 0 0 1 12 4zm-3.5.93A11.97 11.97 0 0 1 10.06 9H4.26A8.02 8.02 0 0 1 8.5 4.93zM4 12c0-.34.02-.67.06-1h6.01c-.04.33-.07.66-.07 1s.03.67.07 1H4.06A8.1 8.1 0 0 1 4 12zm.26 3h5.8a11.97 11.97 0 0 1-1.56 4.07A8.02 8.02 0 0 1 4.26 15zm7.81 5c-.74 0-1.46-.1-2.15-.28A9.97 9.97 0 0 0 11.93 15H12a9.97 9.97 0 0 0 2.15 4.72A7.98 7.98 0 0 1 12 20zm3.5-.93A11.97 11.97 0 0 1 13.94 15h5.8a8.02 8.02 0 0 1-4.24 4.07zM20 12c0 .34-.02.67-.06 1h-6.01c.04-.33.07-.66.07-1s-.03-.67-.07-1h6.01c.04.33.06.66.06 1zm-.26-3h-5.8a11.97 11.97 0 0 1 1.56-4.07A8.02 8.02 0 0 1 19.74 9z" />
    </svg>
  )
}

function SoccerIcon({ size = 14 }) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" width={size} height={size}>
      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 14.93V15h2v1.93c-.32.04-.66.07-1 .07s-.68-.03-1-.07zM11 13v-2h2v2h-2zm-2 1.72c-.69-.3-1.31-.72-1.84-1.23l1.23-1.23L9.6 13.5l-1.23 1.23.63-.01zM8.34 11H7v-2h1.18c-.05.32-.08.66-.08 1 0 .34.03.67.08 1H8.34zm-1.52-4h1.23l1.23 1.23L8.05 9.46 6.82 8.23c.3-.43.67-.83 1.08-1.14l-.08-.09zM13 9.07V7h-2v2.07c-.35.14-.65.35-.93.58L8.84 8.42C9.72 7.52 10.8 7 12 7c1.2 0 2.28.52 3.16 1.42l-1.23 1.23c-.28-.23-.58-.44-.93-.58zm3.66-1c.41.31.78.71 1.08 1.14l-1.23 1.23L15.28 9.46l1.23-1.23.15.84zM15.66 11c.05.33.08.66.08 1 0 .34-.03.67-.08 1H15v-2h.66zm.18 4.23l-1.23-1.23 1.23-1.23c.53.52.95 1.14 1.23 1.84l-1.23-.38zm-1.89.27L12.72 17H11v-1.93c.32.04.66.07 1 .07.34 0 .68-.03 1-.07v.07l1.17-.24-.22-.57z"/>
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
