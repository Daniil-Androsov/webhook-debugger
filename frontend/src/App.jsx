import { useState, useEffect } from 'react'
import RequestList from './RequestList'
import RequestDetail from './RequestDetail'
import './style.css'

const go = window?.go?.main?.App

export default function App() {
  const [requests, setRequests] = useState([])
  const [selected, setSelected] = useState(null)
  const [forwardURL, setForwardURL] = useState('')
  const [copied, setCopied] = useState(false)
  const [tunnelURL, setTunnelURL] = useState('')
  const [tunnelLoading, setTunnelLoading] = useState(false)
  const [tunnelError, setTunnelError] = useState('')

  useEffect(() => {
    if (!go) return

    go.GetRequests().then(data => {
      setRequests(data ?? [])
    })

    const runtime = window?.runtime
    if (!runtime) return

    runtime.EventsOn('new_request', req => {
      setRequests(prev => [req, ...prev])
    })
    runtime.EventsOn('tunnel_started', url => {
      setTunnelURL(url)
      setTunnelLoading(false)
    })
    runtime.EventsOn('tunnel_stopped', () => {
      setTunnelURL('')
    })

    return () => {
      runtime.EventsOff('new_request')
      runtime.EventsOff('tunnel_started')
      runtime.EventsOff('tunnel_stopped')
    }
  }, [])

  function handleForwardURL(e) {
    if (e.key === 'Enter' || e.type === 'blur') {
      go?.SetForwardURL(forwardURL)
    }
  }

  async function handleTunnel() {
    if (!go) return
    if (tunnelURL) {
      await go.StopTunnel()
      setTunnelURL('')
      return
    }
    setTunnelLoading(true)
    setTunnelError('')
    try {
      const url = await go.StartTunnel()
      setTunnelURL(url)
    } catch (e) {
      setTunnelError(String(e))
    } finally {
      setTunnelLoading(false)
    }
  }

  async function handleCopyCurl(id) {
    if (!go) return
    const curl = await go.ExportAsCurl(id)
    await navigator.clipboard.writeText(curl)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  async function handleReplay(id, targetURL) {
    if (!go) throw new Error('Wails not available')
    await go.ReplayRequest(id, targetURL)
  }

  function handleCopyTunnelURL() {
    navigator.clipboard.writeText(tunnelURL)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="flex flex-col h-screen bg-[#0d1117] text-[#e6edf3]">
      {/* toolbar */}
      <div className="flex items-center gap-3 px-4 py-2 border-b border-[#30363d] bg-[#161b22] shrink-0 flex-wrap">
        {/* tunnel */}
        <button
          onClick={handleTunnel}
          disabled={tunnelLoading}
          className={`px-3 py-1 text-xs rounded border transition-colors ${
            tunnelURL
              ? 'bg-red-900/40 border-red-700 text-red-300 hover:bg-red-900/60'
              : 'bg-[#1f2937] border-[#30363d] text-gray-300 hover:bg-[#2d3748]'
          } disabled:opacity-50`}
        >
          {tunnelLoading ? 'Starting…' : tunnelURL ? 'Stop Tunnel' : 'Start Tunnel'}
        </button>

        {tunnelURL && (
          <button
            onClick={handleCopyTunnelURL}
            className="flex items-center gap-1 text-xs text-blue-400 hover:text-blue-300 truncate max-w-xs"
            title="Click to copy"
          >
            <span className="w-2 h-2 rounded-full bg-green-500 shrink-0"></span>
            {tunnelURL}
          </button>
        )}

        {tunnelError && (
          <span className="text-xs text-red-400 max-w-xs truncate" title={tunnelError}>
            {tunnelError}
          </span>
        )}

        {/* forward */}
        <div className="flex items-center gap-2 ml-auto">
          <span className="text-xs text-gray-500">Forward →</span>
          <input
            type="text"
            value={forwardURL}
            onChange={e => setForwardURL(e.target.value)}
            onKeyDown={handleForwardURL}
            onBlur={handleForwardURL}
            placeholder="http://localhost:3000"
            className="w-56 bg-[#0d1117] border border-[#30363d] rounded px-2 py-1 text-xs text-gray-200 outline-none focus:border-blue-500"
          />
        </div>

        {copied && <span className="text-xs text-green-400">Copied!</span>}
        {requests.length > 0 && !copied && (
          <button
            onClick={() => { setRequests([]); setSelected(null) }}
            className="text-xs text-gray-600 hover:text-gray-400"
          >
            Clear
          </button>
        )}
      </div>

      {/* main */}
      <div className="flex flex-1 overflow-hidden">
        <div className="w-80 shrink-0 border-r border-[#30363d] overflow-y-auto bg-[#0d1117]">
          <RequestList
            requests={requests}
            selectedId={selected?.id}
            onSelect={setSelected}
          />
        </div>
        <div className="flex-1 overflow-hidden bg-[#0d1117]">
          <RequestDetail
            request={selected}
            onReplay={handleReplay}
            onCopyCurl={handleCopyCurl}
          />
        </div>
      </div>
    </div>
  )
}
