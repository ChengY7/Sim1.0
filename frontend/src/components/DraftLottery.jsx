import { useState, useEffect, useMemo } from 'react'
import { simulateDraftLottery } from '../api'
import { espnLogo } from '../utils/espnLogo'
import styles from './DraftLottery.module.css'

// ── Static probability table ───────────────────────────────────────────────
// ODDS[i][j] = probability (%) that lottery seed i+1 receives pick j+1
const ODDS = [
  [14.0, 13.4, 12.7, 12.0, 47.9, 0,    0,    0,    0,    0,    0,    0,    0,    0   ],
  [14.0, 13.4, 12.7, 12.0, 27.8, 20.0, 0,    0,    0,    0,    0,    0,    0,    0   ],
  [14.0, 13.4, 12.7, 12.0, 14.8, 26.0, 7.0,  0,    0,    0,    0,    0,    0,    0   ],
  [12.5, 12.2, 11.9, 11.5, 7.2,  25.7, 16.7, 2.2,  0,    0,    0,    0,    0,    0   ],
  [10.5, 10.5, 10.6, 10.5, 2.2,  19.6, 26.7, 8.7,  0.6,  0,    0,    0,    0,    0   ],
  [9.0,  9.2,  9.4,  9.6,  0,    8.6,  29.8, 20.5, 3.7,  0.1,  0,    0,    0,    0   ],
  [7.5,  7.8,  8.1,  8.5,  0,    0,    19.7, 34.1, 12.9, 1.3,  0.01, 0,    0,    0   ],
  [6.0,  6.3,  6.7,  7.2,  0,    0,    0,    34.5, 32.1, 6.7,  0.4,  0.01, 0,    0   ],
  [4.5,  4.8,  5.2,  5.7,  0,    0,    0,    0,    50.7, 25.9, 3.0,  0.1,  0.01, 0   ],
  [3.0,  3.3,  3.6,  4.0,  0,    0,    0,    0,    0,    65.9, 19.0, 1.2,  0.01, 0.01],
  [2.0,  2.2,  2.4,  2.8,  0,    0,    0,    0,    0,    0,    77.6, 12.6, 0.4,  0.01],
  [1.5,  1.7,  1.9,  2.1,  0,    0,    0,    0,    0,    0,    0,    86.1, 6.7,  0.1 ],
  [1.0,  1.1,  1.2,  1.4,  0,    0,    0,    0,    0,    0,    0,    0,    92.9, 2.3 ],
  [0.5,  0.6,  0.6,  0.7,  0,    0,    0,    0,    0,    0,    0,    0,    0,    97.6],
]
const AVG  = [3.7, 3.9, 4.1, 4.4, 5.0, 5.5, 6.2, 7.0, 8.0, 9.2, 10.3, 11.4, 12.5, 13.7]
const TOP4 = ODDS.map(row => +(row[0] + row[1] + row[2] + row[3]).toFixed(1))
const P1   = ODDS.map(row => row[0])

function fmtPct(val) {
  if (val === 0) return ''
  if (val < 0.1) return '>0'
  return val.toFixed(1)
}

