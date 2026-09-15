import { type KeyboardEvent, type ReactNode } from 'react'

export type VisualSelectorOption = {
  id: string
  title: string
  facts: readonly string[]
  visual: ReactNode
  availability?: 'supported' | 'planned'
  availabilityLabel?: string
}

type VisualSelectorProps = {
  label: string
  options: readonly VisualSelectorOption[]
  selectedId: string
  onChange: (id: string) => void
  previousLabel: string
  nextLabel: string
  positionLabel: (current: number, total: number) => string
}

export function VisualSelector({ label, options, selectedId, onChange, previousLabel, nextLabel, positionLabel }: VisualSelectorProps) {
  const selectedIndex = Math.max(0, options.findIndex((option) => option.id === selectedId))
  const selected = options[selectedIndex]
  if (!selected) return null

  function step(delta: number) {
    const nextIndex = Math.min(options.length - 1, Math.max(0, selectedIndex + delta))
    if (nextIndex === selectedIndex) return
    onChange(options[nextIndex].id)
  }

  function onKeyDown(event: KeyboardEvent<HTMLElement>) {
    if (event.key === 'ArrowLeft') {
      event.preventDefault()
      step(-1)
    } else if (event.key === 'ArrowRight') {
      event.preventDefault()
      step(1)
    } else if (event.key === 'Home') {
      event.preventDefault()
      onChange(options[0].id)
    } else if (event.key === 'End') {
      event.preventDefault()
      onChange(options[options.length - 1].id)
    }
  }

  return (
    <section
      className="visual-selector"
      data-availability={selected.availability ?? 'supported'}
      aria-label={label}
      aria-roledescription="carousel"
      tabIndex={0}
      onKeyDown={onKeyDown}
    >
      <h2 className="visual-selector-heading">{label}</h2>

      <div className="visual-selector-art" aria-live="polite">{selected.visual}</div>

      <div className="visual-selector-controls">
        <button type="button" className="visual-selector-arrow" onClick={() => step(-1)} disabled={selectedIndex <= 0} aria-label={previousLabel}>&lt;</button>
        <div className="visual-selector-current">
          <h3>{selected.title}</h3>
          {selected.availabilityLabel && <span className="visual-selector-availability">{selected.availabilityLabel}</span>}
        </div>
        <button type="button" className="visual-selector-arrow" onClick={() => step(1)} disabled={selectedIndex >= options.length - 1} aria-label={nextLabel}>&gt;</button>
      </div>

      <div className="visual-selector-facts">
        {selected.facts.map((fact) => <span key={fact}>{fact}</span>)}
      </div>

      <div className="visual-selector-position" aria-label={positionLabel(selectedIndex + 1, options.length)}>
        <div className="visual-selector-dots" aria-hidden="true">
          {options.map((option, index) => <i key={option.id} data-current={index === selectedIndex ? 'true' : 'false'} />)}
        </div>
        <span>{positionLabel(selectedIndex + 1, options.length)}</span>
      </div>
    </section>
  )
}
