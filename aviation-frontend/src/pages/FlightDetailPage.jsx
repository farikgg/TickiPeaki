import { useState, useEffect } from 'react'
import { useParams, useNavigate, useLocation } from 'react-router-dom'
import { getFlightById, createTicket } from '../api/client.js'
import Loader from '../components/Loader.jsx'
import { useAuth } from '../context/AuthContext.jsx'

const classLabel = (cls) =>
  cls === 'first' ? 'Первый класс'
  : cls === 'business' ? 'Бизнес'
  : 'Эконом'

const classColors = {
  first: {
    available: 'bg-violet-100 border border-violet-300 text-violet-700 hover:bg-violet-200',
    selected: 'bg-violet-500 border-violet-600 text-white',
  },
  business: {
    available: 'bg-sky-100 border border-sky-300 text-sky-700 hover:bg-sky-200',
    selected: 'bg-sky-500 border-sky-600 text-white',
  },
  economy: {
    available: 'bg-emerald-100 border border-emerald-300 text-emerald-700 hover:bg-emerald-200',
    selected: 'bg-[#FF6D00] border-orange-400 text-white',
  },
}

const BOOKED_CLS = 'bg-gray-200 border border-gray-300 text-gray-400 cursor-not-allowed opacity-60'

function SeatMap({ seats, takenSeats, selectedSeat, onSeatSelect }) {
  const parseSeats = (list) => {
    const rows = {}
    list.forEach((seat) => {
      const match = seat.seat_number.match(/^(\d+)([A-F])$/)
      if (!match) return
      const row = parseInt(match[1], 10)
      const col = match[2]
      if (!rows[row]) rows[row] = {}
      rows[row][col] = seat
    })
    return rows
  }

  const rowsMap = parseSeats(seats)
  const rowNumbers = Object.keys(rowsMap)
    .map((n) => parseInt(n, 10))
    .sort((a, b) => a - b)

  const allCols = new Set()
  seats.forEach((s) => {
    const m = s.seat_number.match(/^(\d+)([A-F])$/)
    if (m) allCols.add(m[2])
  })
  const sortedCols = Array.from(allCols).sort()

  const leftCols = sortedCols.filter((c) => c <= 'C')
  const rightCols = sortedCols.filter((c) => c >= 'D')
  const headerCols = [...leftCols, '', ...rightCols]

  return (
    <div className="bg-white rounded-2xl border border-gray-100 shadow-sm p-6">
      <div className="flex flex-wrap gap-4 mb-6 text-xs text-gray-600">
        <span className="flex items-center gap-1.5">
          <span className="w-5 h-5 rounded bg-violet-100 border border-violet-300 inline-block" />
          Первый
        </span>
        <span className="flex items-center gap-1.5">
          <span className="w-5 h-5 rounded bg-sky-100 border border-sky-300 inline-block" />
          Бизнес
        </span>
        <span className="flex items-center gap-1.5">
          <span className="w-5 h-5 rounded bg-emerald-100 border border-emerald-300 inline-block" />
          Эконом
        </span>
        <span className="flex items-center gap-1.5">
          <span className="w-5 h-5 rounded bg-gray-200 border border-gray-300 inline-block opacity-60" />
          Занято
        </span>
        <span className="flex items-center gap-1.5">
          <span className="w-5 h-5 rounded bg-[#FF6D00] inline-block" />
          Выбрано
        </span>
      </div>

      <div className="text-center text-2xl mb-4">✈️</div>

      <div className="flex justify-center mb-2">
        <div className="w-8 text-center text-xs text-gray-400 mr-2" />
        {headerCols.map((col, i) => (
          <div
            key={i}
            className={`text-center text-xs font-semibold text-gray-500 ${col === '' ? 'w-4' : 'w-8'}`}
          >
            {col}
          </div>
        ))}
      </div>

      <div className="flex flex-col gap-1 overflow-y-auto max-h-80">
        {rowNumbers.map((row) => {
          const rowSeats = rowsMap[row]
          return (
            <div key={row} className="flex justify-center items-center gap-1">
              <div className="w-8 text-center text-xs text-gray-400 mr-1">{row}</div>
              {headerCols.map((col, i) => {
                if (col === '') return <div key={`gap-${i}`} className="w-4" />
                const seat = rowSeats[col]
                if (!seat) {
                  return <div key={`empty-${row}-${col}`} className="w-8 h-8" />
                }
                const isBooked =
                  seat.status === 'booked' ||
                  (Array.isArray(takenSeats) && takenSeats.includes(seat.seat_number))
                const isSelected = selectedSeat?.id === seat.id
                const colors = classColors[seat.class] || classColors.economy

                let stateCls
                if (isBooked) stateCls = BOOKED_CLS
                else if (isSelected) stateCls = colors.selected
                else stateCls = colors.available

                return (
                  <button
                    key={seat.id}
                    type="button"
                    onClick={() => {
                      if (isBooked) return
                      onSeatSelect(seat)
                    }}
                    disabled={isBooked}
                    className={`w-8 h-8 rounded text-xs font-medium transition-colors ${stateCls}`}
                  >
                    {col}
                  </button>
                )
              })}
            </div>
          )
        })}
      </div>
    </div>
  )
}

