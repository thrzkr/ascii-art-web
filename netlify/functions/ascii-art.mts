import { createHash } from 'node:crypto'
import { existsSync, readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

// Port of the Go server (main.go + functions/*.go) so the app can run on Netlify.

const DEFAULT_BANNER = 'standard'
const MAX_INPUT_SIZE = 10 * 1024
const MAX_TEXT_LENGTH = 250
const MAX_LINES = 250

// Allowed normalized (LF-only) SHA-256 checksums
const bannerChecksums: Record<string, string> = {
  standard: 'c3ec7584fb7ecfbd739e6b3f6f63fd1fe557d2ae3e24f870730d9cf8b2559e94',
  shadow: '78ccd616680eb9068fe1465db1c852ceaffd8c0f318e3aa0414e1635508e85bf',
  thinkertoy: 'e3c7a11f41a473d9b0d3bf2132a8f6dabb754bd16efa3897fa835a432d3b9caa',
}

class NotFoundError extends Error {}

function resolveFile(relative: string): string {
  const here = dirname(fileURLToPath(import.meta.url))
  const candidates = [join(process.cwd(), relative), join(here, '..', '..', relative), join(here, relative)]
  for (const candidate of candidates) {
    if (existsSync(candidate)) return candidate
  }
  throw new NotFoundError(relative)
}

function readText(relative: string): string {
  return readFileSync(resolveFile(relative), 'utf8')
}

const normalizeText = (text: string) => text.replace(/\r\n/g, '\n')

const escapeHtml = (value: string) =>
  value.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&#34;').replace(/'/g, '&#39;')

// Returns an error message when an existing banner file has been modified; missing files are skipped.
function checkBannerChanges(): string | null {
  for (const [name, expected] of Object.entries(bannerChecksums)) {
    let data: string
    try {
      data = readText(`banners/${name}.txt`)
    } catch (err) {
      if (err instanceof NotFoundError) continue
      throw err
    }
    const actual = createHash('sha256').update(normalizeText(data)).digest('hex')
    if (actual !== expected) return `banner file has been modified: banners/${name}.txt`
  }
  return null
}

function validInput(text: string, banner: string): boolean {
  if (text === '' || [...text].length > MAX_TEXT_LENGTH || text.split('\n').length > MAX_LINES) return false
  if (!(banner in bannerChecksums)) return false
  for (const ch of text) {
    const code = ch.codePointAt(0)!
    if (ch !== '\n' && (code < 32 || code > 126)) return false
  }
  return true
}

function generateASCII(text: string, banner: string): string {
  const lines = normalizeText(readText(`banners/${banner}.txt`)).split('\n')
  if (lines.length !== 855) throw new Error('invalid banner file')

  let result = ''
  for (const line of normalizeText(text).split('\n')) {
    for (let row = 0; row < 8; row++) {
      for (const ch of line) {
        result += lines[(ch.charCodeAt(0) - 32) * 9 + row + 1]
      }
      result += '\n'
    }
  }
  return result
}

function randomSplash(): string {
  try {
    const lines = readText('static/splash.txt')
      .split('\n')
      .map((line) => line.trim())
      .filter(Boolean)
    if (lines.length > 0) return lines[Math.floor(Math.random() * lines.length)]
  } catch {}
  return 'A NEW WORLD AWAITS'
}

const html = (body: string, status: number) =>
  new Response(body, { status, headers: { 'Content-Type': 'text/html; charset=utf-8' } })

function errorPage(status: number, message: string, headers: Record<string, string> = {}): Response {
  let body: string
  try {
    body = readText('templates/error.html')
      .replace(/\{\{\.StatusCode\}\}/g, String(status))
      .replace(/\{\{\.Message\}\}/g, escapeHtml(message))
  } catch {
    return new Response(message, { status, headers: { 'Content-Type': 'text/plain; charset=utf-8', ...headers } })
  }
  const res = html(body, status)
  for (const [key, value] of Object.entries(headers)) res.headers.set(key, value)
  return res
}

function renderPage(data: { text: string; banner: string; splash: string; result: string }): Response {
  let template: string
  try {
    template = readText('templates/index.html')
  } catch (err) {
    if (err instanceof NotFoundError) return errorPage(404, 'Missing page template file')
    return errorPage(500, 'Internal server error')
  }

  const body = template
    .replace(/\{\{if eq \.Banner "(\w+)"\}\}(.*?)\{\{end\}\}/g, (_, name, inner) => (data.banner === name ? inner : ''))
    .replace(/\{\{if \.Result\}\}([\s\S]*?)\{\{else\}\}([\s\S]*?)\{\{end\}\}/, (_, withResult, without) =>
      data.result ? withResult : without,
    )
    .replace(/\{\{\.Splash\}\}/g, () => escapeHtml(data.splash))
    .replace(/\{\{\.Text\}\}/g, () => escapeHtml(data.text))
    .replace(/\{\{\.Result\}\}/g, () => escapeHtml(data.result))
  return html(body, 200)
}

async function asciiArt(req: Request): Promise<Response> {
  const raw = await req.text()
  if (new TextEncoder().encode(raw).length > MAX_INPUT_SIZE) return errorPage(400, 'Invalid form data')

  let form: URLSearchParams
  try {
    form = new URLSearchParams(raw)
  } catch {
    return errorPage(400, 'Invalid form data')
  }

  const text = normalizeText(form.get('text') ?? '')
  const banner = form.get('banner') || DEFAULT_BANNER
  if (!validInput(text, banner)) return errorPage(400, 'Invalid input')

  let result: string
  try {
    result = generateASCII(text, banner)
  } catch (err) {
    if (err instanceof NotFoundError) return errorPage(404, 'Missing Banner file')
    return errorPage(500, 'Internal server error')
  }

  return renderPage({ text, banner, splash: randomSplash(), result })
}

export default async (req: Request) => {
  const tampered = checkBannerChanges()
  if (tampered) {
    console.error(tampered)
    return errorPage(500, 'Internal server error')
  }

  const { pathname } = new URL(req.url)

  if (pathname === '/') {
    if (req.method !== 'GET') return errorPage(405, 'Method not allowed', { Allow: 'GET' })
    return renderPage({ text: '', banner: DEFAULT_BANNER, splash: randomSplash(), result: '' })
  }

  if (pathname === '/ascii-art') {
    if (req.method !== 'POST') return errorPage(405, 'Method not allowed', { Allow: 'POST' })
    return asciiArt(req)
  }

  return errorPage(404, 'Page not found')
}

export const config = {
  path: '/*',
  excludedPath: '/static/*',
}