export default function DraftLottery({ eastStandings, westStandings, playin }) {
  const seasonSimulated = eastStandings !== null
  const playinSimulated = playin !== null

  const [lotterySeeds, setLotterySeeds] = useState([])
  const [lotteryResult, setLotteryResult] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [subTab, setSubTab] = useState('odds')

  // Build team lookup: teamId → season stat row
  const teamMap = useMemo(() => {
    const m = {}
    if (eastStandings) eastStandings.forEach(t => { m[t.team_id] = t })
    if (westStandings) westStandings.forEach(t => { m[t.team_id] = t })
    return m
  }, [eastStandings, westStandings])

  // Recompute lottery seed order whenever season / play-in results change
  useEffect(() => {
    setLotteryResult(null)

    if (!eastStandings || !westStandings) {
      setLotterySeeds([])
      return
    }

    // Guaranteed lottery teams: conference positions 11-15 (0-indexed: 10-14)
    const guaranteed = [
      ...eastStandings.slice(10),
      ...westStandings.slice(10),
    ].map(t => ({ teamId: t.team_id, w: t.w, l: t.l, name: t.team_name }))

    if (!playin) {
      // Sort worst-to-best, preserve API order for ties
      guaranteed.sort((a, b) => a.w - b.w)
      const seeds = guaranteed.slice(0, 10)
      while (seeds.length < 14) seeds.push(null) // TBD play-in slots
      setLotterySeeds(seeds)
      return
    }

    // Play-in losers: G2 loser (9/10 seed eliminated) + G3 loser (final miss)
    const loserIds = [
      playin.east.game2.loser,
      playin.east.game3.loser,
      playin.west.game2.loser,
      playin.west.game3.loser,
    ]
    const losers = loserIds
      .map(id => teamMap[id])
      .filter(Boolean)
      .map(t => ({ teamId: t.team_id, w: t.w, l: t.l, name: t.team_name }))

    // Combine all 14, sort worst-first with random coin flip for ties
    const all14 = [...guaranteed, ...losers].map(t => ({ ...t, rand: Math.random() }))
    all14.sort((a, b) => a.w !== b.w ? a.w - b.w : a.rand - b.rand)
    setLotterySeeds(all14.map(({ rand, ...rest }) => rest))
  }, [eastStandings, westStandings, playin]) // eslint-disable-line react-hooks/exhaustive-deps

  const canSimulate = seasonSimulated && playinSimulated &&
    lotterySeeds.length === 14 && lotterySeeds.every(s => s !== null)

  async function handleSimulate() {
    if (!canSimulate) return
    setError(null)
    setLoading(true)
    try {
      const teams = lotterySeeds.map(s => s.teamId)
      const data = await simulateDraftLottery(teams)
      setLotteryResult(data)
      setSubTab('results')
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const note = !seasonSimulated
    ? 'Base odds without ties — simulate a season to seed teams'
    : !playinSimulated
      ? 'Seeds 11–14 are TBD — simulate play-in to complete seeding'
      : lotteryResult
        ? 'Lottery complete'
        : 'All 14 teams seeded — ready for the lottery'

  return (
    <div className={styles.wrap}>
      {error && <div className={styles.errorBox}>{error}</div>}

      <div className={styles.header}>
        <div className={styles.subTabs}>
          <button
            className={`${styles.subTab} ${subTab === 'odds' ? styles.subTabActive : ''}`}
            onClick={() => setSubTab('odds')}
          >
            Odds
          </button>
          <button
            className={`${styles.subTab} ${subTab === 'results' ? styles.subTabActive : ''} ${!lotteryResult ? styles.subTabDisabled : ''}`}
            onClick={() => lotteryResult && setSubTab('results')}
          >
            Results
          </button>
        </div>
        <span className={styles.note}>{note}</span>
        <div className={styles.headerRight}>
          <button
            className={styles.btnSim}
            disabled={!canSimulate || loading}
            onClick={handleSimulate}
          >
            {loading ? <span className={styles.spinner} /> : <BallIcon />}
            <span>{loading ? 'Drawing…' : 'Simulate Lottery'}</span>
          </button>
        </div>
      </div>

      {subTab === 'results' && lotteryResult ? (
        <ResultsTable picks={lotteryResult.picks} teamMap={teamMap} />
      ) : (
        <OddsTable seeds={lotterySeeds} seasonSimulated={seasonSimulated} />
      )}
    </div>
  )
}

// ── Odds table ─────────────────────────────────────────────────────────────

function OddsTable({ seeds, seasonSimulated }) {
  return (
    <div className={styles.tableWrap}>
      <table className={styles.table}>
        <thead>
          <tr>
            <th className={styles.thSeed}>
              {seasonSimulated ? 'Seed · Team' : 'Seed'}
            </th>
            {Array.from({ length: 14 }, (_, i) => (
              <th key={i} className={styles.thPick}>{i + 1}</th>
            ))}
            <th className={styles.thAvg}>AVG</th>
          </tr>
        </thead>
        <tbody>
          {ODDS.map((row, i) => {
            const seed = seeds[i]
            return (
              <tr key={i} className={styles.row}>
                <td className={styles.tdSeed}>
                  {!seasonSimulated ? (
                    <span className={styles.seedNum}>{i + 1}</span>
                  ) : seed ? (
                    <SeedTeamCell num={i + 1} seed={seed} />
                  ) : (
                    <TbdCell num={i + 1} />
                  )}
                </td>
                {row.map((p, j) => (
                  <td
                    key={j}
                    className={`${styles.tdPick} ${p >= 20 ? styles.pickHigh : p > 0 && p < 0.1 ? styles.pickTiny : ''}`}
                  >
                    {fmtPct(p)}
                  </td>
                ))}
                <td className={styles.tdAvg}>{AVG[i]}</td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

function SeedTeamCell({ num, seed }) {
  return (
    <div className={styles.teamCell}>
      <span className={styles.seedBadge}>{num}</span>
      <img
        src={espnLogo(seed.teamId)}
        alt={seed.teamId}
        className={styles.teamLogo}
        onError={e => { e.currentTarget.style.opacity = '0' }}
      />
      <span className={styles.teamAbbr}>{seed.teamId}</span>
      <span className={styles.teamRecord}>{seed.w}–{seed.l}</span>
    </div>
  )
}

function TbdCell({ num }) {
  return (
    <div className={styles.teamCell}>
      <span className={styles.seedBadge}>{num}</span>
      <div className={styles.logoGhost} />
      <span className={styles.tbdText}>TBD</span>
    </div>
  )
}

// ── Results table (post-lottery) ───────────────────────────────────────────

function ResultsTable({ picks, teamMap }) {
  return (
    <div className={styles.resultWrap}>
      <table className={styles.resultTable}>
        <thead>
          <tr>
            <th className={styles.rthPick}>Pick</th>
            <th className={styles.rthTeam}>Team</th>
            <th className={styles.rthNum}>Record</th>
            <th className={styles.rthNum}>Top 4 %</th>
            <th className={styles.rthNum}>#1 Overall %</th>
            <th className={styles.rthNum}>Pick %</th>
          </tr>
        </thead>
        <tbody>
          {picks.map(pick => {
            const team = teamMap[pick.team_id]
            const isFirst = pick.pick === 1
            const top4    = TOP4[pick.seed - 1]
            const p1      = P1[pick.seed - 1]
            const pickPct = ODDS[pick.seed - 1][pick.pick - 1]

            return (
              <tr key={pick.pick} className={`${styles.rRow} ${isFirst ? styles.rRowFirst : ''}`}>
                <td className={styles.rtdPick}>
                  {isFirst
                    ? <span className={styles.pick1Badge}>#1</span>
                    : <span className={styles.pickNum}>{pick.pick}</span>
                  }
                </td>
                <td className={styles.rtdTeam}>
                  <div className={styles.resultTeamCell}>
                    <img
                      src={espnLogo(pick.team_id)}
                      alt={pick.team_id}
                      className={styles.resultLogo}
                      onError={e => { e.currentTarget.style.opacity = '0' }}
                    />
                    <span className={styles.resultAbbr}>{pick.team_id}</span>
                    {team && <span className={styles.resultName}>{team.team_name}</span>}
                  </div>
                </td>
                <td className={styles.rtdNum}>
                  {team ? `${team.w}–${team.l}` : '—'}
                </td>
                <td className={styles.rtdNum}>{top4}%</td>
                <td className={styles.rtdNum}>{p1}%</td>
                <td className={styles.rtdNum}>{fmtPct(pickPct)}{pickPct ? '%' : ''}</td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

function BallIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" width="15" height="15">
      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 14H9V8h2v8zm4 0h-2V8h2v8z"/>
    </svg>
  )
}
