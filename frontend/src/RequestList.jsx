const METHOD_COLORS = {
  GET:    'text-green-400',
  POST:   'text-blue-400',
  PUT:    'text-yellow-400',
  PATCH:  'text-orange-400',
  DELETE: 'text-red-400',
}

function statusColor(status) {
  if (!status || status === 0) return 'text-gray-500'
  if (status < 300) return 'text-green-400'
  if (status < 400) return 'text-yellow-400'
  return 'text-red-400'
}

function formatTime(ts) {
  if (!ts) return ''
  const d = new Date(ts)
  return d.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

export default function RequestList({ requests, selectedId, onSelect }) {
  return (
    <div className="flex flex-col h-full overflow-y-auto">
      {requests.length === 0 && (
        <div className="flex-1 flex items-center justify-center text-gray-600 text-xs">
          Waiting for requests on :9000
        </div>
      )}
      {requests.map(r => (
        <div
          key={r.id}
          onClick={() => onSelect(r)}
          className={`flex items-center gap-2 px-3 py-2 border-b border-[#30363d] cursor-pointer select-none hover:bg-[#1f2937] transition-colors ${selectedId === r.id ? 'bg-[#1f2937] border-l-2 border-l-blue-500' : ''}`}
        >
          <span className={`w-14 shrink-0 font-bold text-xs ${METHOD_COLORS[r.method] ?? 'text-gray-400'}`}>
            {r.method}
          </span>
          <span className="flex-1 truncate text-gray-200 text-xs">{r.path}</span>
          <span className={`text-xs ${statusColor(r.status)}`}>
            {r.status > 0 ? r.status : '—'}
          </span>
          <span className="text-gray-600 text-xs shrink-0">{formatTime(r.created_at)}</span>
        </div>
      ))}
    </div>
  )
}
