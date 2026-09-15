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

function assert(condition, message) {
  if (!condition) failures.push(message)
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
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
  const candidate = chromeCandidates().find((item) => fs.existsSync(item))
  if (!candidate) throw new Error(`Chrome/Chromium not found. Set CHROME_PATH. Checked: ${chromeCandidates().join(', ')}`)
  return candidate
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
  const endpoint = `http://127.0.0.1:${port}/json`
  for (let attempt = 0; attempt < 80; attempt++) {
    try {
      const targets = await fetch(endpoint).then((response) => response.json())
      const target = targets.find((candidate) => candidate.type === 'page')
      if (target?.webSocketDebuggerUrl) return target
    } catch {
      // Chrome is still starting.
    }
    await sleep(100)
  }
  throw new Error(`Chrome DevTools target did not appear on port ${port}`)
}

function terminateChrome() {
  if (!chrome?.pid) return
  if (process.platform === 'win32') {
    spawnSync('taskkill.exe', ['/PID', String(chrome.pid), '/T', '/F'], { stdio: 'ignore' })
  } else {
    try { chrome.kill('SIGTERM') } catch {}
  }
}

async function main() {
  const health = await fetch(`${baseURL}/healthz`)
  assert(health.ok, `MOX health endpoint failed at ${baseURL}/healthz`)
  if (!health.ok) throw new Error(`MOX server is not healthy at ${baseURL}`)

  const port = await freePort()
  const executable = findChrome()
  userDataDir = fs.mkdtempSync(path.join(os.tmpdir(), 'moox-selector-browser-'))
  chrome = spawn(executable, [
    '--headless=new',
    '--disable-gpu',
    '--no-first-run',
    '--no-sandbox',
    `--remote-debugging-port=${port}`,
    `--user-data-dir=${userDataDir}`,
    'about:blank',
  ], { stdio: 'ignore' })

  const target = await waitForTarget(port)
  ws = new WebSocket(target.webSocketDebuggerUrl)
  await new Promise((resolve, reject) => {
    ws.onopen = resolve
    ws.onerror = reject
  })

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
  const key = async (keyName, code, virtualKeyCode) => {
    await send('Input.dispatchKeyEvent', { type: 'keyDown', key: keyName, code, windowsVirtualKeyCode: virtualKeyCode, nativeVirtualKeyCode: virtualKeyCode })
    await send('Input.dispatchKeyEvent', { type: 'keyUp', key: keyName, code, windowsVirtualKeyCode: virtualKeyCode, nativeVirtualKeyCode: virtualKeyCode })
  }
  const touchTap = async (selector) => {
    const point = await evaluate(`(() => { const r = document.querySelector(${JSON.stringify(selector)})?.getBoundingClientRect(); return r ? { x: r.x + r.width / 2, y: r.y + r.height / 2 } : null })()`)
    if (!point) throw new Error(`touch target not found: ${selector}`)
    await send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [{ x: point.x, y: point.y, radiusX: 1, radiusY: 1, force: 1 }] })
    await send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] })
  }

  await send('Page.enable')
  await send('Runtime.enable')
  await send('Emulation.setDeviceMetricsOverride', { width: 390, height: 844, deviceScaleFactor: 1, mobile: true })
  await send('Page.navigate', { url: `${baseURL}/?gate3-browser-smoke=1#/new-game` })
  await sleep(800)
  await evaluate(`localStorage.setItem('moox.locale', 'de'); true`)
  await send('Page.reload', { ignoreCache: true })
  await sleep(850)

  const snapshotExpression = `(() => {
    const q = (selector) => document.querySelector(selector)
    const rect = (element) => { const r = element?.getBoundingClientRect(); return r ? { width: r.width, height: r.height } : null }
    const image = q('.new-game-galaxy-art img')
    const art = q('.visual-selector-art')
    const arrows = [...document.querySelectorAll('.visual-selector-arrow')].map(rect)
    return {
      language: document.documentElement.lang,
      selected: q('.visual-selector-current h3')?.textContent,
      source: image?.getAttribute('src'),
      imageReady: Boolean(image?.complete && image?.naturalWidth === 1200 && image?.naturalHeight === 675),
      art: rect(art),
      radius: art ? getComputedStyle(art).borderRadius : '',
      info: rect(q('.visual-selector-info-button')),
      arrows,
      scrollWidth: document.documentElement.scrollWidth,
      viewportWidth: document.documentElement.clientWidth,
      createDisabled: q('.new-game-form button[type=submit]')?.disabled,
      dotCount: document.querySelectorAll('.visual-selector-dots i').length,
      dialogOpen: Boolean(q('.visual-selector-dialog')),
    }
  })()`

  let snap = await evaluate(snapshotExpression)
  assert(snap.language === 'de', `expected German locale, got ${snap.language}`)
  assert(snap.selected === 'Klein', `expected initial Small/Klein, got ${snap.selected}`)
  assert(snap.source === '/assets/new-game/galaxy-size/small.svg', `unexpected initial art source ${snap.source}`)
  assert(snap.imageReady, 'Small artwork did not load at native 1200x675')
  assert(snap.scrollWidth === snap.viewportWidth && snap.viewportWidth === 390, `mobile horizontal overflow: ${snap.scrollWidth}/${snap.viewportWidth}`)
  assert(snap.info?.width >= 44 && snap.info?.height >= 44, `mobile info target below 44px: ${JSON.stringify(snap.info)}`)
  assert(snap.arrows.every((item) => item?.width >= 44 && item?.height >= 44), `mobile arrow target below 44px: ${JSON.stringify(snap.arrows)}`)
  assert(snap.radius === '16px', `expected 16px mobile art radius, got ${snap.radius}`)
  assert(snap.dotCount === 5, `expected five position dots, got ${snap.dotCount}`)
  assert(snap.createDisabled === false, 'Small should remain the currently supported Create Game option')

  await evaluate(`document.querySelectorAll('.visual-selector-arrow')[0].click(); true`)
  await sleep(100)
  snap = await evaluate(snapshotExpression)
  assert(snap.selected === 'Winzig', `previous arrow should select Tiny/Winzig, got ${snap.selected}`)
  assert(snap.source === '/assets/new-game/galaxy-size/tiny.svg', `Tiny art path wrong: ${snap.source}`)
  assert(snap.createDisabled === true, 'Tiny preview must keep Create Game disabled')

  await touchTap('.visual-selector-arrow:not(:disabled):last-of-type')
  await sleep(100)
  snap = await evaluate(snapshotExpression)
  assert(snap.selected === 'Klein', `touch tap on next arrow should select Small/Klein, got ${snap.selected}`)
  assert(snap.createDisabled === false, 'Touch navigation back to Small should re-enable Create Game')

  await evaluate(`document.querySelector('.visual-selector').focus(); true`)
  await key('ArrowLeft', 'ArrowLeft', 37)
  await sleep(80)
  snap = await evaluate(snapshotExpression)
  assert(snap.selected === 'Winzig', `ArrowLeft should select Tiny/Winzig, got ${snap.selected}`)
  await key('ArrowRight', 'ArrowRight', 39)
  await sleep(80)
  snap = await evaluate(snapshotExpression)
  assert(snap.selected === 'Klein', `ArrowRight should return to Small/Klein, got ${snap.selected}`)
  assert(snap.createDisabled === false, 'Keyboard navigation back to Small should re-enable Create Game')

  await key('End', 'End', 35)
  await sleep(80)
  snap = await evaluate(snapshotExpression)
  assert(snap.selected === 'Riesig', `End should select Huge/Riesig, got ${snap.selected}`)
  assert(snap.source === '/assets/new-game/galaxy-size/huge.svg', `Huge art path wrong: ${snap.source}`)
  assert(snap.createDisabled === true, 'Huge preview must keep Create Game disabled')

  await evaluate(`document.querySelector('.visual-selector-info-button').click(); true`)
  await sleep(80)
  const dialog = await evaluate(`(() => ({
    open: Boolean(document.querySelector('.visual-selector-dialog')),
    title: document.querySelector('.visual-selector-dialog h3')?.textContent,
    paragraphs: [...document.querySelectorAll('.visual-selector-dialog-copy p')].map((item) => item.textContent),
    activeClass: document.activeElement?.className,
    bodyModal: document.body.classList.contains('modal-open'),
  }))()`)
  assert(dialog.open, 'info dialog did not open')
  assert(dialog.title === 'Riesig', `info dialog title mismatch: ${dialog.title}`)
  assert(dialog.paragraphs.length >= 3, `expected detailed info paragraphs, got ${dialog.paragraphs.length}`)
  assert(dialog.activeClass === 'visual-selector-dialog-close', `dialog close control did not receive focus: ${dialog.activeClass}`)
  assert(dialog.bodyModal, 'modal-open body state missing while info dialog is open')

  await key('Escape', 'Escape', 27)
  await sleep(100)
  const closed = await evaluate(`({
    open: Boolean(document.querySelector('.visual-selector-dialog')),
    activeClass: document.activeElement?.className,
    bodyModal: document.body.classList.contains('modal-open'),
  })`)
  assert(!closed.open, 'Escape did not close info dialog')
  assert(closed.activeClass === 'visual-selector-info-button', `focus did not return to info control: ${closed.activeClass}`)
  assert(!closed.bodyModal, 'modal-open body state remained after close')

  await evaluate(`document.querySelector('.visual-selector').focus(); true`)
  await key('Home', 'Home', 36)
  await sleep(80)
  snap = await evaluate(snapshotExpression)
  assert(snap.selected === 'Winzig', `Home should select Tiny/Winzig, got ${snap.selected}`)

  const expected = [
    ['Winzig', 'tiny', true],
    ['Klein', 'small', false],
    ['Mittel', 'medium', true],
    ['Groß', 'large', true],
    ['Riesig', 'huge', true],
  ]
  for (let index = 0; index < expected.length; index++) {
    if (index > 0) {
      await evaluate(`document.querySelectorAll('.visual-selector-arrow')[1].click(); true`)
      await sleep(70)
    }
    snap = await evaluate(snapshotExpression)
    const [label, id, disabled] = expected[index]
    assert(snap.selected === label, `option ${index + 1} label mismatch: expected ${label}, got ${snap.selected}`)
    assert(snap.source === `/assets/new-game/galaxy-size/${id}.svg`, `option ${label} art path mismatch: ${snap.source}`)
    assert(snap.imageReady, `${label} artwork did not load at native 1200x675`)
    assert(snap.createDisabled === disabled, `${label} supported/planned create state mismatch`)
  }

  await send('Emulation.setDeviceMetricsOverride', { width: 1280, height: 900, deviceScaleFactor: 1, mobile: false })
  await sleep(120)
  snap = await evaluate(snapshotExpression)
  assert(snap.scrollWidth === snap.viewportWidth, `desktop horizontal overflow: ${snap.scrollWidth}/${snap.viewportWidth}`)
  assert(snap.art?.width >= 590 && snap.art?.width <= 620, `desktop artwork width outside frozen compact range: ${snap.art?.width}`)
  assert(snap.radius === '20px', `expected 20px desktop art radius, got ${snap.radius}`)
  assert(snap.arrows.every((item) => item?.width >= 52 && item?.height >= 52), `desktop arrow target below 52px: ${JSON.stringify(snap.arrows)}`)

  if (failures.length > 0) {
    console.error(`New Game selector browser smoke failed (${failures.length}):`)
    for (const failure of failures) console.error(`- ${failure}`)
    process.exitCode = 1
    return
  }

  console.log('New Game selector browser smoke passed: mobile + desktop layout, five options, mouse, touch, keyboard, info modal, focus return and supported/planned behavior.')
}

main().catch((error) => {
  console.error(error.stack || error.message || String(error))
  process.exitCode = 1
}).finally(async () => {
  try { ws?.close() } catch {}
  terminateChrome()
  await sleep(200)
  if (userDataDir) {
    try {
      fs.rmSync(userDataDir, { recursive: true, force: true, maxRetries: 20, retryDelay: 100 })
    } catch (cleanupError) {
      console.warn('Browser smoke cleanup warning: ' + cleanupError.message)
    }
  }
})
