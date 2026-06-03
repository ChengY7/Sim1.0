import styles from './PlayInBracket.module.css'
import { espnLogo } from '../utils/espnLogo'

export default function PlayInBracket({ eastSeeds, westSeeds, playin, seasonSimulated }) {
  const simulated = playin !== null

  return (
    <div className={styles.wrap}>
      <div className={styles.grid}>
        <ConferenceColumn
          conf="WEST"
          seeds={westSeeds}
          result={playin?.west ?? null}
          simulated={simulated}
          seasonSimulated={seasonSimulated}
        />
        <ConferenceColumn
          conf="EAST"
          seeds={eastSeeds}
          result={playin?.east ?? null}
          simulated={simulated}
          seasonSimulated={seasonSimulated}
        />
      </div>
    </div>
  )
}

function ConferenceColumn({ conf, seeds, result, simulated, seasonSimulated }) {
  return (
    <div className={styles.confCol}>
      <div className={styles.confHeader}>{conf}</div>

      <GameSection label="Game 1" note="Winner → 7 seed">
        <GameCard
          homeId={result?.game1.home ?? seeds?.s7}
          awayId={result?.game1.away ?? seeds?.s8}
          homeSeed="7" awaySeed="8"
          homeScore={result?.game1.home_score}
          awayScore={result?.game1.away_score}
          winner={result?.game1.winner}
          simulated={simulated}
          seasonSimulated={seasonSimulated}
        />
      </GameSection>

      <GameSection label="Game 2" note="Loser eliminated">
        <GameCard
          homeId={result?.game2.home ?? seeds?.s9}
          awayId={result?.game2.away ?? seeds?.s10}
          homeSeed="9" awaySeed="10"
          homeScore={result?.game2.home_score}
          awayScore={result?.game2.away_score}
          winner={result?.game2.winner}
          simulated={simulated}
          seasonSimulated={seasonSimulated}
        />
      </GameSection>

      <GameSection label="Game 3" note="Winner → 8 seed">
        <GameCard
          homeId={result?.game3.home ?? null}
          awayId={result?.game3.away ?? null}
          homePlaceholder="Loser of G1"
          awayPlaceholder="Winner of G2"
          homeScore={result?.game3.home_score}
          awayScore={result?.game3.away_score}
          winner={result?.game3.winner}
          simulated={simulated}
          seasonSimulated={seasonSimulated}
        />
      </GameSection>

      {simulated && (
        <div className={styles.playoffResult}>
          <PlayoffSeedRow n={7} teamId={result.playoff_7} />
          <PlayoffSeedRow n={8} teamId={result.playoff_8} />
        </div>
      )}
    </div>
  )
}

function GameSection({ label, note, children }) {
  return (
    <div className={styles.gameSection}>
      <div className={styles.gameLabel}>
        <span className={styles.gameName}>{label}</span>
        <span className={styles.gameNote}>{note}</span>
      </div>
      {children}
    </div>
  )
}

function GameCard({ homeId, awayId, homeSeed, awaySeed, homePlaceholder, awayPlaceholder, homeScore, awayScore, winner, simulated, seasonSimulated }) {
  const homeWon = simulated && winner === homeId
  const awayWon = simulated && winner === awayId

  return (
    <div className={styles.card}>
      <TeamRow
        teamId={homeId}
        seed={homeSeed}
        placeholder={homePlaceholder ?? `Seed ${homeSeed}`}
        score={homeScore}
        won={homeWon}
        simulated={simulated}
        seasonSimulated={seasonSimulated}
      />
      <div className={styles.divider} />
      <TeamRow
        teamId={awayId}
        seed={awaySeed}
        placeholder={awayPlaceholder ?? `Seed ${awaySeed}`}
        score={awayScore}
        won={awayWon}
        simulated={simulated}
        seasonSimulated={seasonSimulated}
      />
    </div>
  )
}

function TeamRow({ teamId, seed, placeholder, score, won, simulated, seasonSimulated }) {
  if (!seasonSimulated || !teamId) {
    return (
      <div className={`${styles.row} ${styles.rowEmpty}`}>
        <div className={styles.logoGhost} />
        <span className={styles.placeholder}>{placeholder}</span>
        <span className={styles.dash}>—</span>
      </div>
    )
  }

  return (
    <div className={`${styles.row} ${simulated ? (won ? styles.rowWon : styles.rowLost) : ''}`}>
      <span className={styles.seedNum}>{seed ?? ''}</span>
      <img
        src={espnLogo(teamId)}
        alt={teamId}
        className={styles.logo}
        onError={e => { e.currentTarget.style.opacity = '0' }}
      />
      <span className={styles.abbr}>{teamId}</span>
      {simulated
        ? <span className={styles.score}>{score}</span>
        : <span className={styles.dash}>—</span>
      }
    </div>
  )
}

function PlayoffSeedRow({ n, teamId }) {
  return (
    <div className={styles.playoffSeedRow}>
      <span className={styles.playoffSeedNum}>{n}</span>
      <img
        src={espnLogo(teamId)}
        alt={teamId}
        className={styles.playoffLogo}
        onError={e => { e.currentTarget.style.opacity = '0' }}
      />
      <span className={styles.playoffTeam}>{teamId}</span>
      <span className={styles.playoffLabel}>Playoff Seed {n}</span>
    </div>
  )
}

