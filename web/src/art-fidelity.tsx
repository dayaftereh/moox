import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BuildingArt } from './components/BuildingArt'
import { OrbitalBodyArt } from './components/OrbitalBodyArt'
import './art-fidelity.css'

const climates = [
  ['terran', 'Terran', 401],
  ['ocean', 'Ocean', 402],
  ['desert', 'Desert', 403],
  ['barren', 'Barren', 404],
  ['tundra', 'Tundra', 405],
  ['arctic', 'Arctic', 406],
  ['toxic', 'Toxic', 407],
  ['radiated', 'Radiated', 408],
  ['swamp', 'Swamp', 409],
] as const

const stations = [
  ['star_base', 'Star Base', 'Light orbital command and docking station'],
  ['battlestation', 'Battlestation', 'Armored combat station with heavy weapon pods'],
  ['star_fortress', 'Star Fortress', 'Massive layered fortress and command hub'],
] as const

const surfaceBuildings = [
  ['capitol', 'Capitol'],
  ['colony_base', 'Colony Base'],
  ['marine_barracks', 'Marine Barracks'],
] as const

function App() {
  return (
    <main className="art-lab">
      <header className="art-lab-hero">
        <div>
          <p className="art-lab-eyebrow">Slice 15.3 · Art Fidelity Pass</p>
          <h1>Planets & Building Illustrations</h1>
          <p>Actual runtime React/SVG components. No copied game art, no placeholder icon font. These same BuildingArt components are intended to be reusable later on a colony/planet surface.</p>
        </div>
        <a href="/?v=art-fidelity-1#/game/game-1/galaxy">Zurück zum Spiel</a>
      </header>

      <section className="art-lab-section">
        <div className="art-lab-section-title"><div><p>Orbital body fidelity</p><h2>Climate-driven procedural planets</h2></div><span>Gradient · Fractal Noise · Clouds · deterministic seed</span></div>
        <div className="planet-gallery">
          {climates.map(([climate, label, id]) => (
            <article className="planet-card" key={climate}>
              <div className="planet-card-art"><OrbitalBodyArt kind="planet" id={id} climateId={climate} /></div>
              <strong>{label}</strong>
              <small>{climate === 'terran' || climate === 'ocean' ? 'Cloud layer enabled' : 'Textured surface'}</small>
            </article>
          ))}
          <article className="planet-card"><div className="planet-card-art"><OrbitalBodyArt kind="gas_giant" id={511} /></div><strong>Gas Giant</strong><small>Bands + storm + warp noise</small></article>
          <article className="planet-card"><div className="planet-card-art"><OrbitalBodyArt kind="asteroid_belt" id={612} /></div><strong>Asteroid Belt</strong><small>Individual deterministic rocks</small></article>
        </div>
      </section>

      <section className="art-lab-section">
        <div className="art-lab-section-title"><div><p>Command stations</p><h2>Three different station images</h2></div><span>Not one generic Star Base icon</span></div>
        <div className="station-gallery">
          {stations.map(([id, title, detail]) => (
            <article className="station-card" key={id}>
              <BuildingArt buildingId={id} variant="hero" />
              <div><strong>{title}</strong><small>{detail}</small><code>{id}</code></div>
            </article>
          ))}
        </div>
      </section>

      <section className="art-lab-section">
        <div className="art-lab-section-title"><div><p>Planetary buildings</p><h2>Reusable building art</h2></div><span>Construction now · surface composition later</span></div>
        <div className="surface-building-gallery">
          {surfaceBuildings.map(([id, title]) => (
            <article className="surface-building-card" key={id}><BuildingArt buildingId={id} variant="surface" /><strong>{title}</strong><code>{id}</code></article>
          ))}
        </div>
      </section>

      <section className="art-lab-section art-lab-future">
        <div className="art-lab-section-title"><div><p>Next composition step</p><h2>Future colony surface</h2></div><span>Conceptual placement target</span></div>
        <div className="surface-concept">
          <div className="surface-concept-planet"><OrbitalBodyArt kind="planet" id={901} climateId="terran" /></div>
          <div className="surface-concept-buildings">
            {surfaceBuildings.map(([id, title]) => <div className="surface-concept-building" key={id}><BuildingArt buildingId={id} variant="surface" /><span>{title}</span></div>)}
          </div>
          <p>Der nächste Schritt kann dieselben Gebäude auf einer skalierbaren Planetensurface platzieren, ohne Construction-Art doppelt zu pflegen.</p>
        </div>
      </section>
    </main>
  )
}

createRoot(document.getElementById('root')!).render(<StrictMode><App /></StrictMode>)
