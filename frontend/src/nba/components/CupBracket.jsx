import styles from './CupBracket.module.css'
import { espnLogo } from '../../utils/espnLogo'

// Layout constants
const SLOT    = 88     // px per QF slot → BH = 4 * SLOT = 352px
const BH      = SLOT * 4
const LABEL_H = 24     // column label height
const CC      = 'rgba(255,255,255,0.12)'  // connector color

export default function CupBracket({ cup }) {
  const sim = cup !== null

  // seed lookup: teamId → cup seed (1-4)
  const seedOf = {}
  if (cup) {
    cup.west.seeds.forEach((id, i) => { if (id) seedOf[id] = i + 1 })
    cup.east.seeds.forEach((id, i) => { if (id) seedOf[id] = i + 1 })
  }

  // SF spatial order: top = winner of QF[0], bottom = winner of QF[1]
  const wSFTop = cup?.west.qf[0]?.winner
  const wSFBot = cup?.west.qf[1]?.winner
  const eSFTop = cup?.east.qf[0]?.winner
  const eSFBot = cup?.east.qf[1]?.winner

  // QF seed labels
  const wSeeds = cup?.west.seeds ?? []
  const eSeeds = cup?.east.seeds ?? []

  return (
    <div className={styles.bracketWrap}>
      <div className={styles.bracket}>

        {/* ── Quarterfinals (4 games) ── */}
        <BCol label="Quarterfinals">
          {/* West QF */}
          <Slot h={SLOT}>
            <GameCard
              game={sim ? cup.west.qf[0] : null}
              topId={wSeeds[0]} topSeed={seedOf[wSeeds[0]] ?? 1}
              botId={wSeeds[3]} botSeed={seedOf[wSeeds[3]] ?? 4}
              topPh="W · 1" botPh="W · 4"
              sim={sim} confBadge="W"
            />
          </Slot>
          <Slot h={SLOT}>
            <GameCard
              game={sim ? cup.west.qf[1] : null}
              topId={wSeeds[1]} topSeed={seedOf[wSeeds[1]] ?? 2}
              botId={wSeeds[2]} botSeed={seedOf[wSeeds[2]] ?? 3}
              topPh="W · 2" botPh="W · 3"
              sim={sim}
            />
          </Slot>
          {/* Conference divider */}
          <div className={styles.confDivider} />
          {/* East QF */}
          <Slot h={SLOT}>
            <GameCard
              game={sim ? cup.east.qf[0] : null}
              topId={eSeeds[0]} topSeed={seedOf[eSeeds[0]] ?? 1}
              botId={eSeeds[3]} botSeed={seedOf[eSeeds[3]] ?? 4}
              topPh="E · 1" botPh="E · 4"
              sim={sim} confBadge="E"
            />
          </Slot>
          <Slot h={SLOT}>
            <GameCard
              game={sim ? cup.east.qf[1] : null}
              topId={eSeeds[1]} topSeed={seedOf[eSeeds[1]] ?? 2}
              botId={eSeeds[2]} botSeed={seedOf[eSeeds[2]] ?? 3}
              topPh="E · 2" botPh="E · 3"
              sim={sim}
            />
          </Slot>
        </BCol>

        <Conn12 />

        {/* ── Semifinals (2 games) ── */}
        <BCol label="Semifinals">
          <Slot h={SLOT * 2}>
            <GameCard
              game={sim ? cup.west.sf : null}
              topId={sim ? wSFTop : null} topSeed={seedOf[wSFTop]}
              botId={sim ? wSFBot : null} botSeed={seedOf[wSFBot]}
              topPh="W of QF1" botPh="W of QF2"
              sim={sim} confBadge="W"
            />
          </Slot>
          <Slot h={SLOT * 2}>
            <GameCard
              game={sim ? cup.east.sf : null}
              topId={sim ? eSFTop : null} topSeed={seedOf[eSFTop]}
              botId={sim ? eSFBot : null} botSeed={seedOf[eSFBot]}
              topPh="W of QF1" botPh="W of QF2"
              sim={sim} confBadge="E"
            />
          </Slot>
        </BCol>

        <Conn23 />

        {/* ── Final ── */}
        <BCol label="NBA Cup Final" gold>
          <Slot h={BH}>
            <GameCard
              game={sim ? cup.final : null}
              topId={sim ? cup.final.home : null} topSeed={seedOf[cup?.final?.home]}
              botId={sim ? cup.final.away : null} botSeed={seedOf[cup?.final?.away]}
              topPh="W Conf" botPh="E Conf"
              sim={sim} isFinal
            />
          </Slot>
        </BCol>

        {/* ── Champion ── */}
        {sim && cup?.final?.winner && (
          <div className={styles.champion}>
            <TrophyIcon />
            <img
              src={espnLogo(cup.final.winner)}
              alt={cup.final.winner}
              className={styles.champLogo}
              onError={e => { e.currentTarget.style.opacity = '0' }}
            />
            <span>{cup.final.winner}</span>
          </div>
        )}

      </div>
    </div>
  )
}

