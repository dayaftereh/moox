import { type KeyboardEvent, type ReactNode, useEffect, useRef, useState } from 'react'

export type VisualSelectorOption<TId extends string = string> = {
  id: TId
  title: string
  details: readonly string[]
  visual: ReactNode
  availability?: 'supported' | 'planned'
  availabilityLabel?: string
}

type VisualSelectorProps<TId extends string> = {
  label: string
  options: readonly VisualSelectorOption<TId>[]
  selectedId: TId
  onChange: (id: TId) => void
  previousLabel: string
  nextLabel: string
  positionLabel: (current: number, total: number) => string
  infoLabel: (optionTitle: string) => string
  closeInfoLabel: string
}

export function VisualSelector<TId extends string>({
  label,
  options,
  selectedId,
  onChange,
  previousLabel,
  nextLabel,
  positionLabel,
  infoLabel,
  closeInfoLabel,
}: VisualSelectorProps<TId>) {
  const [infoOpen, setInfoOpen] = useState(false)
  const infoButtonRef = useRef<HTMLButtonElement>(null)
  const closeButtonRef = useRef<HTMLButtonElement>(null)
  const selectedIndex = Math.max(0, options.findIndex((option) => option.id === selectedId))
  const selected = options[selectedIndex]

  useEffect(() => {
    if (!infoOpen) return
    document.body.classList.add('modal-open')
    closeButtonRef.current?.focus()

    function onEscape(event: globalThis.KeyboardEvent) {
      if (event.key === 'Escape') closeInfo()
    }

    window.addEventListener('keydown', onEscape)
    return () => {
      document.body.classList.remove('modal-open')
      window.removeEventListener('keydown', onEscape)
    }
  }, [infoOpen])

  if (!selected) return null

  function step(delta: number) {
    const nextIndex = Math.min(options.length - 1, Math.max(0, selectedIndex + delta))
    if (nextIndex === selectedIndex) return
    onChange(options[nextIndex].id)
  }

  function closeInfo() {
    setInfoOpen(false)
    window.setTimeout(() => infoButtonRef.current?.focus(), 0)
  }

  function onKeyDown(event: KeyboardEvent<HTMLElement>) {
    if (infoOpen) return
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

      <div className="visual-selector-art" aria-live="polite">
        {selected.visual}
        <button
          ref={infoButtonRef}
          type="button"
          className="visual-selector-info-button"
          onClick={() => setInfoOpen(true)}
          aria-label={infoLabel(selected.title)}
        >
          ?
        </button>
      </div>

      <div className="visual-selector-controls">
        <button type="button" className="visual-selector-arrow" onClick={() => step(-1)} disabled={selectedIndex <= 0} aria-label={previousLabel}>&lt;</button>
        <div className="visual-selector-current">
          <h3>{selected.title}</h3>
        </div>
        <button type="button" className="visual-selector-arrow" onClick={() => step(1)} disabled={selectedIndex >= options.length - 1} aria-label={nextLabel}>&gt;</button>
      </div>

      <div className="visual-selector-position" aria-label={positionLabel(selectedIndex + 1, options.length)}>
        <div className="visual-selector-dots" aria-hidden="true">
          {options.map((option, index) => <i key={option.id} data-current={index === selectedIndex ? 'true' : 'false'} />)}
        </div>
        <span className="visual-selector-position-label">{positionLabel(selectedIndex + 1, options.length)}</span>
      </div>

      {infoOpen && (
        <div className="visual-selector-dialog-backdrop" onPointerDown={(event) => {
          if (event.target === event.currentTarget) closeInfo()
        }}>
          <section
            className="visual-selector-dialog"
            role="dialog"
            aria-modal="true"
            aria-labelledby={`visual-selector-dialog-${selected.id}`}
            onKeyDown={(event) => {
              if (event.key === 'Tab') {
                event.preventDefault()
                closeButtonRef.current?.focus()
              }
            }}
          >
            <div className="visual-selector-dialog-header">
              <div>
                <p className="eyebrow">{label}</p>
                <h3 id={`visual-selector-dialog-${selected.id}`}>{selected.title}</h3>
              </div>
              <button ref={closeButtonRef} type="button" className="visual-selector-dialog-close" onClick={closeInfo} aria-label={closeInfoLabel}>×</button>
            </div>
            {selected.availabilityLabel && <span className="visual-selector-dialog-status" data-availability={selected.availability ?? 'supported'}>{selected.availabilityLabel}</span>}
            <div className="visual-selector-dialog-copy">
              {selected.details.map((detail) => <p key={detail}>{detail}</p>)}
            </div>
          </section>
        </div>
      )}
    </section>
  )
}
