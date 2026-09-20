import fs from 'node:fs'
import net from 'node:net'
import os from 'node:os'
import path from 'node:path'
import { spawn, spawnSync } from 'node:child_process'

const baseURL = (process.argv[2] || process.env.MOOX_BASE_URL || 'http://127.0.0.1:7171').replace(/\/$/, '')
const failures = []
let chrome
let userDataDir
let ws

const assert = (condition, message) => { if (!condition) failures.push(message) }
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))

async function fetchJSON(url, init) {
  const response = await fetch(url, init)
  const text = await response.text()
  let payload = {}
  try { payload = text ? JSON.parse(text) : {} } catch {}
  if (!response.ok) throw new Error(`${response.status} ${url}: ${text}`)
  return payload
}

function chromeCandidates() {
  const candidates = []
  if (process.env.CHROME_PATH) candidates.push(process.env.CHROME_PATH)
  if (process.platform === 'win32') {
    for (const root of [process.env.PROGRAMFILES, process.env['PROGRAMFILES(X86)'], process.env.LOCALAPPDATA]) {
      if (root) candidates.push(path.join(root, 'Google', 'Chrome', 'Application', 'chrome.exe'))
    }
  } else if (process.platform === 'darwin') {
    candidates.push('/Applications/Google Chrome.app/Contents/MacOS/Google Chrome')
    candidates.push('/Applications/Chromium.app/Contents/MacOS/Chromium')
  } else {
    candidates.push('/usr/bin/google-chrome', '/usr/bin/google-chrome-stable', '/usr/bin/chromium', '/usr/bin/chromium-browser')
  }
  return candidates.filter((candidate, index, all) => candidate && all.indexOf(candidate) === index)
}

function findChrome() {
  const executable = chromeCandidates().find((item) => fs.existsSync(item))
  if (!executable) throw new Error(`Chrome/Chromium not found. Set CHROME_PATH. Checked: ${chromeCandidates().join(', ')}`)
  return executable
}

async function freePort() {
  return new Promise((resolve, reject) => {
    const server = net.createServer()
    server.unref()
    server.on('error', reject)
    server.listen(0, '127.0.0.1', () => {
      const address = server.address()
      const port = typeof address === 'object' && address ? address.port : 0
      server.close(() => resolve(port))
    })
  })
}

async function waitForTarget(port) {
  for (let attempt = 0; attempt < 80; attempt++) {
    try {
      const targets = await fetch(`http://127.0.0.1:${port}/json`).then((response) => response.json())
      const target = targets.find((candidate) => candidate.type === 'page')
      if (target?.webSocketDebuggerUrl) return target
    } catch {}
    await sleep(100)
  }
  throw new Error('Chrome DevTools target did not appear')
}

function cleanup() {
  if (ws) {
    try { ws.close() } catch {}
  }
  if (chrome?.pid) {
    if (process.platform === 'win32') spawnSync('taskkill.exe', ['/PID', String(chrome.pid), '/T', '/F'], { stdio: 'ignore' })
    else { try { chrome.kill('SIGTERM') } catch {} }
  }
  if (userDataDir) {
    try { fs.rmSync(userDataDir, { recursive: true, force: true }) } catch {}
  }
}