// ── Layout helpers ──────────────────────────────────────────────────────────

function BCol({ label, gold, children }) {
  return (
    <div className={styles.bcol}>
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

// R1→R2 style: two forks connecting 4 QF slots to 2 SF slots
function Conn12() {
  return (
    <div style={{ width: 12, height: BH + LABEL_H, display: 'flex', flexDirection: 'column', flexShrink: 0 }}>
      <div style={{ height: LABEL_H }} />
      <div style={{ height: SLOT / 2 }} />
      <Fork h={SLOT} />
      <div style={{ height: SLOT }} />
      <Fork h={SLOT} />
      <div style={{ height: SLOT / 2 }} />
    </div>
  )
}

// R2→Final: single fork connecting 2 SF slots to 1 Final slot
function Conn23() {
  return (
    <div style={{ width: 12, height: BH + LABEL_H, display: 'flex', flexDirection: 'column', flexShrink: 0 }}>
      <div style={{ height: LABEL_H }} />
      <div style={{ height: SLOT }} />
      <Fork h={SLOT * 2} />
      <div style={{ height: SLOT }} />
    </div>
  )
}

function Fork({ h }) {
  return (
    <div style={{ height: h, display: 'flex', alignItems: 'stretch' }}>
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
        <div style={{ flex: 1, borderRight: `1px solid ${CC}`, borderBottom: `1px solid ${CC}`, borderBottomRightRadius: 3 }} />
        <div style={{ flex: 1, borderRight: `1px solid ${CC}`, borderTop:    `1px solid ${CC}`, borderTopRightRadius:    3 }} />
      </div>
      <div style={{ width: 6, height: 1, background: CC, alignSelf: 'center', flexShrink: 0 }} />
    </div>
  )
}

// ── Game card ───────────────────────────────────────────────────────────────

function GameCard({ game, topId, topSeed, botId, botSeed, topPh, botPh, sim, isFinal, confBadge }) {
  const topWon  = sim && !!game && game.winner === topId
  const botWon  = sim && !!game && game.winner === botId
  const topHome = sim && !!game && game.home === topId
  const topScore = sim && game ? (topHome ? game.home_score : game.away_score) : null
  const botScore = sim && game ? (topHome ? game.away_score : game.home_score) : null

  return (
    <div className={`${styles.card} ${isFinal ? styles.cardFinal : ''}`}>
      {confBadge && <div className={styles.confBadge}>{confBadge}</div>}
      <TeamRow id={topId} seed={topSeed} ph={topPh} score={topScore} won={topWon} sim={sim} />
      <div className={styles.divider} />
      <TeamRow id={botId} seed={botSeed} ph={botPh} score={botScore} won={botWon} sim={sim} />
    </div>
  )
}

function TeamRow({ id, seed, ph, score, won, sim }) {
  const hasTeam = !!id
  return (
    <div className={`${styles.teamRow} ${!hasTeam ? styles.rowEmpty : sim && won ? styles.rowWon : sim ? styles.rowLost : ''}`}>
      <span className={styles.seedNum}>{seed ?? ''}</span>
      {hasTeam
        ? <img src={espnLogo(id)} alt={id} className={styles.logo} onError={e => { e.currentTarget.style.opacity = '0' }} />
        : <div className={styles.logoGhost} />
      }
      <span className={styles.abbr}>{hasTeam ? id : ph}</span>
      <span className={`${styles.score} ${!sim || score == null ? styles.dash : ''}`}>
        {sim && score != null ? score : '—'}
      </span>
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
