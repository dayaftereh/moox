import fs from 'node:fs'
import path from 'node:path'

const roots = ['src', 'public']
const textExtensions = new Set(['.ts', '.tsx', '.js', '.jsx', '.mjs', '.css', '.html', '.json', '.md', '.txt', '.svg'])
const suspiciousSequences = [
  'Â·', 'Â ', 'Â°',
  'â€”', 'â€“', 'â€¦', 'â†’', 'â†', 'â€º', 'â€¹', 'â€™', 'â€œ', 'â€',
  'Ã—', 'Ã¤', 'Ã¶', 'Ã¼', 'Ã„', 'Ã–', 'Ãœ', 'ÃŸ', 'Ãƒ', 'Ã‚',
  '�',
]

const findings = []

function scanFile(filePath) {
  const text = fs.readFileSync(filePath, 'utf8')
  const lines = text.split(/\r?\n/)
  lines.forEach((line, index) => {
    const matches = suspiciousSequences.filter((sequence) => line.includes(sequence))
    if (matches.length > 0) findings.push({ filePath, line: index + 1, matches, text: line.trim() })
  })
}

function walk(root) {
  if (!fs.existsSync(root)) return
  for (const entry of fs.readdirSync(root, { withFileTypes: true })) {
    const filePath = path.join(root, entry.name)
    if (entry.isDirectory()) {
      walk(filePath)
      continue
    }
    if (!textExtensions.has(path.extname(entry.name))) continue
    scanFile(filePath)
  }
}

for (const root of roots) walk(root)

if (findings.length > 0) {
  console.error(`Mojibake check failed: ${findings.length} suspicious line(s) found.`)
  for (const finding of findings) {
    console.error(`${finding.filePath}:${finding.line} [${finding.matches.join(', ')}] ${finding.text}`)
  }
  process.exit(1)
}

console.log('Mojibake check passed: runtime text is clean UTF-8.')