async function main() {
  const health = await fetch(`${baseURL}/healthz`)
  if (!health.ok) throw new Error(`MOOX server is not healthy at ${baseURL}`)

  const games = await fetchJSON(`${baseURL}/api/v1/games`)
  assert(games.some((game) => game.game_id === 'game-triangle-all-tech' && game.reference?.profile_id === 'all_tech'), 'All-Tech Triangle is not projected as a trusted reference game')
  assert(games.some((game) => game.game_id === 'game-triangle-mid-tech' && game.reference?.profile_id === 'mid_tech'), 'Mid-Tech Triangle is not projected as a trusted reference game')
  assert(games.some((game) => game.game_id === 'game-1' && !game.reference), 'Legacy game-1 unexpectedly exposes reference capability')

  const port = await freePort()
  userDataDir = fs.mkdtempSync(path.join(os.tmpdir(), 'moox-reference-lab-browser-'))
  chrome = spawn(findChrome(), [
    '--headless=new', '--disable-gpu', '--no-first-run', '--no-sandbox',
    `--remote-debugging-port=${port}`, `--user-data-dir=${userDataDir}`, 'about:blank',
  ], { stdio: 'ignore' })
  const target = await waitForTarget(port)
  ws = new WebSocket(target.webSocketDebuggerUrl)
  await new Promise((resolve, reject) => { ws.onopen = resolve; ws.onerror = reject })

  let requestID = 0
  const pending = new Map()
  ws.onmessage = (event) => {
    const message = JSON.parse(event.data)
    if (!message.id || !pending.has(message.id)) return
    const handlers = pending.get(message.id)
    pending.delete(message.id)
    if (message.error) handlers.reject(new Error(JSON.stringify(message.error)))
    else handlers.resolve(message.result)
  }
  const send = (method, params = {}) => new Promise((resolve, reject) => {
    const id = ++requestID
    pending.set(id, { resolve, reject })
    ws.send(JSON.stringify({ id, method, params }))
  })
  const evaluate = async (expression) => {
    const result = await send('Runtime.evaluate', { expression, returnByValue: true, awaitPromise: true })
    if (result.exceptionDetails) throw new Error(JSON.stringify(result.exceptionDetails))
    return result.result.value
  }
  const navigate = async (url) => {
    await send('Page.navigate', { url })
    await sleep(950)
  }

  const gameID = 'game-triangle-all-tech'
  const snapshotURL = `${baseURL}/api/v1/games/${gameID}/seats/1/snapshot`

  await send('Page.enable')
  await send('Runtime.enable')
  await send('Emulation.setDeviceMetricsOverride', { width: 390, height: 844, deviceScaleFactor: 1, mobile: true })
  await navigate(`${baseURL}/?reference-lab-browser=1#/game/${gameID}/more`)
  await evaluate(`localStorage.setItem('moox.locale', 'de'); true`)
  await send('Page.reload', { ignoreCache: true })
  await sleep(950)

  let state = await evaluate(`(() => {
    const panel = document.querySelector('.reference-lab')
    const buttons = panel ? [...panel.querySelectorAll('button')].map((button) => ({ text: button.textContent?.trim(), disabled: button.disabled })) : []
    const input = panel?.querySelector('input[type="number"]')
    return {
      hasPanel: Boolean(panel),
      text: panel?.textContent || '',
      buttons,
      max: input?.getAttribute('max') || '',
      overflow: document.documentElement.scrollWidth > window.innerWidth + 1,
    }
  })()`)
  assert(state.hasPanel, 'Development tools are missing on More at 390px')
  assert(state.text.includes('All-Tech'), 'All-Tech profile identity is missing')
  assert(state.text.includes('Entwicklungswerkzeuge'), 'German Development tools title is missing')
  assert(state.buttons.some((button) => button.text?.includes('1 Runde')), 'Advance 1 button is missing')
  assert(state.buttons.some((button) => button.text?.includes('5 Runden')), 'Advance 5 button is missing')
  assert(state.buttons.some((button) => button.text?.includes('+100 BC')), '+100 BC button is missing')
  assert(state.buttons.some((button) => button.text?.includes('+1.000 BC')), '+1.000 BC button is missing')
  assert(state.buttons.some((button) => button.text?.includes('+10.000 BC')), '+10.000 BC button is missing')
  assert(state.max === '25', `Advance N max=${state.max}, want 25`)
  assert(!state.overflow, 'Development tools cause horizontal overflow at 390px')

  const beforeGrant = await fetchJSON(snapshotURL)
  const clickedGrant = await evaluate(`(() => {
    const button = [...document.querySelectorAll('.reference-lab button')].find((item) => item.textContent?.includes('+10.000 BC'))
    if (!button || button.disabled) return false
    button.click()
    return true
  })()`)
  assert(clickedGrant, '+10.000 BC was not clickable')
  await sleep(850)
  const afterGrant = await fetchJSON(snapshotURL)
  assert(
    Math.abs(afterGrant.view.empire.treasury.balance_bc - (beforeGrant.view.empire.treasury.balance_bc + 10000)) < 1e-6,
    `BC grant moved treasury ${beforeGrant.view.empire.treasury.balance_bc} -> ${afterGrant.view.empire.treasury.balance_bc}`,
  )

  const beforeAdvance = afterGrant
  const clickedAdvance = await evaluate(`(() => {
    const button = [...document.querySelectorAll('.reference-lab button')].find((item) => item.textContent?.includes('1 Runde vorspulen'))
    if (!button || button.disabled) return false
    button.click()
    return true
  })()`)
  assert(clickedAdvance, 'Advance 1 Turn was not clickable')
  await sleep(1050)
  const afterAdvance = await fetchJSON(snapshotURL)
  assert(afterAdvance.view.turn === beforeAdvance.view.turn + 1, `Advance 1 moved ${beforeAdvance.view.turn} -> ${afterAdvance.view.turn}, expected +1`)

  const colony = afterAdvance.view.colonies[0]
  assert(Boolean(colony?.id), 'All-Tech control seat has no colony for buyout smoke')
  const constructionDecision = afterAdvance.decision?.decisions?.construction?.find((item) => item.colony_id === colony.id)
  const buildingChoice = constructionDecision?.choices?.find((choice) => choice.project_kind === 'building')
  assert(Boolean(buildingChoice), 'No legal building choice available for buyout smoke')
  if (!colony?.id || !buildingChoice) throw new Error('Cannot continue construction buyout smoke without colony/building choice')

  await fetchJSON(`${baseURL}/api/v1/games/${gameID}/seats/1/planning-draft`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      schema_version: 1,
      game_id: gameID,
      seat_id: 1,
      turn: afterAdvance.view.turn,
      base_revision: afterAdvance.view.revision,
      draft_revision: 1,
      orders: [{
        key: `construction:${colony.id}`,
        kind: 'colony.set_construction_queue',
        payload: {
          colony_id: colony.id,
          items: [{ project_kind: buildingChoice.project_kind, project_id: buildingChoice.project_id }],
        },
      }],
    }),
  })
  await sleep(250)
  const queued = await fetchJSON(snapshotURL)
  assert(queued.view.colonies.find((item) => item.id === colony.id)?.construction == null, 'Planning draft unexpectedly mutated authoritative construction before buyout')
  assert(queued.planning_draft?.orders?.some((order) => order.key === `construction:${colony.id}`), 'Server did not retain drafted construction before same-turn buyout')
  const costPP = buildingChoice.production_cost_pp ?? buildingChoice.cost_pp ?? 0

  await navigate(`${baseURL}/#/game/${gameID}/colonies/${colony.id}`)
  await evaluate(`window.confirm = () => true; true`)
  for (let attempt = 0; attempt < 20; attempt++) {
    const ready = await evaluate(`Boolean(document.querySelector('.construction-buyout-actions button'))`)
    if (ready) break
    await sleep(150)
  }
  const buyUI = await evaluate(`(() => {
    const button = document.querySelector('.construction-buyout-actions button')
    return {
      exists: Boolean(button),
      text: button?.textContent?.trim() || '',
      disabled: button?.disabled ?? true,
      overflow: document.documentElement.scrollWidth > window.innerWidth + 1,
    }
  })()`)
  assert(buyUI.exists, 'Normal colony buyout button is missing')
  assert(buyUI.text.includes('Kaufen'), `Buyout button text=${buyUI.text}`)
  assert(!buyUI.disabled, 'Affordable buyout button is disabled')
  assert(!buyUI.overflow, 'Colony buyout UI causes horizontal overflow at 390px')

  const beforeBuy = queued.view.empire.treasury.balance_bc
  const buyCostMatch = buyUI.text.match(/([0-9]+(?:\.[0-9]+)?)\s*BC/)
  const buyCost = Number(buyCostMatch?.[1] ?? 0)
  assert(buyCost > 0, `Displayed draft buyout cost is not positive: ${buyUI.text}`)
  const clickedBuy = await evaluate(`(() => {
    const button = document.querySelector('.construction-buyout-actions button')
    if (!button || button.disabled) return false
    button.click()
    return true
  })()`)
  assert(clickedBuy, 'Normal buyout button was not clickable')
  await sleep(850)
  const bought = await fetchJSON(snapshotURL)
  assert(Math.abs(bought.view.empire.treasury.balance_bc - (beforeBuy - buyCost)) <= 0.06, `Buyout treasury delta differs from displayed 0.1-BC price: ${beforeBuy} - ${buyCost} -> ${bought.view.empire.treasury.balance_bc}`)
  const boughtProject = bought.view.colonies.find((item) => item.id === colony.id)?.construction
  assert(Math.abs((boughtProject?.progress_pp ?? -1) - costPP) < 1e-6, 'Buyout did not fully fund authoritative PP')
  const boughtQuote = bought.construction_buyouts?.find((item) => item.colony_id === colony.id)
  assert((boughtQuote?.cost_bc ?? -1) === 0, `Bought project quote cost=${boughtQuote?.cost_bc}, want 0`)
  assert(!bought.planning_draft?.orders?.some((order) => order.key === `construction:${colony.id}`), 'Construction planning draft survived same-turn buyout')

  await navigate(`${baseURL}/#/game/${gameID}/more`)
  const clickedFinish = await evaluate(`(() => {
    const button = [...document.querySelectorAll('.reference-lab button')].find((item) => item.textContent?.includes('1 Runde vorspulen'))
    if (!button || button.disabled) return false
    button.click()
    return true
  })()`)
  assert(clickedFinish, 'Advance 1 after buyout was not clickable')
  await sleep(1050)
  const completed = await fetchJSON(snapshotURL)
  const completedColony = completed.view.colonies.find((item) => item.id === colony.id)
  assert(completedColony?.buildings?.includes(buildingChoice.project_id), `Bought project ${buildingChoice.project_id} did not complete on next normal turn`)

  // Slice 17.1 Ship Designer smoke: compact ship-info rows, five ? details, effect-centric popup selection, explicit no-Shield and All-Tech Doom Star save.
  await navigate(`${baseURL}/#/game/${gameID}/shipbuilder`)
  await sleep(250)
  await evaluate(`(() => { const button = document.querySelector('.shipdesigner-new'); if (button) button.click(); return true })()`)
  await sleep(150)
  let designerUI = await evaluate(`(() => {
    const infoCard = document.querySelector('.shipdesigner-ship-info-card')
    const rows = [...document.querySelectorAll('.shipdesigner-system-row')].map((row) => ({
      system: row.getAttribute('data-system') || '',
      text: row.textContent || '',
      help: Boolean(row.querySelector('.shipdesigner-system-help')),
    }))
    return {
      infoCard: Boolean(infoCard),
      rows,
      helpCount: document.querySelectorAll('.shipdesigner-system-help').length,
      summaryCount: document.querySelectorAll('.shipdesigner-system-summary-grid > div').length,
      dropdowns: document.querySelectorAll('.shipdesigner-ship-info-card select').length,
      hullText: document.querySelector('.shipdesigner-hull-current strong')?.textContent || '',
      previewHeight: document.querySelector('.shipdesigner-preview')?.getBoundingClientRect().height || 0,
      hullSelectorHeight: document.querySelector('.shipdesigner-hull-current')?.getBoundingClientRect().height || 0,
      overflow: document.documentElement.scrollWidth > window.innerWidth + 1,
    }
  })()`)
  assert(designerUI.infoCard, 'Ship information card is missing below the ship preview')
  assert(designerUI.rows.length === 5, `Ship information has ${designerUI.rows.length} mandatory-system rows, want 5`)
  assert(designerUI.helpCount === 5 && designerUI.rows.every((row) => row.help), `Ship information help affordances=${designerUI.helpCount}, want one for every system`)
  assert(designerUI.summaryCount === 4, `Ship information summary stat count=${designerUI.summaryCount}, want 4`)
  assert(designerUI.dropdowns === 0, `Ship information still renders ${designerUI.dropdowns} dropdown(s)`)
  assert(!designerUI.rows.some((row) => row.text.includes('Strategische FTL') || row.text.includes('Strahl-Angriff') || row.text.includes('Panzerungs-Haltbarkeit') || row.text.includes('Blockt 10 Schaden') || row.text.includes('Unbegrenzte strategische Reichweite')), 'Compact ship information still exposes verbose effect copy instead of delegating it to ? details')
  assert(!designerUI.rows.some((row) => /\bM\d+\b/.test(row.text) || row.text.includes('Miniaturisierung')), 'Compact ship information exposes technical miniaturization notation')
  assert(designerUI.previewHeight <= 185, `Mobile Ship Designer preview height=${designerUI.previewHeight}, want <=185`)
  assert(designerUI.hullSelectorHeight <= 52, `Mobile Ship Designer hull selector height=${designerUI.hullSelectorHeight}, want <=52`)
  assert(!designerUI.overflow, 'Ship information card causes horizontal overflow at 390px')

  const openedComputer = await evaluate(`(() => { const button = document.querySelector('[data-system="computer"] .shipdesigner-system-select-trigger'); if (!button) return false; button.click(); return true })()`)
  assert(openedComputer, 'Computer row did not open selection popup')
  await sleep(120)
  const computerPopup = await evaluate(`(() => ({
    open: Boolean(document.querySelector('.shipdesigner-system-dialog')),
    choices: [...document.querySelectorAll('.shipdesigner-computer-dialog .shipdesigner-choice-card')].map((button) => ({ id: button.getAttribute('data-component-id') || '', disabled: button.disabled, text: button.textContent || '' })),
    overflow: document.documentElement.scrollWidth > window.innerWidth + 1,
  }))()`)
  assert(computerPopup.open, 'Computer selection popup is missing')
  assert(computerPopup.choices.some((choice) => choice.id === 'electronic_computer' && !choice.disabled && choice.text.includes('Strahl-Angriff +25') && choice.text.includes('-75 % Baukosten')), 'Electronic Computer popup card/effect/player-facing reduction missing')
  assert(!computerPopup.choices.some((choice) => /\bM\d+\b/.test(choice.text) || choice.text.includes('Miniaturisierung')), 'Computer popup still exposes technical M-level miniaturization notation')
  assert(computerPopup.choices.some((choice) => choice.id === 'moleculartronic_computer' && !choice.disabled && choice.text.includes('Strahl-Angriff +125')), 'Moleculartronic Computer popup card/effect missing')
  assert(!computerPopup.overflow, 'Computer selection popup causes horizontal overflow at 390px')
  await evaluate(`(() => { document.querySelector('.shipdesigner-system-dialog-close')?.click(); return true })()`)
  await sleep(100)

  const openedFuelInfo = await evaluate(`(() => { const button = document.querySelector('[data-system="fuel"] .shipdesigner-system-help'); if (!button) return false; button.click(); return true })()`)
  assert(openedFuelInfo, 'Fuel ? did not open detail popup')
  await sleep(100)
  const fuelInfoText = await evaluate(`document.querySelector('.shipdesigner-system-dialog')?.textContent || ''`)
  assert(fuelInfoText.includes('Unbegrenzte strategische Reichweite'), 'Fuel detail popup does not explain the installed effect')
  await evaluate(`(() => { document.querySelector('.shipdesigner-system-dialog-close')?.click(); return true })()`)
  await sleep(100)
  const measureShipVisualSize = () => evaluate(`(() => {
    const ship = document.querySelector('.shipdesigner-preview-ship')
    if (!ship) return null
    const rects = [...ship.querySelectorAll('path, circle, ellipse')].map((node) => node.getBoundingClientRect()).filter((rect) => rect.width > 0 && rect.height > 0)
    if (!rects.length) return null
    const left = Math.min(...rects.map((rect) => rect.left))
    const right = Math.max(...rects.map((rect) => rect.right))
    const top = Math.min(...rects.map((rect) => rect.top))
    const bottom = Math.max(...rects.map((rect) => rect.bottom))
    return {
      hull: document.querySelector('.shipdesigner-hull-current strong')?.textContent || '',
      maxSize: Math.max(right - left, bottom - top),
    }
  })()`)
  const hullVisualSizes = [await measureShipVisualSize()]
  for (let step = 0; step < 5; step++) {
    const advanced = await evaluate(`(() => {
      const buttons = [...document.querySelectorAll('.shipdesigner-hull-stepper button')]
      const button = buttons[buttons.length - 1]
      if (!button || button.disabled) return false
      button.click()
      return true
    })()`)
    assert(advanced, `Could not advance Ship Designer hull at step ${step + 1}`)
    await sleep(120)
    hullVisualSizes.push(await measureShipVisualSize())
  }
  assert(hullVisualSizes.length === 6 && hullVisualSizes.every(Boolean), 'Could not measure all six Ship Designer hull visuals')
  for (let index = 1; index < hullVisualSizes.length; index++) {
    const previous = hullVisualSizes[index - 1]
    const current = hullVisualSizes[index]
    assert(current.maxSize > previous.maxSize + 2, `Ship visual size did not grow ${previous.hull}=${previous.maxSize.toFixed(1)} -> ${current.hull}=${current.maxSize.toFixed(1)}`)
  }
  designerUI = await evaluate(`(() => ({
    hullText: document.querySelector('.shipdesigner-hull-current strong')?.textContent || '',
    saveDisabled: document.querySelector('.shipdesigner-save-row .button-primary')?.disabled ?? true,
  }))()`)
  assert(designerUI.hullText.includes('Doom'), `All-Tech hull stepper ended at ${designerUI.hullText}, want Doom Star`)
  const doomPreview = await evaluate(`(() => {
    const preview = document.querySelector('.shipdesigner-preview')?.getBoundingClientRect()
    const ship = document.querySelector('.shipdesigner-preview-ship')?.getBoundingClientRect()
    if (!preview || !ship) return null
    return {
      previewWidth: preview.width,
      previewHeight: preview.height,
      shipWidth: ship.width,
      shipHeight: ship.height,
      centeredX: Math.abs((ship.left + ship.width / 2) - (preview.left + preview.width / 2)),
      centeredY: Math.abs((ship.top + ship.height / 2) - (preview.top + preview.height / 2)),
      contained: ship.left >= preview.left - 1 && ship.right <= preview.right + 1 && ship.top >= preview.top - 1 && ship.bottom <= preview.bottom + 1,
    }
  })()`)
  assert(Boolean(doomPreview), 'Doom Star preview geometry is unavailable')
  assert(doomPreview.contained, `Mobile Doom Star SVG escapes preview bounds: ship=${doomPreview.shipWidth}x${doomPreview.shipHeight}, preview=${doomPreview.previewWidth}x${doomPreview.previewHeight}`)
  assert(doomPreview.centeredX <= 2 && doomPreview.centeredY <= 2, `Mobile Doom Star SVG is not centered: dx=${doomPreview.centeredX}, dy=${doomPreview.centeredY}`)

  const openedComputerForEdit = await evaluate(`(() => { const button = document.querySelector('[data-system="computer"] .shipdesigner-system-select-trigger'); if (!button) return false; button.click(); return true })()`)
  assert(openedComputerForEdit, 'Could not open Computer selector for Doom Star edit')
  await sleep(100)
  const selectedComputerForEdit = await evaluate(`(() => { const button = document.querySelector('.shipdesigner-computer-dialog [data-component-id="electronic_computer"]'); if (!button || button.disabled) return false; button.click(); return true })()`)
  assert(selectedComputerForEdit, 'Could not select Electronic Computer from popup')
  await sleep(100)
  const openedShieldForEdit = await evaluate(`(() => { const button = document.querySelector('[data-system="shield"] .shipdesigner-system-select-trigger'); if (!button) return false; button.click(); return true })()`)
  assert(openedShieldForEdit, 'Could not open Shield selector for Doom Star edit')
  await sleep(100)
  const selectedNoShield = await evaluate(`(() => { const button = document.querySelector('.shipdesigner-shield-dialog [data-shield-id=""]'); if (!button || button.disabled) return false; button.click(); return true })()`)
  assert(selectedNoShield, 'Could not select explicit no-Shield from popup')
  const edited = await evaluate(`(() => {
    const input = document.querySelector('#ship-design-name')
    if (!input) return false
    const inputSetter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')?.set
    inputSetter?.call(input, 'G3 Doom Star')
    input.dispatchEvent(new Event('input', { bubbles: true }))
    return true
  })()`)
  assert(edited, 'Could not edit All-Tech Doom Star Computer/Shield/name controls')
  await sleep(250)
  const clickedSave = await evaluate(`(() => {
    const button = document.querySelector('.shipdesigner-save-row .button-primary')
    if (!button || button.disabled) return false
    button.click()
    return true
  })()`)
  assert(clickedSave, 'All-Tech Doom Star save button was not clickable')
  await sleep(1200)
  const afterDesignerSave = await fetchJSON(snapshotURL)
  const savedDoom = afterDesignerSave.decision?.strategic?.ship_designs?.find((design) => design.name === 'G3 Doom Star')
  assert(Boolean(savedDoom), 'All-Tech Doom Star design was not saved')
  assert(savedDoom?.spec?.hull_id === 'doom_star', `Saved G3 hull=${savedDoom?.spec?.hull_id}`)
  assert(savedDoom?.spec?.computer_id === 'electronic_computer', `Saved G3 computer=${savedDoom?.spec?.computer_id}`)
  assert(!savedDoom?.spec?.shield_id, `Saved G3 shield=${savedDoom?.spec?.shield_id}, want none`)
  assert((savedDoom?.spec?.production_cost_pp ?? 0) > 0, 'Saved G3 Doom Star has no authoritative production cost')

  await send('Emulation.setDeviceMetricsOverride', { width: 1280, height: 900, deviceScaleFactor: 1, mobile: false })
  await sleep(250)
  const openedDesktopComputer = await evaluate(`(() => { const button = document.querySelector('[data-system="computer"] .shipdesigner-system-select-trigger'); if (!button) return false; button.click(); return true })()`)
  assert(openedDesktopComputer, 'Desktop Computer selector did not open')
  await sleep(100)
  const designerDesktop = await evaluate(`(() => {
    const info = document.querySelector('.shipdesigner-ship-info-card')?.getBoundingClientRect()
    const cards = [...document.querySelectorAll('.shipdesigner-computer-dialog .shipdesigner-choice-card')].slice(0, 2).map((card) => card.getBoundingClientRect())
    return {
      infoWidth: info?.width || 0,
      popupTwoColumns: cards.length === 2 && Math.abs(cards[0].top - cards[1].top) < 8 && cards[0].left < cards[1].left,
      choiceWidth: cards[0]?.width || 0,
      previewHeight: document.querySelector('.shipdesigner-preview')?.getBoundingClientRect().height || 0,
      hullSelectorHeight: document.querySelector('.shipdesigner-hull-current')?.getBoundingClientRect().height || 0,
      overflow: document.documentElement.scrollWidth > window.innerWidth + 1,
    }
  })()`)
  assert(designerDesktop.infoWidth > 500, `Desktop ship information card width=${designerDesktop.infoWidth}, want >500`)
  assert(designerDesktop.popupTwoColumns, 'Desktop Computer popup does not render choice cards in two columns')
  assert(designerDesktop.choiceWidth > 250, `Desktop Computer popup choice width=${designerDesktop.choiceWidth}, want >250`)
  assert(designerDesktop.previewHeight <= 205, `Desktop Ship Designer preview height=${designerDesktop.previewHeight}, want <=205`)
  assert(designerDesktop.hullSelectorHeight <= 52, `Desktop Ship Designer hull selector height=${designerDesktop.hullSelectorHeight}, want <=52`)
  assert(!designerDesktop.overflow, 'Desktop Ship Designer information/popup causes horizontal overflow')
  await evaluate(`(() => { document.querySelector('.shipdesigner-system-dialog-close')?.click(); return true })()`)
  await send('Emulation.setDeviceMetricsOverride', { width: 390, height: 844, deviceScaleFactor: 1, mobile: true })
  await sleep(150)
  await navigate(`${baseURL}/#/game/${gameID}/more`)
  await sleep(250)
  await send('Emulation.setDeviceMetricsOverride', { width: 1280, height: 900, deviceScaleFactor: 1, mobile: false })
  await sleep(250)
  const desktop = await evaluate(`(() => {
    const panel = document.querySelector('.reference-lab')
    return { hasPanel: Boolean(panel), width: panel?.getBoundingClientRect().width || 0, overflow: document.documentElement.scrollWidth > window.innerWidth + 1 }
  })()`)
  assert(desktop.hasPanel && desktop.width > 500, 'Development tools did not adapt to desktop width')
  assert(!desktop.overflow, 'Development tools cause horizontal overflow on desktop')

  await navigate(`${baseURL}/#/game/game-1/more`)
  const ordinary = await evaluate(`Boolean(document.querySelector('.reference-lab'))`)
  assert(!ordinary, 'Ordinary game-1 renders development credit/turn controls')

  if (failures.length) throw new Error('Reference/buyout/ship-designer browser smoke failed:\n- ' + failures.join('\n- '))
  console.log('Reference/buyout/ship-designer browser smoke passed: six-hull All-Tech Designer, compact ship-information rows, ? detail popups and player-facing miniaturization reductions, Computer/Shield card selectors, explicit no-Shield save, mobile/desktop Designer layout, More-only dev tools, +BC grant, +1/+5/N presence, same-turn drafted construction buyout with BC deduction, next-turn completion, 390px/desktop layout and ordinary-game isolation.')
}

main().catch((error) => {
  console.error(error)
  process.exitCode = 1
}).finally(cleanup)
