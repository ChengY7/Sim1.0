import { useState, useEffect } from 'react'
import { fetchSeasons, fetchTeams, simulate } from './nba/api'
import { fetchFIFATeams, simulateFIFAMatch, simulateGroupStage } from './fifa/api'
import TeamPicker from './nba/components/TeamPicker'
import Scoreboard from './nba/components/Scoreboard'
import PlayByPlay from './nba/components/PlayByPlay'
import SeasonStandings from './nba/components/SeasonStandings'
import FIFATeamPicker from './fifa/components/TeamPicker'
import GroupStandings from './fifa/components/GroupStandings'
import { flagEmoji } from './fifa/utils/flagEmoji'
import styles from './App.module.css'

export default function App() {
  const [sport, setSport]   = useState('nba')  // 'nba' | 'fifa'

  // ── NBA state ──────────────────────────────────────────────────────────────
  const [nbaMode, setNbaMode]   = useState('game')
  const [teams, setTeams]       = useState([])
  const [seasons, setSeasons]   = useState([])
  const [season, setSeason]     = useState(null)
  const [homeId, setHomeId]     = useState('LAL')
  const [awayId, setAwayId]     = useState('BOS')
  const [nbaResult, setNbaResult] = useState(null)
  const [nbaLoading, setNbaLoading] = useState(false)
  const [nbaError, setNbaError]   = useState(null)

  // ── FIFA state ─────────────────────────────────────────────────────────────
  const [fifaMode, setFifaMode]         = useState('game')
  const [fifaTeams, setFifaTeams]       = useState([])
  const [fifaTeam1, setFifaTeam1]       = useState('ESP')
  const [fifaTeam2, setFifaTeam2]       = useState('FRA')
  const [fifaMatchResult, setFifaMatchResult] = useState(null)
  const [groupResult, setGroupResult]   = useState(null)
  const [fifaLoading, setFifaLoading]   = useState(false)
  const [fifaError, setFifaError]       = useState(null)

  // ── Load NBA data ──────────────────────────────────────────────────────────
  useEffect(() => {
    fetchTeams()
      .then(setTeams)
      .catch(() => setNbaError('Cannot reach the API. Start the backend with: cd backend && go run ./cmd/server'))
    fetchSeasons()
      .then(({ seasons, default_season }) => { setSeasons(seasons); setSeason(default_season) })
      .catch(() => {})
  }, [])

  // ── Load FIFA data ─────────────────────────────────────────────────────────
  useEffect(() => {
    fetchFIFATeams()
      .then(setFifaTeams)
      .catch(() => {})
  }, [])

  // ── NBA handlers ───────────────────────────────────────────────────────────
  function handleSeasonChange(s) {
    if (s === season) return
    setSeason(s)
    setNbaResult(null)
  }

  async function handleNBASimulate() {
    if (homeId === awayId) { setNbaError('Please select two different teams.'); return }
    setNbaError(null); setNbaResult(null); setNbaLoading(true)
    try { setNbaResult(await simulate(homeId, awayId, season)) }
    catch (err) { setNbaError(err.message) }
    finally { setNbaLoading(false) }
  }

  // ── FIFA handlers ──────────────────────────────────────────────────────────
  async function handleFIFASimulateMatch() {
    if (fifaTeam1 === fifaTeam2) { setFifaError('Please select two different teams.'); return }
    setFifaError(null); setFifaMatchResult(null); setFifaLoading(true)
    try { setFifaMatchResult(await simulateFIFAMatch(fifaTeam1, fifaTeam2, false)) }
    catch (err) { setFifaError(err.message) }
    finally { setFifaLoading(false) }
  }

  async function handleSimulateGroupStage() {
    setFifaError(null); setGroupResult(null); setFifaLoading(true)
    try { setGroupResult(await simulateGroupStage()) }
    catch (err) { setFifaError(err.message) }
    finally { setFifaLoading(false) }
  }

  function handleSportChange(s) {
    if (s === sport) return
    setSport(s)
    setFifaMatchResult(null)
    setFifaError(null)
    setNbaError(null)
  }

  const activeModes = sport === 'nba'
    ? [{ id: 'game', label: 'Game Sim', icon: <DiceIcon /> }, { id: 'season', label: 'Season Sim', icon: <CalendarIcon /> }]
    : [{ id: 'game', label: 'Game Sim', icon: <DiceIcon /> }, { id: 'tournament', label: 'World Cup', icon: <TrophyIcon /> }]

  const activeMode   = sport === 'nba' ? nbaMode : fifaMode
  const setActiveMode = sport === 'nba' ? setNbaMode : setFifaMode

  return (
    <div className={styles.app}>
      {/* ── Navbar ── */}
      <nav className={styles.navbar}>
        <div className={styles.logo}>
          <div className={styles.logoIcon}>
            {sport === 'fifa' ? <SoccerBallIcon /> : <BasketballIcon />}
          </div>
          <span className={styles.wordmark}>Sim1.0</span>
        </div>

        <div className={styles.sportTabs}>
          <button className={`${styles.sportTab} ${sport === 'nba' ? styles.sportTabActive : ''}`} onClick={() => handleSportChange('nba')}>
            <BasketballSmallIcon /> NBA
          </button>
          <button className={`${styles.sportTab} ${sport === 'fifa' ? styles.sportTabActive : ''}`} onClick={() => handleSportChange('fifa')}>
            <SoccerSmallIcon /> World Cup
          </button>
        </div>

        <div className={styles.modeTabs}>
          {activeModes.map(m => (
            <button
              key={m.id}
              className={`${styles.modeTab} ${activeMode === m.id ? styles.modeTabActive : ''}`}
              onClick={() => setActiveMode(m.id)}
            >
              {m.icon}{m.label}
            </button>
          ))}
        </div>
      </nav>

      {/* ── NBA: Game Sim ── */}
      {sport === 'nba' && nbaMode === 'game' && (
        <>
          <TeamPicker teams={teams} homeId={homeId} onHomeChange={setHomeId} awayId={awayId} onAwayChange={setAwayId} />
          <div className={styles.simWrap}>
            <button className={styles.btnSim} onClick={handleNBASimulate} disabled={nbaLoading || !teams.length}>
              {nbaLoading ? <span className={styles.spinner} /> : <PlayIcon />}
              <span>{nbaLoading ? 'Simulating…' : 'Simulate Game'}</span>
            </button>
          </div>
          {nbaError && <div className={styles.errorBox}>{nbaError}</div>}
          {nbaResult && (
            <>
              <Scoreboard state={nbaResult.state} seed={nbaResult.seed} homeId={homeId} awayId={awayId} />
              <PlayByPlay events={nbaResult.events} />
            </>
          )}
        </>
      )}

      {/* ── NBA: Season Sim ── */}
      {sport === 'nba' && nbaMode === 'season' && (
        <div className={styles.seasonContent}>
          <div className={styles.seasonBar}>
            <span className={styles.seasonLabel}>Season</span>
            <div className={styles.seasonTabs}>
              {seasons.length > 0
                ? seasons.map(s => (
                    <button key={s.season}
                      className={`${styles.seasonTab} ${season === s.season ? styles.seasonTabActive : ''}`}
                      onClick={() => handleSeasonChange(s.season)}>{s.season}</button>
                  ))
                : season
                  ? <button className={`${styles.seasonTab} ${styles.seasonTabActive}`}>{season}</button>
                  : <span className={styles.seasonLoading}>Loading…</span>
              }
            </div>
          </div>
          <SeasonStandings key={season ?? 'default'} teams={teams} season={season}
            actualStandings={seasons.find(s => s.season === season)?.standings ?? []} />
        </div>
      )}

      {/* ── FIFA: Game Sim ── */}
      {sport === 'fifa' && fifaMode === 'game' && (
        <>
          <FIFATeamPicker
            teams={fifaTeams}
            team1Id={fifaTeam1} onTeam1Change={setFifaTeam1}
            team2Id={fifaTeam2} onTeam2Change={setFifaTeam2}
          />
          <div className={styles.simWrap}>
            <button className={styles.btnSim} onClick={handleFIFASimulateMatch} disabled={fifaLoading || !fifaTeams.length}>
              {fifaLoading ? <span className={styles.spinner} /> : <PlayIcon />}
              <span>{fifaLoading ? 'Simulating…' : 'Simulate Match'}</span>
            </button>
          </div>
          {fifaError && <div className={styles.errorBox}>{fifaError}</div>}
          {fifaMatchResult && <FIFAMatchResult result={fifaMatchResult} teams={fifaTeams} />}
        </>
      )}

      {/* ── FIFA: World Cup (Group Stage) ── */}
      {sport === 'fifa' && fifaMode === 'tournament' && (
        <>
          <div className={styles.simWrap}>
            <button className={styles.btnSim} onClick={handleSimulateGroupStage} disabled={fifaLoading || !fifaTeams.length}>
              {fifaLoading ? <span className={styles.spinner} /> : <PlayIcon />}
              <span>{fifaLoading ? 'Simulating…' : 'Simulate Group Stage'}</span>
            </button>
          </div>
          {fifaError && <div className={styles.errorBox}>{fifaError}</div>}
          <GroupStandings teams={fifaTeams} result={groupResult} />
        </>
      )}
    </div>
  )
}

