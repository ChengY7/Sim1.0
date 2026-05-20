import { useState, useEffect, useRef } from 'react'
import styles from './TeamPicker.module.css'
import { espnLogo } from '../utils/espnLogo'

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
  const [open, setOpen] = useState(false)
  const [focused, setFocused] = useState(0)
  const containerRef = useRef(null)
  const listRef = useRef(null)

  const selected = teams.find(t => t.id === value)

  useEffect(() => {
    if (!open) return
    const idx = teams.findIndex(t => t.id === value)
    setFocused(idx >= 0 ? idx : 0)
  }, [open])

  useEffect(() => {
    if (!open || !listRef.current) return
    const item = listRef.current.children[focused]
    item?.scrollIntoView({ block: 'nearest' })
  }, [focused, open])

  useEffect(() => {
    function onClickOutside(e) {
      if (!containerRef.current?.contains(e.target)) setOpen(false)
    }
    document.addEventListener('mousedown', onClickOutside)
    return () => document.removeEventListener('mousedown', onClickOutside)
  }, [])

  function onKeyDown(e) {
    if (!open) {
      if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown') {
        e.preventDefault()
        setOpen(true)
      }
      return
    }
    if (e.key === 'Escape') { setOpen(false); return }
    if (e.key === 'ArrowDown') { e.preventDefault(); setFocused(f => Math.min(f + 1, teams.length - 1)) }
    if (e.key === 'ArrowUp')   { e.preventDefault(); setFocused(f => Math.max(f - 1, 0)) }
    if (e.key === 'Enter') { onChange(teams[focused].id); setOpen(false) }
  }

  return (
    <div className={`${styles.card} ${styles[side]} ${open ? styles.cardOpen : ''}`} ref={containerRef}>
      <div className={styles.label}>
        {side === 'home' ? <HomeIcon /> : <PlaneIcon />}
        {label}
      </div>

      <button
        className={`${styles.trigger} ${open ? styles.triggerOpen : ''} ${!teams.length ? styles.triggerDisabled : ''}`}
        onClick={() => teams.length && setOpen(o => !o)}
        onKeyDown={onKeyDown}
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-label={label}
        disabled={teams.length === 0}
      >
        {selected ? (
          <>
            <img
              src={espnLogo(selected.id)}
              alt={selected.id}
              className={styles.triggerLogo}
              onError={e => { e.currentTarget.style.opacity = '0' }}
            />
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
        <ul
          className={styles.dropdown}
          role="listbox"
          ref={listRef}
          aria-label={label}
        >
          {teams.map((t, i) => (
            <li
              key={t.id}
              role="option"
              aria-selected={t.id === value}
              className={`${styles.option} ${t.id === value ? styles.optionSelected : ''} ${i === focused ? styles.optionFocused : ''}`}
              onMouseEnter={() => setFocused(i)}
              onMouseDown={() => { onChange(t.id); setOpen(false) }}
            >
              <img
                src={espnLogo(t.id)}
                alt={t.id}
                className={styles.optionLogo}
                onError={e => { e.currentTarget.style.opacity = '0' }}
              />
              <span className={styles.optionAbbr}>{t.id}</span>
              <span className={styles.optionName}>{t.name}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

function ChevronIcon({ open }) {
  return (
    <svg
      viewBox="0 0 24 24" fill="currentColor" width="14" height="14"
      className={`${styles.chevron} ${open ? styles.chevronOpen : ''}`}
    >
      <path d="M7 10l5 5 5-5z" />
    </svg>
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
