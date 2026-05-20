import styles from './TeamPicker.module.css'

export default function TeamPicker({ teams, homeId, onHomeChange, awayId, onAwayChange }) {
  return (
    <div className={styles.matchup}>
      <TeamCard side="home" label="Home Team" teams={teams} value={homeId} onChange={onHomeChange} />
      <div className={styles.vsCol}>
        <div className={styles.vsLine} />
        <span className={styles.vsText}>VS</span>
        <div className={styles.vsLine} />
      </div>
      <TeamCard side="away" label="Away Team" teams={teams} value={awayId} onChange={onAwayChange} />
    </div>
  )
}

function TeamCard({ side, label, teams, value, onChange }) {
  return (
    <div className={`${styles.card} ${styles[side]}`}>
      <div className={styles.label}>
        {side === 'home' ? <HomeIcon /> : <PlaneIcon />}
        {label}
      </div>
      <div className={styles.selectWrap}>
        <select
          value={value}
          onChange={e => onChange(e.target.value)}
          aria-label={label}
          disabled={teams.length === 0}
        >
          {teams.map(t => (
            <option key={t.id} value={t.id}>{t.id} — {t.name}</option>
          ))}
        </select>
      </div>
    </div>
  )
}

function HomeIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" width="12" height="12">
      <path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z" />
    </svg>
  )
}

function PlaneIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" width="12" height="12">
      <path d="M21 16v-2l-8-5V3.5A1.5 1.5 0 0 0 11.5 2 1.5 1.5 0 0 0 10 3.5V9l-8 5v2l8-2.5V19l-2 1.5V22l3.5-1 3.5 1v-1.5L13 19v-5.5z" />
    </svg>
  )
}