// ── FIFA Match Result ──────────────────────────────────────────────────────────
function FIFAMatchResult({ result, teams }) {
  const t1 = teams.find(t => t.id === result.team1_id)
  const t2 = teams.find(t => t.id === result.team2_id)

  const score1 = result.team1_goals + (result.team1_goals_aet ?? 0)
  const score2 = result.team2_goals + (result.team2_goals_aet ?? 0)

  const periodLabel = { '90': 'FT', 'aet': 'AET', 'pens': 'Pens' }[result.period] ?? result.period

  return (
    <div className={styles.fifaResult}>
      <div className={`${styles.fifaTeam} ${result.winner === result.team1_id ? styles.fifaWinner : result.winner ? styles.fifaLoser : ''}`}>
        <span className={styles.fifaFlag}>{flagEmoji(result.team1_id)}</span>
        <span className={styles.fifaName}>{t1?.name ?? result.team1_id}</span>
      </div>
      <div className={styles.fifaScore}>
        <span className={styles.fifaScoreNum}>{score1}</span>
        <span className={styles.fifaScoreSep}>–</span>
        <span className={styles.fifaScoreNum}>{score2}</span>
        <div className={styles.fifaPeriod}>{periodLabel}</div>
        {result.penalties && (
          <div className={styles.fifaPens}>({result.penalties.team1_scored}–{result.penalties.team2_scored} pens)</div>
        )}
      </div>
      <div className={`${styles.fifaTeam} ${styles.fifaTeamRight} ${result.winner === result.team2_id ? styles.fifaWinner : result.winner ? styles.fifaLoser : ''}`}>
        <span className={styles.fifaName}>{t2?.name ?? result.team2_id}</span>
        <span className={styles.fifaFlag}>{flagEmoji(result.team2_id)}</span>
      </div>
    </div>
  )
}

