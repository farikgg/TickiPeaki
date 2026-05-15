export default function TicketCard({ ticket }) {
  const statusMap = {
    reserved: { label: 'Ожидает', cls: 'bg-amber-100 text-amber-700' },
    paid: { label: 'Оплачен', cls: 'bg-emerald-100 text-emerald-700' },
    cancelled: { label: 'Отменён', cls: 'bg-rose-100 text-rose-600' },
  }
  const status = statusMap[ticket.status] || {
    label: ticket.status || '—',
    cls: 'bg-gray-100 text-gray-600'
  }

  const dep = ticket.departure_time ? new Date(ticket.departure_time) : null
  const fmtDate = (d) => d
    ? d.toLocaleDateString('ru-RU', { day: '2-digit', month: 'long', year: 'numeric' })
    : '—'

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-5 flex">
      <div className="flex-1 pr-4">
        <div className="text-lg font-bold text-[#1A1A1A]">
          {ticket.origin || '???'} → {ticket.destination || '???'}
        </div>
        <div className="text-sm text-gray-500 mt-0.5">
          {ticket.carrier ? `${ticket.carrier} · ` : ''}
          {ticket.flight_number || `Рейс #${ticket.flight_id}`}
        </div>

        <div className="text-sm text-gray-600 mt-3">
          <span className="text-gray-400">Пассажир: </span>
          <span className="text-gray-900">{ticket.passenger_name || '—'}</span>
        </div>
        <div className="text-sm text-gray-600 mt-1 flex items-center gap-2 flex-wrap">
          <span>
            <span className="text-gray-400">Место: </span>
            <span className="text-gray-900">{ticket.seat_number || '—'}</span>
          </span>
          <span className="inline-block text-xs px-2 py-0.5 rounded bg-blue-50 text-blue-600 uppercase font-medium">
            {ticket.class || 'economy'}
          </span>
        </div>
      </div>

      <div className="border-l border-dashed border-gray-200 mx-2"></div>

      <div className="w-32 flex flex-col justify-between text-right">
        <div>
          <div className="text-lg font-bold text-[#FF6D00]">
            {ticket.price ?? 0} ₸
          </div>
          <span className={`inline-block mt-2 text-xs px-2.5 py-1 rounded-full font-medium ${status.cls}`}>
            {status.label}
          </span>
        </div>
        <div className="text-xs text-gray-400 mt-3">{fmtDate(dep)}</div>
      </div>
    </div>
  )
}
