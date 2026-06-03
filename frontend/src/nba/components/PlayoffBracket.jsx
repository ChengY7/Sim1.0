import styles from './PlayoffBracket.module.css'
import { espnLogo } from '../../utils/espnLogo'

const SLOT = 130           // px height per R1 matchup slot
const BH   = SLOT * 4     // total bracket body height = 520px
const CC   = 'rgba(255,255,255,0.12)'  // connector color

// ── Entry point ────────────────────────────────────────────────────────────

export default function PlayoffBracket({
  eastSeeds, westSeeds,
  playoffResult,
  onSimulate, loading,
  seasonSimulated, playinSimulated,
}) {
  const simulated   = playoffResult !== null
  const canSimulate = seasonSimulated && playinSimulated && !simulated

  const note = !seasonSimulated
    ? 'Simulate a season to populate the bracket'
    : !playinSimulated
      ? 'Seeds 7 & 8 are TBD — simulate play-in first'
      : simulated ? null
        : 'Ready to simulate playoffs'

  // seed lookup: teamId → seed number
  const seedOf = {}
  ;[...eastSeeds, ...westSeeds].forEach(s => { if (s?.teamId) seedOf[s.teamId] = s.seed })

  const team = (seeds, n) => seeds.find(s => s?.seed === n)

  // shortcuts into result
  const wr1  = playoffResult?.west.r1
  const wr2  = playoffResult?.west.r2
  const wr3  = playoffResult?.west.r3
  const er1  = playoffResult?.east.r1
  const er2  = playoffResult?.east.r2
  const er3  = playoffResult?.east.r3
  const fins = playoffResult?.finals

  // Spatial top/bottom for each round.
  // R2[i] top = winner of R1[i*2], bottom = winner of R1[i*2+1].
  // CF top = R2[0] winner, CF bottom = R2[1] winner.
  // Finals top = home_team (backend sets home = better record).
  function r2Display(r1arr, i) {
    return { top: r1arr?.[i * 2]?.winner, bot: r1arr?.[i * 2 + 1]?.winner }
  }

  // R1 matchup order: [lowerSeed, higherSeed]
  const R1_PAIRS = [[1,8],[4,5],[3,6],[2,7]]

  return (
    <div className={styles.wrap}>
      {/* ── Header ── */}
      <div className={styles.header}>
        {simulated && (
          <div className={styles.champion}>
            <TrophyIcon />
            <span>Champion:</span>
            <img
              src={espnLogo(playoffResult.champion)}
              alt=""
              className={styles.champLogo}
              onError={e => { e.currentTarget.style.opacity = '0' }}
            />
            <strong>{playoffResult.champion}</strong>
          </div>
        )}
        {note && <span className={styles.note}>{note}</span>}
        <button
          className={styles.btnSim}
          disabled={!canSimulate || loading}
          onClick={onSimulate}
        >
          {loading ? <span className={styles.spinner} /> : <TrophyIcon />}
          <span>{loading ? 'Simulating…' : 'Simulate Playoffs'}</span>
        </button>
      </div>

      {/* ── Bracket ── */}
      <div className={styles.bracketOuter}>
        <div className={styles.bracketRow}>

          {/* ══ WEST ══ */}

          {/* W-R1 */}
          <BCol label="WEST · 1st Round">
            {R1_PAIRS.map(([ts, bs], i) => (
              <Slot key={i} h={SLOT}>
                <SeriesCard
                  topId={team(westSeeds, ts)?.teamId} topSeed={ts}
                  botId={team(westSeeds, bs)?.teamId} botSeed={bs}
                  topPh={`Seed ${ts}`} botPh={`Seed ${bs}`}
                  series={wr1?.[i]}
                  simulated={simulated} seasonSimulated={seasonSimulated}
                />
              </Slot>
            ))}
          </BCol>

          <Conn12 />

          {/* W-R2 */}
          <BCol label="2nd Round">
            {[0, 1].map(i => {
              const { top, bot } = simulated ? r2Display(wr1, i) : {}
              return (
                <Slot key={i} h={SLOT * 2}>
                  <SeriesCard
                    topId={top} topSeed={seedOf[top]}
                    botId={bot} botSeed={seedOf[bot]}
                    topPh={i === 0 ? 'W of 1-8' : 'W of 3-6'}
                    botPh={i === 0 ? 'W of 4-5' : 'W of 2-7'}
                    series={wr2?.[i]}
                    simulated={simulated} seasonSimulated={seasonSimulated}
                  />
                </Slot>
              )
            })}
          </BCol>

          <Conn23 />

          {/* W-CF */}
          <BCol label="Conf Finals">
            <Slot h={BH}>
              <SeriesCard
                topId={simulated ? wr2?.[0]?.winner : null}
                topSeed={seedOf[wr2?.[0]?.winner]}
                botId={simulated ? wr2?.[1]?.winner : null}
                botSeed={seedOf[wr2?.[1]?.winner]}
                topPh="W of R2 top" botPh="W of R2 bot"
                series={wr3}
                simulated={simulated} seasonSimulated={seasonSimulated}
              />
            </Slot>
          </BCol>

          {/* CF→Finals arrow */}
          <Arrow />

          {/* ══ FINALS ══ */}
          <BCol label="NBA Finals" gold>
            <Slot h={BH}>
              <SeriesCard
                topId={fins?.home_team} topSeed={seedOf[fins?.home_team]}
                botId={fins?.away_team} botSeed={seedOf[fins?.away_team]}
                topPh="W Conf" botPh="E Conf"
                series={fins}
                simulated={simulated} seasonSimulated={seasonSimulated}
                isFinals
              />
            </Slot>
          </BCol>

          {/* Finals→CF arrow */}
          <Arrow mirror />

          {/* ══ EAST ══ */}

          {/* E-CF */}
          <BCol label="Conf Finals">
            <Slot h={BH}>
              <SeriesCard
                topId={simulated ? er2?.[0]?.winner : null}
                topSeed={seedOf[er2?.[0]?.winner]}
                botId={simulated ? er2?.[1]?.winner : null}
                botSeed={seedOf[er2?.[1]?.winner]}
                topPh="W of R2 top" botPh="W of R2 bot"
                series={er3}
                simulated={simulated} seasonSimulated={seasonSimulated}
              />
            </Slot>
          </BCol>

          <Conn23 mirror />

          {/* E-R2 */}
          <BCol label="2nd Round">
            {[0, 1].map(i => {
              const { top, bot } = simulated ? r2Display(er1, i) : {}
              return (
                <Slot key={i} h={SLOT * 2}>
                  <SeriesCard
                    topId={top} topSeed={seedOf[top]}
                    botId={bot} botSeed={seedOf[bot]}
                    topPh={i === 0 ? 'W of 1-8' : 'W of 3-6'}
                    botPh={i === 0 ? 'W of 4-5' : 'W of 2-7'}
                    series={er2?.[i]}
                    simulated={simulated} seasonSimulated={seasonSimulated}
                  />
                </Slot>
              )
            })}
          </BCol>

          <Conn12 mirror />

          {/* E-R1 */}
          <BCol label="1st Round · EAST">
            {R1_PAIRS.map(([ts, bs], i) => (
              <Slot key={i} h={SLOT}>
                <SeriesCard
                  topId={team(eastSeeds, ts)?.teamId} topSeed={ts}
                  botId={team(eastSeeds, bs)?.teamId} botSeed={bs}
                  topPh={`Seed ${ts}`} botPh={`Seed ${bs}`}
                  series={er1?.[i]}
                  simulated={simulated} seasonSimulated={seasonSimulated}
                />
              </Slot>
            ))}
          </BCol>

        </div>
      </div>
    </div>
  )
}