export default function FlightDetailPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const location = useLocation()
  const { isAuthenticated } = useAuth()

  const [flight, setFlight] = useState(null)
  const [seats, setSeats] = useState([])
  const [takenSeats, setTakenSeats] = useState([])
  const [availableCount, setAvailableCount] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const [selectedSeat, setSelectedSeat] = useState(null)
  const [submitting, setSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState(null)
  const [profileMissing, setProfileMissing] = useState(false)
  const [success, setSuccess] = useState(false)

  useEffect(() => {
    let active = true
    setLoading(true)
    setError(null)

    getFlightById(id)
      .then((res) => {
        if (!active) return
        const data = res.data
        const flightData = data?.flight || data?.data || data
        setFlight(flightData)
        setSeats(Array.isArray(data?.seats) ? data.seats : [])
        setTakenSeats(Array.isArray(data?.taken_seats) ? data.taken_seats : [])
        setAvailableCount(
          typeof data?.available_count === 'number' ? data.available_count : 0
        )
      })
      .catch((err) => {
        if (!active) return
        setError(
          err.response?.data?.error ||
            err.response?.data?.message ||
            err.message ||
            'Не удалось загрузить рейс'
        )
      })
      .finally(() => {
        if (active) setLoading(false)
      })

    return () => {
      active = false
    }
  }, [id])

  const handleSeatSelect = (seat) => {
    setSelectedSeat(seat)
  }

  const handleSubmit = async (e) => {
    e.preventDefault()

    if (!selectedSeat) {
      setSubmitError('Выберите место на схеме')
      return
    }

    setSubmitting(true)
    setSubmitError(null)
    setProfileMissing(false)
    try {
      await createTicket({
        flight_id: Number(id),
        seat_id: selectedSeat.id,
      })
      setSeats((prev) =>
        prev.map((s) =>
          s.id === selectedSeat.id ? { ...s, status: 'booked' } : s
        )
      )
      setTakenSeats((prev) =>
        prev.includes(selectedSeat.seat_number)
          ? prev
          : [...prev, selectedSeat.seat_number]
      )
      setAvailableCount((c) => Math.max(0, c - 1))
      setSelectedSeat(null)
      setSuccess(true)
    } catch (err) {
      if (err.response?.status === 403) {
        setProfileMissing(true)
      } else {
        setSubmitError(
          err.response?.data?.error ||
            err.response?.data?.message ||
            err.message ||
            'Не удалось оформить билет'
        )
      }
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return <Loader />
  if (error) {
    return (
      <div className="bg-rose-50 text-rose-600 rounded-2xl p-6 text-center">
        {error}
      </div>
    )
  }
  if (!flight) return null

  const seatBadge = availableCount > 5
    ? { cls: 'bg-emerald-50 text-emerald-600', text: `${availableCount} мест свободно` }
    : availableCount > 0
    ? { cls: 'bg-amber-50 text-amber-600', text: `Осталось ${availableCount}` }
    : { cls: 'bg-rose-50 text-rose-500', text: 'Мест нет' }

  const dep = flight.departure_time ? new Date(flight.departure_time) : null
  const arr = flight.arrival_time ? new Date(flight.arrival_time) : null
  const fmt = (d) => d
    ? d.toLocaleString('ru-RU', {
        day: '2-digit', month: 'long', year: 'numeric',
        hour: '2-digit', minute: '2-digit'
      })
    : '—'
  let durationStr = '—'
  if (dep && arr) {
    const mins = Math.max(0, Math.round((arr - dep) / 60000))
    const h = Math.floor(mins / 60)
    const m = mins % 60
    durationStr = `${h}ч ${m}м`
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div className="bg-white rounded-2xl shadow-sm p-6 md:p-8">
        <div className="text-sm text-gray-500 mb-4">
          <span className="font-medium text-gray-700">{flight.carrier || 'Carrier'}</span>
          {' · '}
          {flight.flight_number || flight.number}
        </div>

        <div className="flex items-center justify-between mb-8">
          <div className="text-3xl md:text-4xl font-bold text-gray-900">
            {flight.origin}
          </div>
          <div className="flex-1 mx-3 flex flex-col items-center">
            <div className="text-xs text-gray-400 mb-1">{durationStr}</div>
            <div className="w-full flex items-center">
              <div className="flex-1 border-t border-dashed border-gray-300"></div>
              <span className="mx-2 text-[#FF6D00] text-2xl">✈</span>
              <div className="flex-1 border-t border-dashed border-gray-300"></div>
            </div>
            <div className="text-xs text-gray-400 mt-1">Прямой рейс</div>
          </div>
          <div className="text-3xl md:text-4xl font-bold text-gray-900">
            {flight.destination}
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4 mb-6 text-sm">
          <div>
            <div className="text-gray-400">Вылет</div>
            <div className="text-gray-900 font-medium">{fmt(dep)}</div>
          </div>
          <div className="text-right">
            <div className="text-gray-400">Прилёт</div>
            <div className="text-gray-900 font-medium">{fmt(arr)}</div>
          </div>
        </div>

        <div className="flex justify-between items-center pt-6 border-t border-gray-100 flex-wrap gap-3">
          <span className={`text-sm px-3 py-1.5 rounded-full font-medium ${seatBadge.cls}`}>
            {seatBadge.text}
          </span>
        </div>
      </div>

      <div>
        <SeatMap
          seats={seats}
          takenSeats={takenSeats}
          selectedSeat={selectedSeat}
          onSeatSelect={handleSeatSelect}
        />
      </div>

      <div>
        {!isAuthenticated ? (
          <div className="bg-white rounded-2xl shadow-sm p-8 text-center">
            <div className="text-5xl mb-3">🔒</div>
            <h2 className="text-xl font-bold text-gray-900 mb-2">
              Войдите, чтобы купить билет
            </h2>
            <p className="text-gray-500 mb-6">
              Для оформления билета необходим аккаунт
            </p>
            <div className="flex flex-col gap-3">
              <button
                onClick={() => navigate('/login', { state: { from: location.pathname } })}
                className="w-full bg-[#FF6D00] hover:bg-orange-600 text-white font-semibold py-3 rounded-xl transition"
              >
                Войти
              </button>
              <button
                onClick={() => navigate('/register', { state: { from: location.pathname } })}
                className="w-full border border-[#FF6D00] text-[#FF6D00] hover:bg-orange-50 font-semibold py-3 rounded-xl transition"
              >
                Зарегистрироваться
              </button>
            </div>
          </div>
        ) : success ? (
          <div className="bg-white rounded-2xl shadow-sm p-8">
            <div className="text-center py-6">
              <div className="text-5xl mb-3">🎉</div>
              <h3 className="text-xl font-bold text-gray-900 mb-1">Билет оформлен!</h3>
              <p className="text-gray-500 text-sm mb-4">
                PDF билет будет готов в течение нескольких секунд
              </p>
              <div className="bg-amber-50 border border-amber-200 rounded-xl px-4 py-3 text-amber-700 text-sm mb-4">
                🔄 После оплаты билета PDF появится в личном кабинете
              </div>
              <a
                href="/profile"
                className="inline-block bg-[#FF6D00] text-white rounded-xl px-6 py-2 font-semibold hover:bg-orange-600 transition"
              >
                Перейти в личный кабинет →
              </a>
            </div>
          </div>
        ) : (
          <form
            onSubmit={handleSubmit}
            className="bg-white rounded-2xl shadow-sm p-6 space-y-3"
          >
            <h2 className="text-xl font-bold text-gray-900 mb-2">
              Оформление билета
            </h2>

            {selectedSeat ? (
              <div className="bg-gray-50 rounded-xl p-4 mb-4">
                <div className="flex justify-between items-center">
                  <div>
                    <span className="font-bold text-lg">{selectedSeat.seat_number}</span>
                    <span className="ml-2 text-sm text-gray-500 capitalize">
                      {classLabel(selectedSeat.class)}
                    </span>
                  </div>
                  <span className="font-bold text-[#FF6D00] text-xl">
                    {selectedSeat.price.toLocaleString()} ₸
                  </span>
                </div>
              </div>
            ) : (
              <div className="bg-amber-50 border border-amber-200 rounded-xl p-4 mb-4 text-amber-700 text-sm">
                ← Выберите место на схеме слева
              </div>
            )}

            <div className="flex justify-between items-center py-3 border-t border-gray-100">
              <span className="text-gray-600">Итого:</span>
              <span className="font-bold text-[#FF6D00] text-xl">
                {selectedSeat ? `${selectedSeat.price.toLocaleString()} ₸` : '—'}
              </span>
            </div>

            <button
              type="submit"
              disabled={submitting || !selectedSeat}
              className="w-full bg-[#FF6D00] hover:bg-orange-600 text-white font-bold py-3 rounded-xl transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {submitting ? (
                <>
                  <span className="inline-block w-5 h-5 border-2 border-white/40 border-t-white rounded-full animate-spin"></span>
                  Оформляем...
                </>
              ) : (
                'Купить билет'
              )}
            </button>

            {profileMissing && (
              <div className="bg-amber-50 border border-amber-200 rounded-xl p-4">
                <p className="text-amber-700 font-medium">⚠️ Профиль не заполнен</p>
                <p className="text-amber-600 text-sm mt-1">
                  Для покупки билета необходимо заполнить данные пассажира
                </p>
                <a
                  href="/profile"
                  className="mt-3 inline-block bg-[#FF6D00] text-white rounded-xl px-4 py-2 text-sm font-semibold hover:bg-orange-600 transition"
                >
                  Заполнить профиль →
                </a>
              </div>
            )}

            {submitError && !profileMissing && (
              <div className="text-rose-500 text-sm text-center">{submitError}</div>
            )}
          </form>
        )}
      </div>
    </div>
  )
}
