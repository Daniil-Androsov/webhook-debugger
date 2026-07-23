import { useState } from 'react'

// tokenize formatted JSON string into typed spans
function tokenizeJson(str) {
  const tokens = []
  // regex: captures strings, numbers, booleans/null, keys, punctuation
  const re = /("(?:\\.|[^"\\])*")(\s*:)?|(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)|(\btrue\b|\bfalse\b|\bnull\b)|([{}[\],])/g
  let last = 0
  let match

  while ((match = re.exec(str)) !== null) {
    if (match.index > last) {
      tokens.push({ type: 'whitespace', value: str.slice(last, match.index) })
    }

    if (match[1] !== undefined) {
      if (match[2]) {
        // key: string followed by colon
        tokens.push({ type: 'key', value: match[1] })
        tokens.push({ type: 'punctuation', value: match[2] })
      } else {
        tokens.push({ type: 'string', value: match[1] })
      }
    } else if (match[3] !== undefined) {
      tokens.push({ type: 'number', value: match[3] })
    } else if (match[4] !== undefined) {
      tokens.push({ type: 'keyword', value: match[4] })
    } else if (match[5] !== undefined) {
      tokens.push({ type: 'punctuation', value: match[5] })
    }

    last = re.lastIndex
  }

  if (last < str.length) {
    tokens.push({ type: 'whitespace', value: str.slice(last) })
  }

  return tokens
}

const TOKEN_CLASS = {
  key:        'text-blue-300',
  string:     'text-green-300',
  number:     'text-orange-300',
  keyword:    'text-purple-300',
  punctuation:'text-gray-400',
  whitespace: '',
}

function JsonHighlight({ body }) {
  let parsed
  try {
    parsed = JSON.parse(body)
  } catch {
    return <pre className="text-xs text-gray-300 whitespace-pre-wrap break-all">{body}</pre>
  }

  const formatted = JSON.stringify(parsed, null, 2)
  const tokens = tokenizeJson(formatted)

  return (
    <pre className="text-xs whitespace-pre-wrap break-all">
      {tokens.map((tok, i) =>
        tok.type === 'whitespace'
          ? tok.value
          : <span key={i} className={TOKEN_CLASS[tok.type]}>{tok.value}</span>
      )}
    </pre>
  )
}

function buildRaw(request) {
  if (!request) return ''
  const lines = [`${request.method} ${request.path} HTTP/1.1`]
  for (const [k, v] of Object.entries(request.headers ?? {})) {
    lines.push(`${k}: ${v}`)
  }
  if (request.body) {
    lines.push('', request.body)
  }
  return lines.join('\n')
}

const TABS = ['Headers', 'Body', 'Raw']

export default function RequestDetail({ request, onReplay, onCopyCurl }) {
  const [tab, setTab] = useState('Headers')
  const [replayURL, setReplayURL] = useState('')
  const [replayOpen, setReplayOpen] = useState(false)
  const [replayStatus, setReplayStatus] = useState(null)

  if (!request) {
    return (
      <div className="flex-1 flex items-center justify-center text-gray-600 text-xs">
        Select a request
      </div>
    )
  }

  async function handleReplay() {
    if (!replayURL) return
    setReplayStatus('sending…')
    try {
      await onReplay(request.id, replayURL)
      setReplayStatus('sent')
    } catch (e) {
      setReplayStatus('error: ' + e)
    }
  }

  return (
    <div className="flex flex-col h-full">
      {/* request summary */}
      <div className="px-4 py-3 border-b border-[#30363d] flex items-center gap-3">
        <span className="font-bold text-blue-400">{request.method}</span>
        <span className="flex-1 truncate text-gray-200">{request.path}</span>
        <span className="text-gray-500 text-xs">{request.status > 0 ? request.status : '—'}</span>
        <button
          onClick={() => { setReplayOpen(o => !o); setReplayStatus(null) }}
          className="px-2 py-1 text-xs rounded bg-[#1f2937] hover:bg-[#2d3748] text-gray-300 border border-[#30363d]"
        >
          Replay
        </button>
        <button
          onClick={() => onCopyCurl(request.id)}
          className="px-2 py-1 text-xs rounded bg-[#1f2937] hover:bg-[#2d3748] text-gray-300 border border-[#30363d]"
        >
          Copy as curl
        </button>
      </div>

      {/* replay bar */}
      {replayOpen && (
        <div className="flex items-center gap-2 px-4 py-2 border-b border-[#30363d] bg-[#161b22]">
          <input
            type="text"
            value={replayURL}
            onChange={e => setReplayURL(e.target.value)}
            placeholder="https://your-target.com"
            className="flex-1 bg-[#0d1117] border border-[#30363d] rounded px-2 py-1 text-xs text-gray-200 outline-none focus:border-blue-500"
          />
          <button
            onClick={handleReplay}
            className="px-3 py-1 text-xs rounded bg-blue-600 hover:bg-blue-500 text-white"
          >
            Send
          </button>
          {replayStatus && <span className="text-xs text-gray-500">{replayStatus}</span>}
        </div>
      )}

      {/* tabs */}
      <div className="flex border-b border-[#30363d]">
        {TABS.map(t => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-2 text-xs border-b-2 transition-colors ${tab === t ? 'border-blue-500 text-blue-400' : 'border-transparent text-gray-500 hover:text-gray-300'}`}
          >
            {t}
          </button>
        ))}
      </div>

      {/* tab content */}
      <div className="flex-1 overflow-auto p-4">
        {tab === 'Headers' && (
          <table className="w-full text-xs">
            <tbody>
              {Object.entries(request.headers ?? {}).map(([k, v]) => (
                <tr key={k} className="border-b border-[#1f2937]">
                  <td className="py-1 pr-4 text-gray-500 align-top w-48 shrink-0">{k}</td>
                  <td className="py-1 text-gray-200 break-all">{v}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {tab === 'Body' && (
          request.body
            ? <JsonHighlight body={request.body} />
            : <span className="text-xs text-gray-600">No body</span>
        )}

        {tab === 'Raw' && (
          <pre className="text-xs text-gray-300 whitespace-pre-wrap break-all">
            {buildRaw(request)}
          </pre>
        )}
      </div>
    </div>
  )
}
