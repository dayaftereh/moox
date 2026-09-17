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
    const point = await evaluate(`(() => { const element = document.querySelector(${JSON.stringify(selector)}); if (!element) return null; element.scrollIntoView({ block: 'center', inline: 'nearest' }); const r = element.getBoundingClientRect(); return { x: r.x + r.width / 2, y: r.y + r.height / 2 } })()`)
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
    const root = q('.visual-selector[data-setting-id=\"galaxy-size\"]')
    const image = root?.querySelector('.new-game-galaxy-art img')
    const art = root?.querySelector('.visual-selector-art')
    const arrows = [...(root?.querySelectorAll('.visual-selector-arrow') ?? [])].map(rect)
    return {
      language: document.documentElement.lang,
      selected: root?.querySelector('.visual-selector-current h3')?.textContent,
      source: image?.getAttribute('src'),
      imageReady: Boolean(image?.complete && image?.naturalWidth === 1200 && image?.naturalHeight === 675),
      art: rect(art),
      radius: art ? getComputedStyle(art).borderRadius : '',
      info: rect(root?.querySelector('.visual-selector-info-button')),
      arrows,
      scrollWidth: document.documentElement.scrollWidth,
      viewportWidth: document.documentElement.clientWidth,
      createDisabled: q('.new-game-form button[type=submit]')?.disabled,
      dotCount: root?.querySelectorAll('.visual-selector-dots i').length ?? 0,
      availability: root?.getAttribute('data-availability'),
      dialogOpen: Boolean(q('.visual-selector-dialog')),
    }
  })()`

  let snap = await evaluate(snapshotExpression)
  assert(snap.language === 'de', `expected German locale, got ${snap.language}`)
  assert(snap.selected === 'Klein', `expected initial Small/Klein, got ${snap.selected}`)
  assert(snap.source === '/assets/new-game/galaxy-size/small.svg', `unexpected initial art source ${snap.source}`)
  assert(snap.imageReady, 'Small artwork did not load at native 1200x675')
  assert(snap.scrollWidth === snap.viewportWidth && snap.viewportWidth === 390, `mobile horizontal overflow: ${snap.scrollWidth}/${snap.viewportWidth}`)
  assert(snap.info?.width >= 44 && snap.info?.height >= 44, `mobile Size info target below 44px: ${JSON.stringify(snap.info)}`)
  assert(snap.arrows.every((item) => item?.width >= 44 && item?.height >= 44), `mobile Size arrow target below 44px: ${JSON.stringify(snap.arrows)}`)
  assert(snap.radius === '16px', `expected 16px mobile Size art radius, got ${snap.radius}`)
  assert(snap.dotCount === 4, `expected four Size position dots, got ${snap.dotCount}`)
  assert(snap.availability === 'supported', `Small availability=${snap.availability}`)
  assert(snap.createDisabled === false, 'Authoritative Small+Normal defaults must allow Create Game')

  await touchTap('.visual-selector[data-setting-id="galaxy-size"] .visual-selector-arrow:not(:disabled):last-of-type')
  await sleep(100)
  snap = await evaluate(snapshotExpression)
  assert(snap.selected === 'Mittel', `touch next should select Medium/Mittel, got ${snap.selected}`)
  assert(snap.source === '/assets/new-game/galaxy-size/medium.svg', `Medium art path wrong: ${snap.source}`)
  assert(snap.createDisabled === false, 'Medium is authoritative and must keep Create Game enabled')

  await evaluate(`document.querySelector('.visual-selector[data-setting-id="galaxy-size"]')?.focus(); true`)
  await key('End', 'End', 35)
  await sleep(80)
  snap = await evaluate(snapshotExpression)
  assert(snap.selected === 'Riesig', `End should select Huge/Riesig, got ${snap.selected}`)
  assert(snap.source === '/assets/new-game/galaxy-size/huge.svg', `Huge art path wrong: ${snap.source}`)
  assert(snap.createDisabled === false, 'Huge is authoritative and must keep Create Game enabled')

  await evaluate(`document.querySelector('.visual-selector[data-setting-id="galaxy-size"] .visual-selector-info-button')?.click(); true`)
  await sleep(80)
  const sizeDialog = await evaluate(`(() => ({
    open: Boolean(document.querySelector('.visual-selector-dialog')),
    title: document.querySelector('.visual-selector-dialog h3')?.textContent,
    paragraphs: [...document.querySelectorAll('.visual-selector-dialog-copy p')].map((item) => item.textContent),
    status: document.querySelector('.visual-selector-dialog-status')?.textContent,
    activeClass: document.activeElement?.className,
    bodyModal: document.body.classList.contains('modal-open'),
  }))()`)
  assert(sizeDialog.open, 'Size info dialog did not open')
  assert(sizeDialog.title === 'Riesig', `Size dialog title mismatch: ${sizeDialog.title}`)
  assert(sizeDialog.status === 'Jetzt unterstützt', `Size dialog status mismatch: ${sizeDialog.status}`)
  assert(sizeDialog.paragraphs.some((text) => text?.includes('71 Sternsysteme')), `Huge server star-count fact missing: ${JSON.stringify(sizeDialog.paragraphs)}`)
  assert(sizeDialog.activeClass === 'visual-selector-dialog-close', `Size dialog close control did not receive focus: ${sizeDialog.activeClass}`)
  assert(sizeDialog.bodyModal, 'modal-open body state missing while Size dialog is open')
  await key('Escape', 'Escape', 27)
  await sleep(90)
  const sizeClosed = await evaluate(`({ open: Boolean(document.querySelector('.visual-selector-dialog')), activeClass: document.activeElement?.className, bodyModal: document.body.classList.contains('modal-open') })`)
  assert(!sizeClosed.open, 'Escape did not close Size info dialog')
  assert(sizeClosed.activeClass === 'visual-selector-info-button', `focus did not return to Size info control: ${sizeClosed.activeClass}`)
  assert(!sizeClosed.bodyModal, 'modal-open body state remained after Size close')

  await evaluate(`document.querySelector('.visual-selector[data-setting-id="galaxy-size"]')?.focus(); true`)
  await key('Home', 'Home', 36)
  const expectedSizes = [
    ['Klein', 'small', '20 Sternsysteme'],
    ['Mittel', 'medium', '36 Sternsysteme'],
    ['Groß', 'large', '54 Sternsysteme'],
    ['Riesig', 'huge', '71 Sternsysteme'],
  ]
  for (let index = 0; index < expectedSizes.length; index++) {
    if (index > 0) {
      await evaluate(`document.querySelectorAll('.visual-selector[data-setting-id="galaxy-size"] .visual-selector-arrow')[1]?.click(); true`)
      await sleep(70)
    }
    snap = await evaluate(snapshotExpression)
    const [label, id] = expectedSizes[index]
    assert(snap.selected === label, `Size option ${index + 1} label mismatch: expected ${label}, got ${snap.selected}`)
    assert(snap.source === `/assets/new-game/galaxy-size/${id}.svg`, `Size ${label} art path mismatch: ${snap.source}`)
    assert(snap.imageReady, `Size ${label} artwork did not load at native 1200x675`)
    assert(snap.availability === 'supported', `Size ${label} is not marked supported`)
    assert(snap.createDisabled === false, `Size ${label} incorrectly disabled Create Game`)
  }
  await evaluate(`document.querySelector('.visual-selector[data-setting-id="galaxy-size"]')?.focus(); true`)
  await key('Home', 'Home', 36)
  await sleep(70)
  snap = await evaluate(snapshotExpression)
  assert(snap.selected === 'Klein', 'Galaxy Size must return to Small before Age/Difficulty QA')

  const ageSnapshotExpression = `(() => {
    const root = document.querySelector('.visual-selector[data-setting-id="galaxy-age"]')
    const rect = (element) => { const r = element?.getBoundingClientRect(); return r ? { width: r.width, height: r.height } : null }
    const image = root?.querySelector('.new-game-galaxy-art img')
    const art = root?.querySelector('.visual-selector-art')
    const arrows = [...(root?.querySelectorAll('.visual-selector-arrow') ?? [])].map(rect)
    return {
      selected: root?.querySelector('.visual-selector-current h3')?.textContent,
      source: image?.getAttribute('src'),
      imageReady: Boolean(image?.complete && image?.naturalWidth === 1200 && image?.naturalHeight === 675),
      art: rect(art),
      radius: art ? getComputedStyle(art).borderRadius : '',
      info: rect(root?.querySelector('.visual-selector-info-button')),
      arrows,
      dotCount: root?.querySelectorAll('.visual-selector-dots i').length ?? 0,
      availability: root?.getAttribute('data-availability'),
      createDisabled: document.querySelector('.new-game-form button[type=submit]')?.disabled,
    }
  })()`

  let ageSnap = await evaluate(ageSnapshotExpression)
  assert(ageSnap.selected === 'Normal', `expected initial Galaxy Age Normal, got ${ageSnap.selected}`)
  assert(ageSnap.source === '/assets/new-game/galaxy-age/normal.svg', `unexpected initial Galaxy Age art source ${ageSnap.source}`)
  assert(ageSnap.imageReady, 'Normal Galaxy Age artwork did not load at native 1200x675')
  assert(ageSnap.info?.width >= 44 && ageSnap.info?.height >= 44, `mobile Age info target below 44px: ${JSON.stringify(ageSnap.info)}`)
  assert(ageSnap.arrows.every((item) => item?.width >= 44 && item?.height >= 44), `mobile Age arrow target below 44px: ${JSON.stringify(ageSnap.arrows)}`)
  assert(ageSnap.radius === '16px', `expected 16px mobile Age art radius, got ${ageSnap.radius}`)
  assert(ageSnap.dotCount === 3, `expected three Age position dots, got ${ageSnap.dotCount}`)
  assert(ageSnap.availability === 'supported', `Age Normal availability=${ageSnap.availability}`)
  assert(ageSnap.createDisabled === false, 'Age Normal must allow Create Game')

  await evaluate(`document.querySelector('.visual-selector[data-setting-id="galaxy-age"]')?.focus(); true`)
  await key('Home', 'Home', 36)
  await sleep(80)
  ageSnap = await evaluate(ageSnapshotExpression)
  assert(ageSnap.selected === 'Mineralreich', `Home should select Mineral Rich/Mineralreich, got ${ageSnap.selected}`)
  assert(ageSnap.source === '/assets/new-game/galaxy-age/mineral-rich.svg', `Mineral Rich art path wrong: ${ageSnap.source}`)
  assert(ageSnap.createDisabled === false, 'Mineral Rich must keep Create Game enabled')

  await evaluate(`document.querySelector('.visual-selector[data-setting-id="galaxy-age"] .visual-selector-info-button')?.click(); true`)
  await sleep(80)
  const ageDialog = await evaluate(`(() => ({
    title: document.querySelector('.visual-selector-dialog h3')?.textContent,
    paragraphs: [...document.querySelectorAll('.visual-selector-dialog-copy p')].map((item) => item.textContent),
    status: document.querySelector('.visual-selector-dialog-status')?.textContent,
  }))()`)
  assert(ageDialog.title === 'Mineralreich', `Age dialog title mismatch: ${ageDialog.title}`)
  assert(ageDialog.status === 'Jetzt unterstützt', `Age dialog status mismatch: ${ageDialog.status}`)
  assert(ageDialog.paragraphs.some((text) => text?.includes('Mineralreiche Welten: wahrscheinlicher')), `Mineral Rich server bias fact missing: ${JSON.stringify(ageDialog.paragraphs)}`)
  assert(ageDialog.paragraphs.some((text) => text?.includes('Nahrungsfreundliche Welten: weniger wahrscheinlich')), `Mineral Rich food bias fact missing: ${JSON.stringify(ageDialog.paragraphs)}`)
  await key('Escape', 'Escape', 27)
  await sleep(80)

  const expectedAges = [
    ['Mineralreich', 'mineral-rich'],
    ['Normal', 'normal'],
    ['Organisch reich', 'organic-rich'],
  ]
  await evaluate(`document.querySelector('.visual-selector[data-setting-id="galaxy-age"]')?.focus(); true`)
  await key('Home', 'Home', 36)
  for (let index = 0; index < expectedAges.length; index++) {
    if (index > 0) {
      await evaluate(`document.querySelectorAll('.visual-selector[data-setting-id="galaxy-age"] .visual-selector-arrow')[1]?.click(); true`)
      await sleep(70)
    }
    ageSnap = await evaluate(ageSnapshotExpression)
    const [label, assetId] = expectedAges[index]
    assert(ageSnap.selected === label, `Age option ${index + 1} label mismatch: expected ${label}, got ${ageSnap.selected}`)
    assert(ageSnap.source === `/assets/new-game/galaxy-age/${assetId}.svg`, `Age ${label} art path mismatch: ${ageSnap.source}`)
    assert(ageSnap.imageReady, `Age ${label} artwork did not load at native 1200x675`)
    assert(ageSnap.availability === 'supported', `Age ${label} is not marked supported`)
    assert(ageSnap.createDisabled === false, `Age ${label} incorrectly disabled Create Game`)
  }
  await evaluate(`document.querySelector('.visual-selector[data-setting-id="galaxy-age"]')?.focus(); true`)
  await key('Home', 'Home', 36)
  await key('ArrowRight', 'ArrowRight', 39)
  await sleep(80)
  ageSnap = await evaluate(ageSnapshotExpression)
  assert(ageSnap.selected === 'Normal' && ageSnap.createDisabled === false, 'Galaxy Age must return to Normal before Difficulty QA')

  const difficultySnapshotExpression = `(() => {
    const root = document.querySelector('.visual-selector[data-setting-id="difficulty"]')
    const rect = (element) => { const r = element?.getBoundingClientRect(); return r ? { width: r.width, height: r.height } : null }
    const image = root?.querySelector('.new-game-difficulty-art img')
    const art = root?.querySelector('.visual-selector-art')
    const arrows = [...(root?.querySelectorAll('.visual-selector-arrow') ?? [])].map(rect)
    return {
      selected: root?.querySelector('.visual-selector-current h3')?.textContent,
      source: image?.getAttribute('src'),
      imageReady: Boolean(image?.complete && image?.naturalWidth === 1200 && image?.naturalHeight === 675),
      art: rect(art),
      radius: art ? getComputedStyle(art).borderRadius : '',
      info: rect(root?.querySelector('.visual-selector-info-button')),
      arrows,
      dotCount: root?.querySelectorAll('.visual-selector-dots i').length ?? 0,
      availability: root?.getAttribute('data-availability'),
      createDisabled: document.querySelector('.new-game-form button[type=submit]')?.disabled,
    }
  })()`

  let difficultySnap = await evaluate(difficultySnapshotExpression)
  assert(difficultySnap.selected === 'Normal', `expected initial Difficulty Normal, got ${difficultySnap.selected}`)
  assert(difficultySnap.source === '/assets/new-game/difficulty/normal.svg', `unexpected initial Difficulty art source ${difficultySnap.source}`)
  assert(difficultySnap.imageReady, 'Normal Difficulty artwork did not load at native 1200x675')
  assert(difficultySnap.info?.width >= 44 && difficultySnap.info?.height >= 44, `mobile Difficulty info target below 44px: ${JSON.stringify(difficultySnap.info)}`)
  assert(difficultySnap.arrows.every((item) => item?.width >= 44 && item?.height >= 44), `mobile Difficulty arrow target below 44px: ${JSON.stringify(difficultySnap.arrows)}`)
  assert(difficultySnap.radius === '16px', `expected 16px mobile Difficulty art radius, got ${difficultySnap.radius}`)
  assert(difficultySnap.dotCount === 5, `expected five Difficulty position dots, got ${difficultySnap.dotCount}`)
  assert(difficultySnap.availability === 'supported', `Difficulty Normal availability=${difficultySnap.availability}`)
  assert(difficultySnap.createDisabled === false, 'Supported Difficulty Normal must allow Create Game with Small galaxy')

  await evaluate(`document.querySelector('.visual-selector[data-setting-id="difficulty"]')?.focus(); true`)
  await key('End', 'End', 35)
  await sleep(80)
  difficultySnap = await evaluate(difficultySnapshotExpression)
  assert(difficultySnap.selected === 'Unmöglich', `End should select Impossible/Unmöglich, got ${difficultySnap.selected}`)
  assert(difficultySnap.source === '/assets/new-game/difficulty/impossible.svg', `Impossible art path wrong: ${difficultySnap.source}`)
  assert(difficultySnap.createDisabled === false, 'Impossible must remain server-supported and keep Create Game enabled')

  await evaluate(`document.querySelector('.visual-selector[data-setting-id="difficulty"] .visual-selector-info-button')?.click(); true`)
  await sleep(80)
  const difficultyDialog = await evaluate(`(() => ({
    title: document.querySelector('.visual-selector-dialog h3')?.textContent,
    paragraphs: [...document.querySelectorAll('.visual-selector-dialog-copy p')].map((item) => item.textContent),
    status: document.querySelector('.visual-selector-dialog-status')?.textContent,
  }))()`)
  assert(difficultyDialog.title === 'Unmöglich', `Difficulty dialog title mismatch: ${difficultyDialog.title}`)
  assert(difficultyDialog.status === 'Jetzt unterstützt', `Difficulty dialog status mismatch: ${difficultyDialog.status}`)
  assert(difficultyDialog.paragraphs.length >= 6, `expected six server-derived Difficulty facts, got ${difficultyDialog.paragraphs.length}`)
  assert(difficultyDialog.paragraphs.some((text) => text?.includes('+0,75')), `Impossible food fact missing server value: ${JSON.stringify(difficultyDialog.paragraphs)}`)
  assert(difficultyDialog.paragraphs.some((text) => text?.includes('+1,5')), 'Impossible production/research server value missing')
  assert(difficultyDialog.paragraphs.some((text) => text?.includes('8 BC')), 'Impossible command-deficit server value missing')
  await key('Escape', 'Escape', 27)
  await sleep(80)

  await evaluate(`document.querySelector('.visual-selector[data-setting-id="difficulty"]')?.focus(); true`)
  await key('Home', 'Home', 36)
  await sleep(70)
  difficultySnap = await evaluate(difficultySnapshotExpression)
  assert(difficultySnap.selected === 'Leicht', `Home should select Easy/Leicht, got ${difficultySnap.selected}`)
  await touchTap('.visual-selector[data-setting-id="difficulty"] .visual-selector-arrow:not(:disabled):last-of-type')
  await sleep(90)
  difficultySnap = await evaluate(difficultySnapshotExpression)
  assert(difficultySnap.selected === 'Normal', `Difficulty touch next should select Normal, got ${difficultySnap.selected}`)

  await evaluate(`document.querySelector('.visual-selector[data-setting-id="difficulty"]')?.focus(); true`)
  await key('Home', 'Home', 36)
  const expectedDifficulties = [
    ['Leicht', 'easy'],
    ['Normal', 'normal'],
    ['Schwer', 'hard'],
    ['Sehr schwer', 'very-hard'],
    ['Unmöglich', 'impossible'],
  ]
  for (let index = 0; index < expectedDifficulties.length; index++) {
    if (index > 0) {
      await evaluate(`document.querySelectorAll('.visual-selector[data-setting-id="difficulty"] .visual-selector-arrow')[1]?.click(); true`)
      await sleep(70)
    }
    difficultySnap = await evaluate(difficultySnapshotExpression)
    if (!difficultySnap.selected) {
      const diagnostic = await evaluate(`({ href: location.href, body: document.body.innerText.slice(0, 1200), selectors: document.querySelectorAll('.visual-selector').length })`)
      console.error('Difficulty selector disappeared at iteration', index, diagnostic)
    }
    const [label, id] = expectedDifficulties[index]
    assert(difficultySnap.selected === label, `Difficulty option ${index + 1} label mismatch: expected ${label}, got ${difficultySnap.selected}`)
    assert(difficultySnap.source === `/assets/new-game/difficulty/${id}.svg`, `Difficulty ${label} art path mismatch: ${difficultySnap.source}`)
    assert(difficultySnap.imageReady, `Difficulty ${label} artwork did not load at native 1200x675`)
    assert(difficultySnap.availability === 'supported', `Difficulty ${label} is not marked supported`)
    assert(difficultySnap.createDisabled === false, `Difficulty ${label} incorrectly disabled Create Game`)
  }

  const raceSnapshotExpression = `(() => {
    const q = (selector) => document.querySelector(selector)
    const rect = (element) => { const r = element?.getBoundingClientRect(); return r ? { width: r.width, height: r.height } : null }
    const root = q('.visual-selector[data-setting-id="player-race"]')
    const art = root?.querySelector('.visual-selector-art')
    const image = root?.querySelector('.new-game-race-art img')
    return {
      selected: root?.querySelector('.visual-selector-current h3')?.textContent,
      statusText: root?.querySelector('.visual-selector-current-status')?.textContent?.trim() ?? '',
      selectorAria: root?.getAttribute('aria-label') ?? '',
      infoAria: root?.querySelector('.visual-selector-info-button')?.getAttribute('aria-label') ?? '',
      arrowAria: [...(root?.querySelectorAll('.visual-selector-arrow') ?? [])].map((item) => item.getAttribute('aria-label') ?? ''),
      availability: root?.getAttribute('data-availability'),
      source: image?.getAttribute('src') ?? null,
      imageReady: image ? Boolean(image.complete && image.naturalWidth === 1200 && image.naturalHeight === 1500) : false,
      placeholder: Boolean(root?.querySelector('.new-game-race-placeholder')),
      art: rect(art),
      ratio: art ? getComputedStyle(art).aspectRatio : '',
      radius: art ? getComputedStyle(art).borderRadius : '',
      info: rect(root?.querySelector('.visual-selector-info-button')),
      arrows: [...(root?.querySelectorAll('.visual-selector-arrow') ?? [])].map(rect),
      dotCount: root?.querySelectorAll('.visual-selector-dots i').length ?? 0,
      createDisabled: q('.new-game-form button[type=submit]')?.disabled,
      playerEmpire: q('.new-game-form label:nth-of-type(4) input')?.value,
      scrollWidth: document.documentElement.scrollWidth,
      viewportWidth: document.documentElement.clientWidth,
    }
  })()`

  let raceSnap = await evaluate(raceSnapshotExpression)
  assert(raceSnap.selected === 'Menschen', `expected initial Human/Menschen, got ${raceSnap.selected}`)
  assert(raceSnap.availability === 'supported', `Human availability=${raceSnap.availability}`)
  assert(raceSnap.source === '/assets/races/human/portrait.webp', `Human portrait path wrong: ${raceSnap.source}`)
  assert(raceSnap.imageReady, 'Human portrait did not load at native 1200x1500')
  assert(raceSnap.dotCount === 13, `expected 13 Race position dots, got ${raceSnap.dotCount}`)
  assert(raceSnap.art?.width > raceSnap.art?.height, `mobile Race chooser must use the shared landscape frame: ${JSON.stringify(raceSnap.art)}`)
  assert(raceSnap.ratio === '16 / 10', `expected shared 16/10 mobile Race aspect ratio, got ${raceSnap.ratio}`)
  assert(raceSnap.info?.width >= 44 && raceSnap.info?.height >= 44, `mobile Race info target below 44px: ${JSON.stringify(raceSnap.info)}`)
  assert(raceSnap.arrows.every((item) => item?.width >= 44 && item?.height >= 44), `mobile Race arrow target below 44px: ${JSON.stringify(raceSnap.arrows)}`)
  assert(raceSnap.scrollWidth === raceSnap.viewportWidth && raceSnap.viewportWidth === 390, `Race selector mobile horizontal overflow: ${raceSnap.scrollWidth}/${raceSnap.viewportWidth}`)
  assert(raceSnap.createDisabled === false, 'Human must allow Create Game')
  assert(raceSnap.statusText === 'Jetzt unterstützt', `Human visible availability text mismatch: ${raceSnap.statusText}`)
  assert(raceSnap.selectorAria === 'Spielerrasse', `Race selector aria-label mismatch: ${raceSnap.selectorAria}`)
  assert(raceSnap.infoAria.includes('Menschen'), `Race info aria-label must identify current race: ${raceSnap.infoAria}`)
  assert(raceSnap.arrowAria.every((label) => label.length > 0), `Race arrows require text aria-labels: ${JSON.stringify(raceSnap.arrowAria)}`)

  // Darlok planned-player composition preview: browsing must stay coherent without pretending the race is playable.
  await evaluate(`document.querySelector('.visual-selector[data-setting-id="player-race"]')?.focus(); true`)
  await key('Home', 'Home', 36)
  await key('ArrowRight', 'ArrowRight', 39)
  await key('ArrowRight', 'ArrowRight', 39)
  await sleep(90)
  raceSnap = await evaluate(raceSnapshotExpression)
  assert(raceSnap.selected === 'Darlok' && raceSnap.availability === 'planned', `Darlok planned state mismatch: ${JSON.stringify(raceSnap)}`)
  const darlokComposition = await evaluate(`(() => {
    const opponentRoot = document.querySelector('.visual-selector[data-setting-id="opponent-count"]')
    const local = document.querySelector('.new-game-opponent-chip[data-role="local-player"]')
    return {
      localText: local?.textContent ?? '',
      opponentChips: document.querySelectorAll('.new-game-opponent-chip[data-role="opponent"]').length,
      opponentAvailability: opponentRoot?.getAttribute('data-availability'),
      reason: document.querySelector('.new-game-composition-reason')?.textContent ?? '',
      reasonLayout: (() => { const el = document.querySelector('.new-game-composition-reason'); if (!el) return null; const cs = getComputedStyle(el); return { clientHeight: el.clientHeight, scrollHeight: el.scrollHeight, overflow: cs.overflow, whiteSpace: cs.whiteSpace, textOverflow: cs.textOverflow } })(),
      createDisabled: document.querySelector('.new-game-form button[type=submit]')?.disabled,
    }
  })()`)
  assert(darlokComposition.localText.includes('Darlok') && darlokComposition.localText.includes('Du'), `Darlok local-player preview missing: ${JSON.stringify(darlokComposition)}`)
  assert(darlokComposition.opponentChips === 0 && darlokComposition.opponentAvailability === 'planned', `Darlok must not fabricate an opponent assignment: ${JSON.stringify(darlokComposition)}`)
  assert(darlokComposition.reason.length > 0 && darlokComposition.createDisabled === true, `Darlok planned reason/create lock missing: ${JSON.stringify(darlokComposition)}`)
  assert(darlokComposition.reasonLayout && darlokComposition.reasonLayout.scrollHeight <= darlokComposition.reasonLayout.clientHeight + 1 && darlokComposition.reasonLayout.overflow === 'visible' && darlokComposition.reasonLayout.whiteSpace === 'normal' && darlokComposition.reasonLayout.textOverflow === 'clip', `Darlok planned reason is visually clipped: ${JSON.stringify(darlokComposition.reasonLayout)}`)
  await key('ArrowRight', 'ArrowRight', 39)
  await key('ArrowRight', 'ArrowRight', 39)
  await key('ArrowRight', 'ArrowRight', 39)
  await sleep(90)
  raceSnap = await evaluate(raceSnapshotExpression)
  assert(raceSnap.selected === 'Menschen' && raceSnap.availability === 'supported', 'Race selector must restore Human after Darlok preview smoke')
  const restoredCompositionReason = await evaluate(`document.querySelector('.new-game-composition-reason')?.textContent ?? ''`)
  assert(restoredCompositionReason === '', `planned player-race reason must disappear after restoring Human: ${restoredCompositionReason}`)

  await evaluate(`(() => { const image = document.querySelector('.visual-selector[data-setting-id="player-race"] .new-game-race-art img'); if (image) image.style.visibility = 'hidden'; return true })()`)
  raceSnap = await evaluate(raceSnapshotExpression)
  assert(raceSnap.selected === 'Menschen', 'Race name must remain visible with portrait hidden')
  assert(raceSnap.statusText === 'Jetzt unterstützt', 'Race availability text must remain visible with portrait hidden')
  assert(raceSnap.selectorAria === 'Spielerrasse' && raceSnap.infoAria.includes('Menschen'), 'Race ARIA identity must remain sufficient with portrait hidden')
  await evaluate(`(() => { const image = document.querySelector('.visual-selector[data-setting-id="player-race"] .new-game-race-art img'); if (image) image.style.visibility = ''; return true })()`)

  await evaluate(`document.querySelector('.visual-selector[data-setting-id="player-race"]')?.focus(); true`)
  await key('Home', 'Home', 36)
  await sleep(80)
  raceSnap = await evaluate(raceSnapshotExpression)
  assert(raceSnap.selected === 'Alkari', `Race Home should select Alkari, got ${raceSnap.selected}`)
  assert(raceSnap.availability === 'planned', `Alkari availability=${raceSnap.availability}`)
  assert(raceSnap.placeholder && !raceSnap.source, 'Planned Alkari must use the neutral planned visual, not another race portrait')
  assert(raceSnap.createDisabled === true, 'Planned Alkari must lock Create Game')
  assert(raceSnap.statusText === 'Geplant / gesperrt', `Planned Alkari visible availability text mismatch: ${raceSnap.statusText}`)

  await touchTap('.visual-selector[data-setting-id="player-race"] .visual-selector-arrow:not(:disabled):last-of-type')
  await sleep(90)
  raceSnap = await evaluate(raceSnapshotExpression)
  assert(raceSnap.selected === 'Bulrathi', `Race touch next should select Bulrathi, got ${raceSnap.selected}`)
  assert(raceSnap.createDisabled === true, 'Planned Bulrathi must keep Create Game locked')

  await evaluate(`document.querySelector('.visual-selector[data-setting-id="player-race"]')?.focus(); true`)
  await key('Home', 'Home', 36)
  await sleep(60)
  const expectedRaces = [
    ['Alkari', 'alkari', 'planned', false],
    ['Bulrathi', 'bulrathi', 'planned', false],
    ['Darlok', 'darlok', 'planned', true],
    ['Elerian', 'elerian', 'planned', false],
    ['Gnolam', 'gnolam', 'planned', false],
    ['Menschen', 'human', 'supported', true],
    ['Klackon', 'klackon', 'supported', true],
    ['Meklar', 'meklar', 'planned', false],
    ['Mrrshan', 'mrrshan', 'planned', false],
    ['Psilon', 'psilon', 'planned', false],
    ['Sakkra', 'sakkra', 'planned', false],
    ['Silicoid', 'silicoid', 'planned', false],
    ['Trilarian', 'trilarian', 'planned', false],
  ]
  for (let index = 0; index < expectedRaces.length; index++) {
    if (index > 0) {
      await evaluate(`document.querySelectorAll('.visual-selector[data-setting-id="player-race"] .visual-selector-arrow')[1]?.click(); true`)
      await sleep(60)
    }
    raceSnap = await evaluate(raceSnapshotExpression)
    const [label, id, availability, hasPortrait] = expectedRaces[index]
    assert(raceSnap.selected === label, `Race option ${index + 1} label mismatch: expected ${label}, got ${raceSnap.selected}`)
    assert(raceSnap.availability === availability, `Race ${label} availability=${raceSnap.availability} want=${availability}`)
    assert(raceSnap.createDisabled === (availability !== 'supported'), `Race ${label} create lock mismatch: disabled=${raceSnap.createDisabled}`)
    if (hasPortrait) {
      assert(raceSnap.source === `/assets/races/${id}/portrait.webp`, `Race ${label} portrait path mismatch: ${raceSnap.source}`)
      assert(raceSnap.imageReady, `Race ${label} portrait did not load at native 1200x1500`)
      assert(!raceSnap.placeholder, `Race ${label} unexpectedly shows planned placeholder`)
    } else {
      assert(raceSnap.source === null && raceSnap.placeholder, `Race ${label} must use planned placeholder without borrowed portrait`)
    }
  }

  await evaluate(`document.querySelector('.visual-selector[data-setting-id="player-race"]')?.focus(); true`)
  await key('Home', 'Home', 36)
  for (let index = 0; index < 6; index++) {
    await key('ArrowRight', 'ArrowRight', 39)
    await sleep(45)
  }
  raceSnap = await evaluate(raceSnapshotExpression)
  assert(raceSnap.selected === 'Klackon', `expected Klackon after six Race ArrowRight steps, got ${raceSnap.selected}`)
  assert(raceSnap.playerEmpire === 'Klackon', `default player empire did not follow supported Klackon selection: ${raceSnap.playerEmpire}`)
  await evaluate(`document.querySelector('.visual-selector[data-setting-id="player-race"] .visual-selector-info-button')?.click(); true`)
  await sleep(80)
  const raceDialog = await evaluate(`(() => ({
    title: document.querySelector('.visual-selector-dialog h3')?.textContent,
    paragraphs: [...document.querySelectorAll('.visual-selector-dialog-copy p')].map((item) => item.textContent),
    activeClass: document.activeElement?.className,
  }))()`)
  assert(raceDialog.title === 'Klackon', `Klackon dialog title mismatch: ${raceDialog.title}`)
  for (const fact of ['Vereinigung', '+1 Nahrung', '+1 Industrie', 'Unkreativ']) {
    assert(raceDialog.paragraphs.includes(fact), `Klackon server-curated fact missing: ${fact}; got ${JSON.stringify(raceDialog.paragraphs)}`)
  }
  assert(raceDialog.activeClass === 'visual-selector-dialog-close', `Race dialog close control did not receive focus: ${raceDialog.activeClass}`)
  await key('Escape', 'Escape', 27)
  await sleep(80)

  const technologySnapshotExpression = `(() => {
    const root = document.querySelector('.visual-selector[data-setting-id="technology-level"]')
    const rect = (element) => { const r = element?.getBoundingClientRect(); return r ? { width: r.width, height: r.height } : null }
    const image = root?.querySelector('.new-game-technology-art img')
    const art = root?.querySelector('.visual-selector-art')
    const title = root?.querySelector('.visual-selector-current h3')
    const arrows = [...(root?.querySelectorAll('.visual-selector-arrow') ?? [])].map(rect)
    let fingerprint = null
    if (image?.complete && image.naturalWidth > 0) {
      const canvas = document.createElement('canvas')
      canvas.width = 64
      canvas.height = 36
      const context = canvas.getContext('2d', { willReadFrequently: true })
      context.drawImage(image, 0, 0, 64, 36)
      const data = context.getImageData(0, 0, 64, 36).data
      let hash = 2166136261 >>> 0
      for (let index = 0; index < data.length; index += 4) {
        hash ^= data[index]; hash = Math.imul(hash, 16777619)
        hash ^= data[index + 1]; hash = Math.imul(hash, 16777619)
        hash ^= data[index + 2]; hash = Math.imul(hash, 16777619)
      }
      fingerprint = (hash >>> 0).toString(16)
    }
    return {
      selected: title?.textContent,
      source: image?.getAttribute('src'),
      imageReady: Boolean(image?.complete && image?.naturalWidth === 1200 && image?.naturalHeight === 675),
      image: rect(image),
      art: rect(art),
      radius: art ? getComputedStyle(art).borderRadius : '',
      whiteSpace: title ? getComputedStyle(title).whiteSpace : '',
      titleFits: title ? title.scrollWidth <= title.clientWidth : false,
      arrows,
      availability: root?.getAttribute('data-availability'),
      statusText: root?.querySelector('.visual-selector-current-status')?.textContent,
      createDisabled: document.querySelector('.new-game-form button[type=submit]')?.disabled,
      scrollWidth: document.documentElement.scrollWidth,
      viewportWidth: document.documentElement.clientWidth,
      fingerprint,
    }
  })()`

  await evaluate(`document.querySelector('.visual-selector[data-setting-id="technology-level"]')?.focus(); true`)
  await key('Home', 'Home', 36)
  await sleep(90)
  const expectedTechnologies = [
    ['Pre-Warp', 'pre-warp', 'supported'],
    ['Durchschnitt', 'average', 'supported'],
    ['Fortschrittlich', 'advanced', 'planned'],
  ]
  const technologyFingerprints = new Set()
  let technologySnap
  for (let index = 0; index < expectedTechnologies.length; index++) {
    if (index > 0) {
      await evaluate(`document.querySelectorAll('.visual-selector[data-setting-id="technology-level"] .visual-selector-arrow')[1]?.click(); true`)
      await sleep(90)
    }
    technologySnap = await evaluate(technologySnapshotExpression)
    const [label, assetID, availability] = expectedTechnologies[index]
    assert(technologySnap.selected === label, `Technology option ${index + 1} label mismatch: expected ${label}, got ${technologySnap.selected}`)
    assert(technologySnap.source === `/assets/new-game/technology-level/${assetID}.svg?v=slice16-5-g3-art2`, `Technology ${label} art path mismatch: ${technologySnap.source}`)
    assert(technologySnap.imageReady, `Technology ${label} artwork did not load at native 1200x675`)
    assert(technologySnap.art?.width > 300 && technologySnap.art?.width < 360, `mobile Technology ${label} art width unexpected: ${technologySnap.art?.width}`)
    assert(Math.abs((technologySnap.image?.width ?? 0) - (technologySnap.art?.width ?? 0)) <= 3, `Technology ${label} image is not scaled to its frame: ${JSON.stringify(technologySnap)}`)
    assert(Math.abs((technologySnap.image?.height ?? 0) - (technologySnap.art?.height ?? 0)) <= 3, `Technology ${label} image height does not match frame`)
    assert(technologySnap.whiteSpace === 'nowrap' && technologySnap.titleFits, `Technology ${label} title is not a one-line fit at 390px`)
    assert(technologySnap.arrows.every((item) => item?.width >= 44 && item?.height >= 44), `mobile Technology arrow target below 44px: ${JSON.stringify(technologySnap.arrows)}`)
    assert(technologySnap.availability === availability, `Technology ${label} availability=${technologySnap.availability} want=${availability}`)
    assert(technologySnap.createDisabled === (availability !== 'supported'), `Technology ${label} Create Game lock mismatch`)
    assert(technologySnap.scrollWidth === technologySnap.viewportWidth && technologySnap.viewportWidth === 390, `Technology ${label} caused mobile horizontal overflow: ${technologySnap.scrollWidth}/${technologySnap.viewportWidth}`)
    if (technologySnap.fingerprint) technologyFingerprints.add(technologySnap.fingerprint)
  }
  assert(technologyFingerprints.size === 3, `Technology artwork fingerprints are not distinct: ${JSON.stringify([...technologyFingerprints])}`)
  assert(technologySnap.statusText === 'Geplant / gesperrt', `Advanced visible availability text mismatch: ${technologySnap.statusText}`)
  // Opponent composition mobile smoke: restore supported technology, then browse 1/2/planned-3.
  await evaluate(`document.querySelector('.visual-selector[data-setting-id="technology-level"]')?.focus(); true`)
  await key('Home', 'Home', 36)
  await sleep(90)
  const opponentSnapshotExpression = `(() => {
    const root = document.querySelector('.visual-selector[data-setting-id="opponent-count"]')
    const image = root?.querySelector('.new-game-opponent-count-art img')
    const title = root?.querySelector('.visual-selector-current h3')
    const rect = image?.getBoundingClientRect()
    const artRect = root?.querySelector('.visual-selector-art')?.getBoundingClientRect()
    return {
      selected: title?.textContent,
      source: image?.getAttribute('src'),
      imageReady: Boolean(image?.complete && image?.naturalWidth === 1200 && image?.naturalHeight === 675),
      objectFit: image ? getComputedStyle(image).objectFit : '',
      objectPosition: image ? getComputedStyle(image).objectPosition : '',
      width: rect?.width,
      artWidth: artRect?.width,
      availability: root?.getAttribute('data-availability'),
      statusText: root?.querySelector('.visual-selector-current-status')?.textContent,
      opponentChips: document.querySelectorAll('.new-game-opponent-chip[data-role="opponent"]').length,
      localPlayerText: document.querySelector('.new-game-opponent-chip[data-role="local-player"]')?.textContent ?? '',
      reason: document.querySelector('.new-game-composition-reason')?.textContent ?? '',
      createDisabled: document.querySelector('.new-game-form button[type=submit]')?.disabled,
      launchText: document.querySelector('.new-game-launch-briefing')?.textContent ?? '',
      scrollWidth: document.documentElement.scrollWidth,
      viewportWidth: document.documentElement.clientWidth,
    }
  })()`
  await evaluate(`document.querySelector('.visual-selector[data-setting-id="opponent-count"]')?.focus(); true`)
  await key('Home', 'Home', 36)
  await sleep(90)
  let opponentSnap = await evaluate(opponentSnapshotExpression)
  assert(opponentSnap.selected === '1 Gegner', `opponent count 1 label mismatch: ${opponentSnap.selected}`)
  assert(opponentSnap.source === '/assets/new-game/opponent-count/1.svg' && opponentSnap.imageReady, `opponent count 1 art invalid: ${JSON.stringify(opponentSnap)}`)
  assert(opponentSnap.objectFit === 'contain' && opponentSnap.objectPosition === '50% 50%', `opponent SVG must scale fully without cropping: ${JSON.stringify(opponentSnap)}`)
  assert(opponentSnap.availability === 'supported' && opponentSnap.opponentChips === 1 && opponentSnap.createDisabled === false, `opponent count 1 support mismatch: ${JSON.stringify(opponentSnap)}`)
  assert(opponentSnap.launchText.includes('Darlok'), 'launch briefing must include Darlok for one opponent')
  await evaluate(`document.querySelectorAll('.visual-selector[data-setting-id="opponent-count"] .visual-selector-arrow')[1]?.click(); true`)
  await sleep(90)
  opponentSnap = await evaluate(opponentSnapshotExpression)
  assert(opponentSnap.selected === '2 Gegner' && opponentSnap.source === '/assets/new-game/opponent-count/2.svg', `opponent count 2 mismatch: ${JSON.stringify(opponentSnap)}`)
  assert(opponentSnap.availability === 'supported' && opponentSnap.opponentChips === 2 && opponentSnap.createDisabled === false, `opponent count 2 support mismatch: ${JSON.stringify(opponentSnap)}`)
  assert(opponentSnap.launchText.includes('Darlok') && (opponentSnap.launchText.includes('Klackon') || opponentSnap.launchText.includes('Human')), 'launch briefing must include both resolved opponents')
  await evaluate(`document.querySelectorAll('.visual-selector[data-setting-id="opponent-count"] .visual-selector-arrow')[1]?.click(); true`)
  await sleep(90)
  opponentSnap = await evaluate(opponentSnapshotExpression)
  assert(opponentSnap.selected === '3 Gegner' && opponentSnap.source === '/assets/new-game/opponent-count/3.svg', `opponent count 3 mismatch: ${JSON.stringify(opponentSnap)}`)
  assert(opponentSnap.availability === 'planned' && opponentSnap.opponentChips === 0 && opponentSnap.createDisabled === true, `planned opponent count mismatch: ${JSON.stringify(opponentSnap)}`)
  assert(opponentSnap.reason.length > 0, 'planned opponent count must show server-derived reason')
  assert(opponentSnap.scrollWidth === opponentSnap.viewportWidth && opponentSnap.viewportWidth === 390, `opponent selector mobile overflow: ${opponentSnap.scrollWidth}/${opponentSnap.viewportWidth}`)

  await send('Emulation.setDeviceMetricsOverride', { width: 1280, height: 900, deviceScaleFactor: 1, mobile: false })
  await sleep(90)
  const chooserParity = await evaluate(`(() => {
    const race = document.querySelector('.visual-selector[data-setting-id="player-race"] .visual-selector-art')?.getBoundingClientRect()
    const opponent = document.querySelector('.visual-selector[data-setting-id="opponent-count"] .visual-selector-art')?.getBoundingClientRect()
    return { race: race ? { width: race.width, height: race.height } : null, opponent: opponent ? { width: opponent.width, height: opponent.height } : null }
  })()`)
  assert(chooserParity.race && chooserParity.opponent && Math.abs(chooserParity.race.height - chooserParity.opponent.height) <= 1 && Math.abs(chooserParity.race.width - chooserParity.opponent.width) <= 1, `Race/Opponent chooser desktop height mismatch: ${JSON.stringify(chooserParity)}`)
  await sleep(120)
  snap = await evaluate(snapshotExpression)
  assert(snap.scrollWidth === snap.viewportWidth, `desktop horizontal overflow: ${snap.scrollWidth}/${snap.viewportWidth}`)
  assert(snap.art?.width >= 400 && snap.art?.width <= 620, `desktop artwork width outside responsive compact range: ${snap.art?.width}`)
  assert(snap.radius === '20px', `expected 20px desktop art radius, got ${snap.radius}`)
  assert(snap.arrows.every((item) => item?.width >= 52 && item?.height >= 52), `desktop arrow target below 52px: ${JSON.stringify(snap.arrows)}`)
  difficultySnap = await evaluate(difficultySnapshotExpression)
  assert(difficultySnap.art?.width >= 400 && difficultySnap.art?.width <= 620, `desktop Difficulty artwork width outside responsive compact range: ${difficultySnap.art?.width}`)
  assert(difficultySnap.radius === '20px', `expected 20px desktop Difficulty art radius, got ${difficultySnap.radius}`)
  assert(difficultySnap.arrows.every((item) => item?.width >= 52 && item?.height >= 52), `desktop Difficulty arrow target below 52px: ${JSON.stringify(difficultySnap.arrows)}`)
  ageSnap = await evaluate(ageSnapshotExpression)
  assert(ageSnap.art?.width >= 400 && ageSnap.art?.width <= 620, `desktop Galaxy Age artwork width outside responsive compact range: ${ageSnap.art?.width}`)
  technologySnap = await evaluate(technologySnapshotExpression)
  assert(technologySnap.art?.width >= 400 && technologySnap.art?.width <= 620, `desktop Technology artwork width outside responsive compact range: ${technologySnap.art?.width}`)
  assert(Math.abs((technologySnap.image?.width ?? 0) - (technologySnap.art?.width ?? 0)) <= 3, `desktop Technology image is not scaled to its frame: ${JSON.stringify(technologySnap)}`)
  assert(technologySnap.radius === '20px', `expected 20px desktop Technology art radius, got ${technologySnap.radius}`)
  assert(technologySnap.whiteSpace === 'nowrap' && technologySnap.titleFits, 'desktop Technology title is not a one-line fit')
  assert(technologySnap.arrows.every((item) => item?.width >= 52 && item?.height >= 52), `desktop Technology arrow target below 52px: ${JSON.stringify(technologySnap.arrows)}`)
  assert(technologySnap.scrollWidth === technologySnap.viewportWidth, `desktop Technology horizontal overflow: ${technologySnap.scrollWidth}/${technologySnap.viewportWidth}`)
  raceSnap = await evaluate(raceSnapshotExpression)
  assert(raceSnap.art?.width > raceSnap.art?.height, `desktop Race chooser must use the shared landscape frame: ${JSON.stringify(raceSnap.art)}`)
  assert(raceSnap.ratio === '16 / 9', `desktop Race aspect ratio mismatch: ${raceSnap.ratio}`)
  assert(raceSnap.arrows.every((item) => item?.width >= 52 && item?.height >= 52), `desktop Race arrow target below 52px: ${JSON.stringify(raceSnap.arrows)}`)
  assert(ageSnap.radius === '20px', `expected 20px desktop Galaxy Age art radius, got ${ageSnap.radius}`)
  assert(ageSnap.arrows.every((item) => item?.width >= 52 && item?.height >= 52), `desktop Galaxy Age arrow target below 52px: ${JSON.stringify(ageSnap.arrows)}`)

  if (failures.length > 0) {
    console.error(`New Game selector browser smoke failed (${failures.length}):`)
    for (const failure of failures) console.error(`- ${failure}`)
    process.exitCode = 1
    return
  }

  console.log('New Game selector browser smoke passed: Difficulty + Galaxy Size + Galaxy Age + Starting Technology + Player Race + Opponent Composition, 35 browsable options, 390px mobile + desktop layout, distinct technology/opponent artwork, server-derived support locks/reasons, launch briefing, focus and authoritative behavior.')
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