// ── Layout primitives ───────────────────────────────────────────────────────

const LABEL_H = 26  // px

function BCol({ label, gold, children }) {
  return (
    <div className={styles.bcol} style={{ flexShrink: 0 }}>
      <div className={`${styles.colLabel} ${gold ? styles.colLabelGold : ''}`}>{label}</div>
      <div style={{ height: BH }}>
        {children}
      </div>
    </div>
  )
}

function Slot({ h, children }) {
  return (
    <div style={{ height: h, display: 'flex', alignItems: 'center' }}>
      {children}
    </div>
  )
}

// Connector between R1 (4 cards) and R2 (2 cards) — two forks
function Conn12({ mirror }) {
  return (
    <div style={{ width: 12, height: BH + LABEL_H, display: 'flex', flexDirection: 'column', flexShrink: 0 }}>
      <div style={{ height: LABEL_H }} />
      <div style={{ height: SLOT / 2 }} />
      <Fork h={SLOT}     mirror={mirror} />
      <div style={{ height: SLOT }} />
      <Fork h={SLOT}     mirror={mirror} />
      <div style={{ height: SLOT / 2 }} />
    </div>
  )
}

// Connector between R2 (2 cards) and CF (1 card) — one tall fork
function Conn23({ mirror }) {
  return (
    <div style={{ width: 12, height: BH + LABEL_H, display: 'flex', flexDirection: 'column', flexShrink: 0 }}>
      <div style={{ height: LABEL_H }} />
      <div style={{ height: SLOT }} />
      <Fork h={SLOT * 2} mirror={mirror} />
      <div style={{ height: SLOT }} />
    </div>
  )
}

