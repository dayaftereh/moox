import { type KeyboardEvent, type ReactNode } from 'react'

export type VisualSelectorOption = {
  id: string
  title: string
  eyebrow?: string
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

  function onKeyDown(event: KeyboardEvent<HTMLDivElement>) {
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
    <section className="visual-selector" aria-label={label}>
      <div className="visual-selector-stage" tabIndex={0} onKeyDown={onKeyDown} aria-roledescription="carousel">
        <button type="button" className="visual-selector-arrow" onClick={() => step(-1)} disabled={selectedIndex <= 0} aria-label={previousLabel}>‹</button>
        <article className="visual-selector-card" data-availability={selected.availability ?? 'supported'} aria-live="polite">
          <div className="visual-selector-art">{selected.visual}</div>
          <div className="visual-selector-copy">
            {selected.eyebrow && <p className="eyebrow">{selected.eyebrow}</p>}
            <div className="visual-selector-title-row">
              <h3>{selected.title}</h3>
              {selected.availabilityLabel && <span className="visual-selector-availability">{selected.availabilityLabel}</span>}
            </div>
            <div className="visual-selector-facts">
              {selected.facts.map((fact) => <span key={fact}>{fact}</span>)}
            </div>
          </div>
        </article>
        <button type="button" className="visual-selector-arrow" onClick={() => step(1)} disabled={selectedIndex >= options.length - 1} aria-label={nextLabel}>›</button>
      </div>
      <div className="visual-selector-position" aria-label={positionLabel(selectedIndex + 1, options.length)}>
        <span>{positionLabel(selectedIndex + 1, options.length)}</span>
        <div className="visual-selector-dots" aria-hidden="true">
          {options.map((option, index) => <i key={option.id} data-current={index === selectedIndex ? 'true' : 'false'} />)}
        </div>
      </div>
    </section>
  )
}
