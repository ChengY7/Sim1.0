import styles from './Scoreboard.module.css'

export default function Scoreboard({ state, seed }) {
  const diff     = Math.abs(state.home_score - state.away_score)
  const winName  = state.home_score > state.away_score ? state.home_name : state.away_name
  const loseName = state.home_score > state.away_score ? state.away_name : state.home_name

  return (
    <>
      <div className={styles.scoreboard}>
        <div className={styles.period}>{state.period || 'FINAL'}</div>
        <div className={styles.row}>
          <div className={`${styles.team} ${styles.home}`}>
            <div className={styles.teamName}>{state.home_name}</div>
            <div className={styles.score}>{state.home_score}</div>
          </div>
          <div className={styles.dash}>—</div>
          <div className={`${styles.team} ${styles.away}`}>
            <div className={styles.teamName}>{state.away_name}</div>
            <div className={styles.score}>{state.away_score}</div>
          </div>
        </div>
        <div className={styles.statusRow}>
          <span className={styles.pill}>
            <span className={styles.dot} />
            Final
          </span>
        </div>
      </div>

      <div className={styles.winner}>
        <div className={styles.winnerLabel}>&#127942; Winner</div>
        <div className={styles.winnerName}>{winName}</div>
        <div className={styles.winnerMargin}>wins by {diff} over {loseName}</div>
      </div>

      <div className={styles.meta}>
        seed {seed} · {state.possession} possessions
      </div>
    </>
  )
}