// Single bracket-arm connector
function Fork({ h, mirror }) {
  const spine = (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ flex: 1, borderRight: `1px solid ${CC}`, borderBottom: `1px solid ${CC}`, borderBottomRightRadius: 3 }} />
      <div style={{ flex: 1, borderRight: `1px solid ${CC}`, borderTop:    `1px solid ${CC}`, borderTopRightRadius:    3 }} />
    </div>
  )
  const horiz = <div style={{ width: 6, height: 1, background: CC, alignSelf: 'center', flexShrink: 0 }} />
  return (
    <div style={{ height: h, display: 'flex', alignItems: 'stretch' }}>
      {mirror ? <>{horiz}{spine}</> : <>{spine}{horiz}</>}
    </div>
  )
}

// Simple arrow between CF and Finals
function Arrow({ mirror }) {
  return (
    <div className={styles.arrow} style={{ height: BH + LABEL_H }}>
      {mirror ? '‹' : '›'}
    </div>
  )
}

// ── Series card ─────────────────────────────────────────────────────────────

function SeriesCard({ topId, topSeed, botId, botSeed, topPh, botPh, series, simulated, seasonSimulated, isFinals }) {
  // Wins per display position
  let topWins = null, botWins = null
  if (simulated && series && topId && botId) {
    const topIsHome = series.home_team === topId
    topWins = topIsHome ? series.home_wins : series.away_wins
    botWins = topIsHome ? series.away_wins : series.home_wins
  }
  const topWon = simulated && !!topId && series?.winner === topId
  const botWon = simulated && !!botId && series?.winner === botId

  return (
    <div className={`${styles.card} ${isFinals ? styles.cardFinals : ''}`}>
      <TeamRow id={topId} seed={topSeed} ph={topPh} wins={topWins} won={topWon} simulated={simulated} seasonSimulated={seasonSimulated} />
      <div className={styles.divider} />
      <TeamRow id={botId} seed={botSeed} ph={botPh} wins={botWins} won={botWon} simulated={simulated} seasonSimulated={seasonSimulated} />
    </div>
  )
}

function TeamRow({ id, seed, ph, wins, won, simulated, seasonSimulated }) {
  const hasTeam = seasonSimulated && !!id
  return (
    <div className={`${styles.teamRow} ${!hasTeam ? styles.rowEmpty : simulated && won ? styles.rowWon : simulated ? styles.rowLost : ''}`}>
      <span className={styles.seedNum}>{seed ?? ''}</span>
      {hasTeam
        ? <img src={espnLogo(id)} alt={id} className={styles.logo} onError={e => { e.currentTarget.style.opacity = '0' }} />
        : <div className={styles.logoGhost} />
      }
      <span className={styles.abbr}>{hasTeam ? id : ph}</span>
      <span className={`${styles.wins} ${!simulated || wins === null ? styles.dash : ''}`}>
        {simulated && wins !== null ? wins : '—'}
      </span>
    </div>
  )
}

// ── Icons ───────────────────────────────────────────────────────────────────

function TrophyIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" width="15" height="15">
      <path d="M19 5h-2V3H7v2H5c-1.1 0-2 .9-2 2v1c0 2.55 1.92 4.63 4.39 4.94A5.01 5.01 0 0 0 11 15.9V18H9v2h6v-2h-2v-2.1a5.01 5.01 0 0 0 3.61-2.96C19.08 12.63 21 10.55 21 8V7c0-1.1-.9-2-2-2zM5 8V7h2v3.82C5.84 10.4 5 9.3 5 8zm14 0c0 1.3-.84 2.4-2 2.82V7h2v1z"/>
    </svg>
  )
}
