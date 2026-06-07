import { useState, useEffect, useRef } from 'react'
import { flagEmoji } from '../utils/flagEmoji'
import styles from './TeamPicker.module.css'

export default function TeamPicker({ teams, team1Id, onTeam1Change, team2Id, onTeam2Change }) {
  return (
    <div className={styles.matchup}>
      <TeamCard label="Team 1" teams={teams} value={team1Id} onChange={onTeam1Change} />
      <div className={styles.vsCol}>
        <div className={styles.vsLine} />
        <span className={styles.vsText}>VS</span>
        <div className={styles.vsLine} />
      </div>
      <TeamCard label="Team 2" teams={teams} value={team2Id} onChange={onTeam2Change} />
    </div>
  )
}

function TeamCard({ label, teams, value, onChange }) {
  const [open, setOpen]       = useState(false)
  const [focused, setFocused] = useState(0)
  const containerRef          = useRef(null)
  const listRef               = useRef(null)

  const sorted   = [...teams].sort((a, b) => a.name.localeCompare(b.name))
  const selected = sorted.find(t => t.id === value) ?? (value ? { id: value, name: value } : null)

  useEffect(() => {
    if (!open) return
    const idx = sorted.findIndex(t => t.id === value)
    setFocused(idx >= 0 ? idx : 0)
  }, [open])

  useEffect(() => {
    if (!open || !listRef.current) return
    listRef.current.children[focused]?.scrollIntoView({ block: 'nearest' })
  }, [focused, open])

  useEffect(() => {
    function onOutside(e) {
      if (!containerRef.current?.contains(e.target)) setOpen(false)
    }
    document.addEventListener('mousedown', onOutside)
    return () => document.removeEventListener('mousedown', onOutside)
  }, [])

  function onKeyDown(e) {
    if (!open) {
      if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown') { e.preventDefault(); setOpen(true) }
      return
    }
    if (e.key === 'Escape')    { setOpen(false); return }
    if (e.key === 'ArrowDown') { e.preventDefault(); setFocused(f => Math.min(f + 1, sorted.length - 1)) }
    if (e.key === 'ArrowUp')   { e.preventDefault(); setFocused(f => Math.max(f - 1, 0)) }
    if (e.key === 'Enter')     { onChange(sorted[focused].id); setOpen(false) }
  }

  return (
    <div className={`${styles.card} ${open ? styles.cardOpen : ''}`} ref={containerRef}>
      <div className={styles.label}>{label}</div>

      <button
        className={`${styles.trigger} ${open ? styles.triggerOpen : ''}`}
        onClick={() => teams.length && setOpen(o => !o)}
        onKeyDown={onKeyDown}
        aria-haspopup="listbox"
        aria-expanded={open}
      >
        {selected ? (
          <>
            <span className={styles.triggerFlag}>{flagEmoji(selected.id)}</span>
            <span className={styles.triggerName}>
              <span className={styles.triggerAbbr}>{selected.id}</span>
              <span className={styles.triggerFull}>{selected.name}</span>
            </span>
          </>
        ) : (
          <span className={styles.triggerPlaceholder}>Select team…</span>
        )}
        <ChevronIcon open={open} />
      </button>

      {open && (
        <ul className={styles.dropdown} role="listbox" ref={listRef}>
          {sorted.map((t, i) => (
            <li
              key={t.id}
              role="option"
              aria-selected={t.id === value}
              className={`${styles.option} ${t.id === value ? styles.optionSelected : ''} ${i === focused ? styles.optionFocused : ''}`}
              onMouseEnter={() => setFocused(i)}
              onMouseDown={() => { onChange(t.id); setOpen(false) }}
            >
              <span className={styles.optionFlag}>{flagEmoji(t.id)}</span>
              <span className={styles.optionAbbr}>{t.id}</span>
              <span className={styles.optionName}>{t.name}</span>
              {t.host && <span className={styles.optionHost}>HOST</span>}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

function ChevronIcon({ open }) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" width="14" height="14"
      className={`${styles.chevron} ${open ? styles.chevronOpen : ''}`}>
      <path d="M7 10l5 5 5-5z" />
    </svg>
  )
}
