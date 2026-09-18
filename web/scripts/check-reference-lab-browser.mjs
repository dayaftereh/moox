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

  await fetchJSON(`${baseURL}/api/v1/games/${gameID}/turn-submissions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      schema_version: 1,
      game_id: gameID,
      seat_id: 1,
      turn: afterAdvance.view.turn,
      base_revision: afterAdvance.view.revision,
      commands: [{
        schema_version: 1,
        sequence: 1,
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
  const quote = queued.construction_buyouts?.find((item) => item.colony_id === colony.id)
  assert(Boolean(quote), 'Server did not project an authoritative construction buyout quote')
  assert((quote?.cost_bc ?? 0) > 0, `Buyout cost is not positive: ${quote?.cost_bc}`)
  assert(quote?.affordable === true, 'Buyout is not affordable after +10,000 BC')

  await navigate(`${baseURL}/#/game/${gameID}/colonies/${colony.id}`)
  await evaluate(`window.confirm = () => true; true`)
  await sleep(350)
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
  const buyCost = quote?.cost_bc ?? 0
  const clickedBuy = await evaluate(`(() => {
    const button = document.querySelector('.construction-buyout-actions button')
    if (!button || button.disabled) return false
    button.click()
    return true
  })()`)
  assert(clickedBuy, 'Normal buyout button was not clickable')
  await sleep(850)
  const bought = await fetchJSON(snapshotURL)
  assert(Math.abs(bought.view.empire.treasury.balance_bc - (beforeBuy - buyCost)) < 1e-6, `Buyout treasury delta incorrect: ${beforeBuy} - ${buyCost} -> ${bought.view.empire.treasury.balance_bc}`)
  const boughtProject = bought.view.colonies.find((item) => item.id === colony.id)?.construction
  assert(Math.abs((boughtProject?.progress_pp ?? -1) - (quote?.production_cost_pp ?? 0)) < 1e-6, 'Buyout did not fully fund authoritative PP')
  const boughtQuote = bought.construction_buyouts?.find((item) => item.colony_id === colony.id)
  assert((boughtQuote?.cost_bc ?? -1) === 0, `Bought project quote cost=${boughtQuote?.cost_bc}, want 0`)

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

  if (failures.length) throw new Error('Reference/buyout browser smoke failed:\n- ' + failures.join('\n- '))
  console.log('Reference/buyout browser smoke passed: More-only dev tools, +BC grant, +1/+5/N presence, normal server buyout with BC deduction, next-turn completion, 390px/desktop layout and ordinary-game isolation.')
}

main().catch((error) => {
  console.error(error)
  process.exitCode = 1
}).finally(cleanup)
