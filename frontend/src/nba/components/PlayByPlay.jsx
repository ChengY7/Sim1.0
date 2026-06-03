import { useEffect, useRef } from 'react'
import styles from './PlayByPlay.module.css'

function evClass(type) {
  if (type === 'make_2pt') return styles.make
  if (type === 'make_3pt') return styles.make3
  if (type === 'turnover') return styles.turn
  if (type === 'foul')     return styles.foul
  return ''
}

function fmt(clockSec) {
  const m = Math.floor(clockSec / 60)
  const s = String(clockSec % 60).padStart(2, '0')
  return `${m}:${s}`
}

export default function PlayByPlay({ events }) {
  const listRef = useRef(null)

  useEffect(() => {
    if (listRef.current) listRef.current.scrollTop = listRef.current.scrollHeight
  }, [events])

  return (
    <div className={styles.wrap}>
      <div className={styles.header}>
        <span className={styles.title}>Play-by-Play</span>
        <span className={styles.count}>{events.length} plays</span>
      </div>
      <div className={styles.glass}>
        <div className={styles.list} ref={listRef}>
          {events.map(ev => (
            <div key={ev.possession} className={`${styles.ev} ${evClass(ev.type)}`}>
              <span className={styles.num}>P{String(ev.possession).padStart(3, '0')}</span>
              <span className={styles.clk}>{ev.period} {fmt(ev.clock_sec)}</span>
              <span className={`${styles.txt} ${styles[ev.team]}`}>{ev.text}</span>
              <span className={styles.scr}>{ev.home_score}–{ev.away_score}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
