export default function TicketCard({ ticket, onPay }) {
  const statusMap = {
    reserved: { label: 'Ожидает', cls: 'bg-amber-100 text-amber-700' },
    paid: { label: 'Оплачен', cls: 'bg-emerald-100 text-emerald-700' },
    cancelled: { label: 'Отменён', cls: 'bg-rose-100 text-rose-600' },
  }
  const status = statusMap[ticket.status] || {
    label: ticket.status || '—',
    cls: 'bg-gray-100 text-gray-600'
  }

  const dep = ticket.flight?.departure_time
    ? new Date(ticket.flight.departure_time)
    : ticket.departure_time ? new Date(ticket.departure_time) : null
  const fmtDate = (d) => d
    ? d.toLocaleDateString('ru-RU', { day: '2-digit', month: 'long', year: 'numeric' })
    : '—'

  const origin = ticket.flight?.origin || '???'
  const destination = ticket.flight?.destination || '???'
  const carrier = ticket.flight?.carrier
  const flightNumber = ticket.flight?.flight_number || `Рейс #${ticket.flight_id}`
  const passengerName = ticket.passenger?.full_name || '—'
  const seatNumber = ticket.seat?.seat_number || '—'
  const seatClass = ticket.seat?.class
  const seatPrice = ticket.seat?.price

  let remaining = null
  if (ticket.status === 'reserved' && ticket.booked_at) {
    const expiresAt = new Date(new Date(ticket.booked_at).getTime() + 5 * 60 * 1000)
    remaining = Math.max(0, Math.floor((expiresAt - Date.now()) / 1000))
  }

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-5 flex">
      <div className="flex-1 pr-4">
        <div className="text-lg font-bold text-[#1A1A1A]">
          {origin} → {destination}
        </div>
        <div className="text-sm text-gray-500 mt-0.5">
          {carrier ? `${carrier} · ` : ''}
          {flightNumber}
        </div>

        <div className="text-sm text-gray-600 mt-3">
          <span className="text-gray-400">Пассажир: </span>
          <span className="text-gray-900">{passengerName}</span>
        </div>
        <div className="text-sm text-gray-600 mt-1 flex items-center gap-2 flex-wrap">
          <span>
            <span className="text-gray-400">Место: </span>
            <span className="text-gray-900">{seatNumber}</span>
          </span>
          {seatClass && (
            <span className="inline-block text-xs px-2 py-0.5 rounded bg-blue-50 text-blue-600 uppercase font-medium">
              {seatClass}
            </span>
          )}
        </div>

        {ticket.status === 'reserved' && remaining !== null && (
          remaining > 0 ? (
            <div className="text-xs text-amber-600 flex items-center gap-1 mt-2">
              ⏱ Оплатите в течение {Math.floor(remaining / 60)}:{String(remaining % 60).padStart(2, '0')}
            </div>
          ) : (
            <div className="text-xs text-rose-500 mt-2">
              Время истекло
            </div>
          )
        )}

        {ticket.status === 'reserved' && remaining > 0 && onPay && (
          <button
            onClick={() => onPay(ticket.id)}
            className="mt-3 w-full bg-[#FF6D00] text-white rounded-xl py-2 text-sm font-semibold hover:bg-orange-600 transition"
          >
            Оплатить билет
          </button>
        )}
      </div>

      <div className="border-l border-dashed border-gray-200 mx-2"></div>

      <div className="w-32 flex flex-col justify-between text-right">
        <div>
          <div className="text-lg font-bold text-[#FF6D00]">
            {seatPrice != null ? `${seatPrice.toLocaleString()} ₸` : '0 ₸'}
          </div>
          <span className={`inline-block mt-2 text-xs px-2.5 py-1 rounded-full font-medium ${status.cls}`}>
            {status.label}
          </span>
          {ticket.status === 'paid' && !ticket.pdf_url && (
            <div className="text-amber-500 text-sm animate-pulse mt-2">
              🔄 PDF генерируется...
            </div>
          )}
          {ticket.status === 'paid' && ticket.pdf_url && (
            <a
              href={ticket.pdf_url}
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm font-medium text-[#FF6D00] hover:underline flex items-center gap-1 justify-end mt-2"
            >
              📄 Скачать билет
            </a>
          )}
        </div>
        <div className="text-xs text-gray-400 mt-3">{fmtDate(dep)}</div>
      </div>
    </div>
  )
}
