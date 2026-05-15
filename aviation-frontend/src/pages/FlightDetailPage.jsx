import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate, useLocation } from 'react-router-dom'
import { getFlightById, createTicket } from '../api/client.js'
import Loader from '../components/Loader.jsx'
import { useAuth } from '../context/AuthContext.jsx'

export default function FlightDetailPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const location = useLocation()
  const { isAuthenticated, user } = useAuth()

  const [flight, setFlight] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const [form, setForm] = useState({
    passenger_name: '',
    passenger_email: '',
    passport_num: '',
    seat_number: '',
    class: 'economy',
  })
  const [submitting, setSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState(null)
  const [success, setSuccess] = useState(false)

  useEffect(() => {
    let active = true
    setLoading(true)
    setError(null)

    getFlightById(id)
      .then((res) => {
        if (!active) return
        const data = res.data?.flight || res.data?.data || res.data
        setFlight(data)
      })
      .catch((err) => {
        if (!active) return
        setError(err.response?.data?.message || err.message || 'Не удалось загрузить рейс')
      })
      .finally(() => { if (active) setLoading(false) })

    return () => { active = false }
  }, [id])

  const change = (field) => (e) =>
    setForm((p) => ({ ...p, [field]: e.target.value }))

  const handleSubmit = async (e) => {
    e.preventDefault()
    setSubmitting(true)
    setSubmitError(null)
    try {
      await createTicket({
        flight_id: Number(id),
        passenger_id: user?.id,
        passenger_name: form.passenger_name,
        passenger_email: form.passenger_email,
        passport_num: form.passport_num,
        seat_number: form.seat_number,
        class: form.class,
        price: flight?.price,
      })
      setSuccess(true)
    } catch (err) {
      setSubmitError(err.response?.data?.message || err.message || 'Не удалось оформить билет')
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

  const seats = flight.available_seats ?? flight.seats_available ?? 0
  const seatBadge = seats > 5
    ? { cls: 'bg-emerald-50 text-emerald-600', text: `${seats} мест свободно` }
    : seats > 0
    ? { cls: 'bg-amber-50 text-amber-600', text: `Осталось ${seats}` }
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

  const inputCls = 'w-full px-4 py-3 bg-white border border-gray-200 focus:border-[#FF6D00] focus:ring-2 focus:ring-orange-100 rounded-xl outline-none transition'

  return (
    <div className="grid grid-cols-1 md:grid-cols-5 gap-6">
      <div className="md:col-span-3 bg-white rounded-2xl shadow-sm p-6 md:p-8">
        <div className="text-sm text-gray-500 mb-4">
          <span className="font-medium text-gray-700">{flight.carrier || 'Carrier'}</span>
          {' · '}
          {flight.flight_number || flight.number}
        </div>

        <div className="flex items-center justify-between mb-8">
          <div className="text-4xl md:text-5xl font-bold text-gray-900">
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
          <div className="text-4xl md:text-5xl font-bold text-gray-900">
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

        <div className="flex justify-between items-center pt-6 border-t border-gray-100">
          <span className={`text-sm px-3 py-1.5 rounded-full font-medium ${seatBadge.cls}`}>
            {seatBadge.text}
          </span>
          <div className="text-right">
            <div className="text-sm text-gray-500">Цена</div>
            <div className="text-3xl md:text-4xl font-bold text-[#FF6D00]">
              {flight.price ?? 0} ₸
            </div>
          </div>
        </div>
      </div>

      <div className="md:col-span-2">
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
          <div className="bg-white rounded-2xl shadow-sm p-8 text-center">
            <div className="text-5xl text-emerald-500 mb-3">✅</div>
            <h2 className="text-xl font-bold text-gray-900 mb-2">
              Билет оформлен!
            </h2>
            <p className="text-gray-500 mb-6">
              Информация отправлена на ваш email
            </p>
            <Link
              to="/profile"
              className="inline-block w-full bg-[#FF6D00] hover:bg-orange-600 text-white font-semibold py-3 rounded-xl transition"
            >
              Перейти в личный кабинет →
            </Link>
          </div>
        ) : (
          <form
            onSubmit={handleSubmit}
            className="bg-white rounded-2xl shadow-sm p-6 space-y-3"
          >
            <h2 className="text-xl font-bold text-gray-900 mb-2">
              Оформление билета
            </h2>

            <div>
              <label className="block text-sm text-gray-500 mb-1">
                Имя пассажира
              </label>
              <input
                required
                value={form.passenger_name}
                onChange={change('passenger_name')}
                placeholder="Иван Иванов"
                className={inputCls}
              />
            </div>

            <div>
              <label className="block text-sm text-gray-500 mb-1">Email</label>
              <input
                required
                type="email"
                value={form.passenger_email}
                onChange={change('passenger_email')}
                placeholder="example@mail.com"
                className={inputCls}
              />
            </div>

            <div>
              <label className="block text-sm text-gray-500 mb-1">
                Номер паспорта
              </label>
              <input
                required
                value={form.passport_num}
                onChange={change('passport_num')}
                placeholder="N1234567"
                className={inputCls}
              />
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm text-gray-500 mb-1">Место</label>
                <input
                  required
                  value={form.seat_number}
                  onChange={change('seat_number')}
                  placeholder="12A"
                  className={inputCls}
                />
              </div>
              <div>
                <label className="block text-sm text-gray-500 mb-1">Класс</label>
                <select
                  value={form.class}
                  onChange={change('class')}
                  className={inputCls}
                >
                  <option value="economy">Эконом</option>
                  <option value="business">Бизнес</option>
                  <option value="first">Первый</option>
                </select>
              </div>
            </div>

            <div className="flex justify-between items-center pt-3 border-t border-gray-100">
              <span className="text-gray-500">Итого:</span>
              <span className="text-lg font-bold text-[#FF6D00]">
                {flight.price ?? 0} ₸
              </span>
            </div>

            <button
              type="submit"
              disabled={submitting || seats === 0}
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

            {submitError && (
              <div className="text-rose-500 text-sm text-center">{submitError}</div>
            )}
          </form>
        )}
      </div>
    </div>
  )
}
