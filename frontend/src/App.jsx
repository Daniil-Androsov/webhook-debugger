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

  useEffect(() => {
    if (!go) return

    go.GetRequests().then(data => {
      setRequests(data ?? [])
    })

    const runtime = window?.runtime
    if (runtime) {
      runtime.EventsOn('new_request', req => {
        setRequests(prev => [req, ...prev])
      })
      return () => runtime.EventsOff('new_request')
    }
  }, [])

  function handleForwardURL(e) {
    if (e.key === 'Enter' || e.type === 'blur') {
      go?.SetForwardURL(forwardURL)
    }
  }

  async function handleReplay(id, targetURL) {
    if (!go) throw new Error('Wails not available')
    await go.ReplayRequest(id, targetURL)
  }

  async function handleCopyCurl(id) {
    if (!go) return
    const curl = await go.ExportAsCurl(id)
    await navigator.clipboard.writeText(curl)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="flex flex-col h-screen bg-[#0d1117] text-[#e6edf3]">
      {/* toolbar */}
      <div className="flex items-center gap-3 px-4 py-2 border-b border-[#30363d] bg-[#161b22] shrink-0">
        <div className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-green-500"></span>
          <span className="text-xs text-gray-400">:9000</span>
        </div>
        <div className="flex items-center gap-2 ml-4">
          <span className="text-xs text-gray-500">Forward →</span>
          <input
            type="text"
            value={forwardURL}
            onChange={e => setForwardURL(e.target.value)}
            onKeyDown={handleForwardURL}
            onBlur={handleForwardURL}
            placeholder="http://localhost:3000"
            className="w-64 bg-[#0d1117] border border-[#30363d] rounded px-2 py-1 text-xs text-gray-200 outline-none focus:border-blue-500"
          />
        </div>
        {copied && <span className="ml-auto text-xs text-green-400">Copied!</span>}
        {requests.length > 0 && !copied && (
          <button
            onClick={() => { setRequests([]); setSelected(null) }}
            className="ml-auto text-xs text-gray-600 hover:text-gray-400"
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
