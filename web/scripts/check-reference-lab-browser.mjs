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

  const games = await fetch(`${baseURL}/api/v1/games`).then((response) => response.json())
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

  await send('Page.enable')
  await send('Runtime.enable')
  await send('Emulation.setDeviceMetricsOverride', { width: 390, height: 844, deviceScaleFactor: 1, mobile: true })
  await navigate(`${baseURL}/?reference-lab-browser=1#/game/game-triangle-all-tech/galaxy`)
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
  assert(state.hasPanel, 'All-Tech reference panel is missing at 390px')
  assert(state.text.includes('All-Tech'), 'All-Tech profile identity is missing')
  assert(state.text.includes('Referenzlabor'), 'German Reference Lab title is missing')
  assert(state.buttons.length >= 2, `expected Advance 1 + Advance N buttons, got ${state.buttons.length}`)
  assert(state.max === '25', `Advance N max=${state.max}, want 25`)
  assert(!state.overflow, 'Reference Lab causes horizontal overflow at 390px')

  const before = await fetch(`${baseURL}/api/v1/games/game-triangle-all-tech/seats/1/snapshot`).then((response) => response.json())
  const clicked = await evaluate(`(() => {
    const button = document.querySelector('.reference-lab__controls button')
    if (!button || button.disabled) return false
    button.click()
    return true
  })()`)
  assert(clicked, 'Advance 1 Turn button was not clickable')
  await sleep(1250)
  const after = await fetch(`${baseURL}/api/v1/games/game-triangle-all-tech/seats/1/snapshot`).then((response) => response.json())
  assert(after.view.turn === before.view.turn + 1, `Advance 1 Turn moved ${before.view.turn} -> ${after.view.turn}, expected +1`)
  state = await evaluate(`(() => ({ text: document.querySelector('.reference-lab')?.textContent || '' }))()`)
  assert(state.text.includes('Runde'), 'Reference result feedback is not visible after advance')

  await send('Emulation.setDeviceMetricsOverride', { width: 1280, height: 900, deviceScaleFactor: 1, mobile: false })
  await sleep(250)
  const desktop = await evaluate(`(() => {
    const panel = document.querySelector('.reference-lab')
    return { hasPanel: Boolean(panel), width: panel?.getBoundingClientRect().width || 0, overflow: document.documentElement.scrollWidth > window.innerWidth + 1 }
  })()`)
  assert(desktop.hasPanel && desktop.width > 500, 'Reference Lab did not adapt to desktop width')
  assert(!desktop.overflow, 'Reference Lab causes horizontal overflow on desktop')

  await navigate(`${baseURL}/#/game/game-1/galaxy`)
  const ordinary = await evaluate(`Boolean(document.querySelector('.reference-lab'))`)
  assert(!ordinary, 'Ordinary game-1 renders Reference Lab controls')

  if (failures.length) throw new Error('Reference Lab browser smoke failed:\n- ' + failures.join('\n- '))
  console.log('Reference Lab browser smoke passed: trusted All-Tech/Mid-Tech projection, 390px + desktop layout, Advance 1 normal turn, ordinary-game isolation.')
}

main().catch((error) => {
  console.error(error)
  process.exitCode = 1
}).finally(cleanup)
