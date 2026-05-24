export default function FlightCard({ flight, onSelect }) {
  const dep = flight.departure_time ? new Date(flight.departure_time) : null
  const arr = flight.arrival_time ? new Date(flight.arrival_time) : null

  let durationStr = '—'
  if (dep && arr) {
    const mins = Math.max(0, Math.round((arr - dep) / 60000))
    const h = Math.floor(mins / 60)
    const m = mins % 60
    durationStr = `${h}ч ${m}м`
  }

  const fmtTime = (d) => d
    ? d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
    : '—'
  const fmtDate = (d) => d
    ? d.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short' })
    : ''

  return (
    <div className="bg-white rounded-2xl shadow-sm hover:shadow-md transition-shadow p-5 border border-gray-100">
      <div className="flex items-center gap-2 mb-4 text-sm text-gray-500">
        <span className="font-medium text-gray-700">{flight.carrier || 'Carrier'}</span>
        <span>·</span>
        <span>{flight.flight_number || flight.number || ''}</span>
      </div>

      <div className="flex items-center justify-between gap-4">
        <div>
          <div className="text-2xl font-bold text-[#1A1A1A]">{flight.origin}</div>
          <div className="text-sm text-gray-500 mt-1">{fmtTime(dep)}</div>
          <div className="text-xs text-gray-400">{fmtDate(dep)}</div>
        </div>

        <div className="flex-1 flex flex-col items-center px-2">
          <div className="text-xs text-gray-400">{durationStr}</div>
          <div className="w-full flex items-center my-1">
            <div className="flex-1 border-t border-dashed border-gray-300"></div>
            <span className="mx-2 text-[#FF6D00] text-lg">✈</span>
            <div className="flex-1 border-t border-dashed border-gray-300"></div>
          </div>
          <div className="text-xs text-gray-400">Прямой</div>
        </div>

        <div className="text-right">
          <div className="text-2xl font-bold text-[#1A1A1A]">{flight.destination}</div>
          <div className="text-sm text-gray-500 mt-1">{fmtTime(arr)}</div>
          <div className="text-xs text-gray-400">{fmtDate(arr)}</div>
        </div>
      </div>

      <div className="border-t border-gray-100 mt-4 pt-4 flex justify-end items-center">
        <button
          onClick={onSelect}
          className="bg-[#FF6D00] text-white rounded-xl px-4 py-2 text-sm font-semibold hover:bg-orange-600 transition"
        >
          Смотреть места →
        </button>
      </div>
    </div>
  )
}
