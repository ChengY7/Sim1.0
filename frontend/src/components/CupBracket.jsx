import styles from './CupBracket.module.css'
import { espnLogo } from '../utils/espnLogo'

export default function CupBracket({ cup }) {
  const sim = cup !== null

  // Build a team-ID → seed-number (1–4) lookup from both conferences.
  const seedMap = {}
  if (cup) {
    cup.west.seeds.forEach((id, i) => { seedMap[id] = i + 1 })
    cup.east.seeds.forEach((id, i) => { seedMap[id] = i + 1 })
  }

  return (
    <div className={styles.bracketWrap}>
      <div className={styles.bracket}>

        {/* ── West QF ── */}
        <div className={styles.qfCol}>
          <div className={styles.colLabel}>
            <span className={styles.conf}>WEST</span>
            <span>Quarterfinals</span>
          </div>
          <GameCard game={cup?.west.qf[0]} tagA="1" tagB="4" sim={sim} seedMap={seedMap} />
          <GameCard game={cup?.west.qf[1]} tagA="2" tagB="3" sim={sim} seedMap={seedMap} />
        </div>

        <div className={styles.connector}>
          <div className={styles.connFork}>
            <div className={styles.connTop} />
            <div className={styles.connBot} />
          </div>
          <div className={styles.connH} />
        </div>

        {/* ── West SF ── */}
        <div className={styles.sfCol}>
          <div className={styles.colLabel}>Semifinal</div>
          <div className={styles.sfGameWrap}>
            <GameCard game={cup?.west.sf} sim={sim} seedMap={seedMap} />
          </div>
        </div>

        <div className={styles.arrowCol}><span className={styles.arrow}>›</span></div>

        {/* ── Cup Final ── */}
        <div className={styles.finalCol}>
          <div className={`${styles.colLabel} ${styles.finalColLabel}`}>NBA Cup Final</div>
          <GameCard game={cup?.final} sim={sim} isFinal seedMap={seedMap} />
          {sim && cup?.final?.winner && (
            <div className={styles.champion}>
              <TrophyIcon />
              <span>{cup.final.winner}</span>
            </div>
          )}
        </div>

        <div className={styles.arrowCol}><span className={styles.arrow}>‹</span></div>

        {/* ── East SF ── */}
        <div className={styles.sfCol}>
          <div className={styles.colLabel}>Semifinal</div>
          <div className={styles.sfGameWrap}>
            <GameCard game={cup?.east.sf} sim={sim} seedMap={seedMap} />
          </div>
        </div>

        <div className={`${styles.connector} ${styles.connectorMirror}`}>
          <div className={styles.connH} />
          <div className={styles.connFork}>
            <div className={styles.connTop} />
            <div className={styles.connBot} />
          </div>
        </div>

        {/* ── East QF ── */}
        <div className={styles.qfCol}>
          <div className={styles.colLabel}>
            <span>Quarterfinals</span>
            <span className={styles.conf}>EAST</span>
          </div>
          <GameCard game={cup?.east.qf[0]} tagA="1" tagB="4" sim={sim} seedMap={seedMap} />
          <GameCard game={cup?.east.qf[1]} tagA="2" tagB="3" sim={sim} seedMap={seedMap} />
        </div>

      </div>
    </div>
  )
}

/* ── Game card ──────────────────────────────────────────────── */

function GameCard({ game, tagA, tagB, sim, isFinal, seedMap = {} }) {
  const homeWon = sim && game?.winner === game?.home
  const awayWon = sim && game?.winner === game?.away

  return (
    <div className={`${styles.card} ${isFinal ? styles.cardFinal : ''}`}>
      {sim ? (
        <>
          <TeamRow teamId={game.home} score={game.home_score} won={homeWon} seed={seedMap[game.home]} />
          <div className={styles.divider} />
          <TeamRow teamId={game.away} score={game.away_score} won={awayWon} seed={seedMap[game.away]} />
        </>
      ) : (
        <>
          <EmptyRow seed={tagA} />
          <div className={styles.divider} />
          <EmptyRow seed={tagB} />
        </>
      )}
    </div>
  )
}

function TeamRow({ teamId, score, won, seed }) {
  return (
    <div className={`${styles.teamRow} ${won ? styles.rowWon : styles.rowLost}`}>
      <span className={styles.seedNum}>{seed}</span>
      <img
        src={espnLogo(teamId)}
        alt={teamId}
        className={styles.logo}
        onError={e => { e.currentTarget.style.opacity = '0' }}
      />
      <span className={styles.abbr}>{teamId}</span>
      <span className={styles.score}>{score}</span>
    </div>
  )
}

function EmptyRow({ seed }) {
  return (
    <div className={`${styles.teamRow} ${styles.rowEmpty}`}>
      <div className={styles.logoGhost} />
      <span className={styles.seedTag}>{seed ? `Seed ${seed}` : 'TBD'}</span>
      <span className={styles.dash}>—</span>
    </div>
  )
}

function TrophyIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" width="14" height="14">
      <path d="M19 5h-2V3H7v2H5c-1.1 0-2 .9-2 2v1c0 2.55 1.92 4.63 4.39 4.94A5.01 5.01 0 0 0 11 15.9V18H9v2h6v-2h-2v-2.1a5.01 5.01 0 0 0 3.61-2.96C19.08 12.63 21 10.55 21 8V7c0-1.1-.9-2-2-2zM5 8V7h2v3.82C5.84 10.4 5 9.3 5 8zm14 0c0 1.3-.84 2.4-2 2.82V7h2v1z"/>
    </svg>
  )
}