// ── Icons ──────────────────────────────────────────────────────────────────────
function BasketballIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="white" width="24" height="24">
      <path d="M12 2C6.477 2 2 6.477 2 12s4.477 10 10 10 10-4.477 10-10S17.523 2 12 2zm0 2c.74 0 1.46.1 2.15.28A9.97 9.97 0 0 0 12.07 9H12a9.97 9.97 0 0 0-2.15-4.72A7.98 7.98 0 0 1 12 4zm-3.5.93A11.97 11.97 0 0 1 10.06 9H4.26A8.02 8.02 0 0 1 8.5 4.93zM4 12c0-.34.02-.67.06-1h6.01c-.04.33-.07.66-.07 1s.03.67.07 1H4.06A8.1 8.1 0 0 1 4 12zm.26 3h5.8a11.97 11.97 0 0 1-1.56 4.07A8.02 8.02 0 0 1 4.26 15zm7.81 5c-.74 0-1.46-.1-2.15-.28A9.97 9.97 0 0 0 11.93 15H12a9.97 9.97 0 0 0 2.15 4.72A7.98 7.98 0 0 1 12 20zm3.5-.93A11.97 11.97 0 0 1 13.94 15h5.8a8.02 8.02 0 0 1-4.24 4.07zM20 12c0 .34-.02.67-.06 1h-6.01c.04-.33.07-.66.07-1s-.03-.67-.07-1h6.01c.04.33.06.66.06 1zm-.26-3h-5.8a11.97 11.97 0 0 1 1.56-4.07A8.02 8.02 0 0 1 19.74 9z" />
    </svg>
  )
}

function SoccerBallIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="white" width="24" height="24">
      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 14.93V15h2v1.93c-.32.04-.66.07-1 .07s-.68-.03-1-.07zM11 13v-2h2v2h-2zm4.12 1.45l-1.41-1.41 1.41-1.41 1.41 1.41-1.41 1.41zM15 9.17l-1.41 1.41-1.41-1.41 1.41-1.41L15 9.17zm-3-3.1V7H9.83l-.59-2.07A7.97 7.97 0 0 1 12 4c.34 0 .68.02 1 .07zM7.59 5.41L8.17 7H6.07C6.57 6.36 7.07 5.85 7.59 5.41zM5.07 9h3.1L7.76 10.41 6.34 9H5.07zM4.07 13H7v2H4.93a8.026 8.026 0 0 1-.86-2zm4.51 4.59L9 16.17l1.41 1.41-1.08 1.08A8.08 8.08 0 0 1 8.58 17.59zm6.84 1.08L14.01 17.59l1.41-1.41.41.41a8.08 8.08 0 0 1-.41.67zM17 15v-2h2.93a8.026 8.026 0 0 1-.86 2H17zm.66-4H16.24l-.41-.41L17.24 9h.76c.23.63.37 1.31.37 2 0 .01 0 .01-.01.01L17.66 11z"/>
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

function SoccerSmallIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" width="14" height="14">
      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 14.93V15h2v1.93c-.32.04-.66.07-1 .07s-.68-.03-1-.07zM11 13v-2h2v2h-2zm4.12 1.45l-1.41-1.41 1.41-1.41 1.41 1.41-1.41 1.41zM15 9.17l-1.41 1.41-1.41-1.41 1.41-1.41L15 9.17zm-3-3.1V7H9.83l-.59-2.07A7.97 7.97 0 0 1 12 4c.34 0 .68.02 1 .07zM7.59 5.41L8.17 7H6.07C6.57 6.36 7.07 5.85 7.59 5.41zM5.07 9h3.1L7.76 10.41 6.34 9H5.07zM4.07 13H7v2H4.93a8.026 8.026 0 0 1-.86-2zm4.51 4.59L9 16.17l1.41 1.41-1.08 1.08A8.08 8.08 0 0 1 8.58 17.59zm6.84 1.08L14.01 17.59l1.41-1.41.41.41a8.08 8.08 0 0 1-.41.67zM17 15v-2h2.93a8.026 8.026 0 0 1-.86 2H17zm.66-4H16.24l-.41-.41L17.24 9h.76c.23.63.37 1.31.37 2 0 .01 0 .01-.01.01L17.66 11z"/>
    </svg>
  )
}

function PlayIcon() {
  return <svg viewBox="0 0 24 24" fill="white" width="16" height="16"><path d="M8 5v14l11-7z" /></svg>
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

function TrophyIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" width="14" height="14">
      <path d="M19 5h-2V3H7v2H5c-1.1 0-2 .9-2 2v1c0 2.55 1.92 4.63 4.39 4.94A5.01 5.01 0 0 0 11 15.9V18H9v2h6v-2h-2v-2.1a5.01 5.01 0 0 0 3.61-2.96C19.08 12.63 21 10.55 21 8V7c0-1.1-.9-2-2-2zM5 8V7h2v3.82C5.84 10.4 5 9.3 5 8zm14 0c0 1.3-.84 2.4-2 2.82V7h2v1z" />
    </svg>
  )
}
