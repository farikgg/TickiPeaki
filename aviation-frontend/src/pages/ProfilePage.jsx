import { useState, useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { getTickets, createPassengerProfile, payTicket } from '../api/client.js'
import TicketCard from '../components/TicketCard.jsx'
import Loader from '../components/Loader.jsx'
import { useAuth } from '../context/AuthContext.jsx'

const TABS = [
  { key: 'all', label: 'Все' },
  { key: 'reserved', label: 'Ожидает' },
  { key: 'paid', label: 'Оплачен' },
  { key: 'cancelled', label: 'Отменён' },
]

const inputCls = 'w-full px-4 py-3 bg-white border border-gray-200 focus:border-[#FF6D00] focus:ring-2 focus:ring-orange-100 rounded-xl outline-none transition'

const classLabel = (cls) =>
  cls === 'first' ? 'Первый класс'
  : cls === 'business' ? 'Бизнес'
  : cls === 'economy' ? 'Эконом'
  : (cls || '—')

export default function ProfilePage() {
  const { user, logout, passenger, fetchMe } = useAuth()
  const navigate = useNavigate()

  const [tickets, setTickets] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [tab, setTab] = useState('all')

  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({
    full_name: '',
    email: '',
    phone: '',
    passport_num: '',
  })
  const [submitting, setSubmitting] = useState(false)
  const [formError, setFormError] = useState(null)

  const [payingTicket, setPayingTicket] = useState(null)
  const [payLoading, setPayLoading] = useState(false)
  const [payError, setPayError] = useState(null)

  const [tick, setTick] = useState(0)

  const passengerId = passenger?.id

  const fetchTickets = () => {
    if (!passengerId) {
      setTickets([])
      return Promise.resolve()
    }
    const params = { limit: 100, passenger_id: passengerId }

    return getTickets(params)
      .then((res) => {
        const data = res.data
        const list = Array.isArray(data)
          ? data
          : (data.items || data.tickets || data.data || [])
        setTickets(list)
      })
      .catch((err) => {
        setError(err.response?.data?.error || err.response?.data?.message || err.message || 'Не удалось загрузить билеты')
        setTickets([])
      })
  }

  useEffect(() => {
    let active = true
    setLoading(true)
    setError(null)

    if (!passengerId) {
      setTickets([])
      setLoading(false)
      return () => { active = false }
    }

    const params = { limit: 100, passenger_id: passengerId }

    getTickets(params)
      .then((res) => {
        if (!active) return
        const data = res.data
        const list = Array.isArray(data)
          ? data
          : (data.items || data.tickets || data.data || [])
        setTickets(list)
      })
      .catch((err) => {
        if (!active) return
        setError(err.response?.data?.error || err.response?.data?.message || err.message || 'Не удалось загрузить билеты')
        setTickets([])
      })
      .finally(() => { if (active) setLoading(false) })

    return () => { active = false }
  }, [passengerId])

  const hasPending = tickets.some(t => t.status === 'paid' && !t.pdf_url)

  useEffect(() => {
    if (!hasPending) return

    const interval = setInterval(() => {
      fetchTickets()
    }, 5000)

    return () => clearInterval(interval)
  }, [tickets])

  useEffect(() => {
    const hasReserved = tickets.some(t => t.status === 'reserved')
    if (!hasReserved) return
    const interval = setInterval(() => {
      setTick(t => t + 1)
    }, 1000)
    return () => clearInterval(interval)
  }, [tickets])

  const handleLogout = () => {
    logout()
    navigate('/')
  }

  const change = (field) => (e) =>
    setForm((p) => ({ ...p, [field]: e.target.value }))

  const handleProfileSubmit = async (e) => {
    e.preventDefault()
    setSubmitting(true)
    setFormError(null)
    try {
      await createPassengerProfile(form)
      await fetchMe()
      setShowForm(false)
      setForm({ full_name: '', email: '', phone: '', passport_num: '' })
    } catch (err) {
      const status = err.response?.status
      if (status === 409) {
        setFormError('Профиль уже заполнен')
        await fetchMe()
      } else if (status === 422) {
        setFormError(err.response?.data?.error || 'Проверьте корректность данных')
      } else {
        setFormError(err.response?.data?.error || err.message || 'Не удалось сохранить профиль')
      }
    } finally {
      setSubmitting(false)
    }
  }

  const handlePay = async () => {
    if (!payingTicket) return
    setPayLoading(true)
    setPayError(null)
    try {
      await payTicket(payingTicket.id)
      setPayingTicket(null)
      await fetchTickets()
    } catch (err) {
      setPayError(err.response?.data?.error || err.response?.data?.message || err.message || 'Не удалось оплатить билет')
    } finally {
      setPayLoading(false)
    }
  }

  const closePayModal = () => {
    if (payLoading) return
    setPayingTicket(null)
    setPayError(null)
  }

  const visible = tab === 'all'
    ? tickets
    : tickets.filter((t) => t.status === tab)

  const initials = (user?.username || '?').slice(0, 2).toUpperCase()

  return (
    <div>
      <div className="flex items-center justify-between mb-6 flex-wrap gap-3">
        <div className="flex items-center gap-3 flex-wrap">
          <h1 className="text-3xl font-bold text-gray-900">Личный кабинет</h1>
          <span className="text-sm px-3 py-1 rounded-full bg-orange-50 text-[#FF6D00] font-semibold">
            {user?.username || 'Пользователь'}
          </span>
        </div>
        <button
          onClick={handleLogout}
          className="text-sm text-gray-400 hover:text-red-500 transition"
        >
          Выйти
        </button>
      </div>

      <div className="bg-white rounded-2xl shadow-sm p-6 mb-6 flex items-center gap-4">
        <div className="w-16 h-16 rounded-full bg-[#FF6D00] text-white text-2xl font-bold flex items-center justify-center shrink-0">
          {initials}
        </div>
        <div className="flex-1">
          <div className="text-lg font-bold text-gray-900">
            {user?.username || 'Пользователь'}
          </div>
          <div className="flex items-center gap-2 flex-wrap mt-1">
            {user?.role && (
              <span className="text-xs px-2 py-0.5 rounded-full bg-blue-50 text-blue-600 uppercase font-medium">
                {user.role}
              </span>
            )}
            {passenger ? (
              <span className="text-xs px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-600 font-medium">
                ID пассажира: {passenger.id}
              </span>
            ) : (
              <span className="text-xs px-2 py-0.5 rounded-full bg-amber-50 text-amber-600 font-medium">
                Профиль не заполнен
              </span>
            )}
          </div>
        </div>
      </div>

      {!passenger ? (
        <div className="bg-orange-50 border border-orange-200 rounded-2xl p-6 mb-6">
          <h3 className="font-bold text-gray-900 text-lg mb-1">
            ✈️ Заполните профиль пассажира
          </h3>
          <p className="text-gray-500 text-sm mb-4">
            Это необходимо для покупки билетов
          </p>
          {!showForm && (
            <button
              onClick={() => setShowForm(true)}
              className="bg-[#FF6D00] text-white rounded-xl px-6 py-2 font-semibold hover:bg-orange-600 transition"
            >
              Заполнить →
            </button>
          )}
        </div>
      ) : (
        <div className="bg-white rounded-2xl shadow-sm p-6 border border-gray-100 mb-6 relative">
          <span className="absolute top-4 right-4 text-xs font-semibold text-emerald-600 bg-emerald-50 px-3 py-1 rounded-full">
            ✅ Профиль заполнен
          </span>
          <h3 className="text-lg font-bold text-gray-900 mb-4">Данные пассажира</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            <div>
              <div className="text-gray-400">ФИО</div>
              <div className="text-gray-900 font-medium">{passenger.full_name}</div>
            </div>
            <div>
              <div className="text-gray-400">Email</div>
              <div className="text-gray-900 font-medium">{passenger.email}</div>
            </div>
            <div>
              <div className="text-gray-400">Телефон</div>
              <div className="text-gray-900 font-medium">{passenger.phone}</div>
            </div>
            <div>
              <div className="text-gray-400">Номер паспорта</div>
              <div className="text-gray-900 font-medium">{passenger.passport_num}</div>
            </div>
          </div>
        </div>
      )}

      {showForm && !passenger && (
        <form
          onSubmit={handleProfileSubmit}
          className="bg-white rounded-2xl shadow-sm p-6 mb-6 space-y-3 border border-gray-100"
        >
          <h3 className="text-lg font-bold text-gray-900 mb-2">Данные пассажира</h3>

          <div>
            <label className="block text-sm text-gray-500 mb-1">ФИО</label>
            <input
              required
              value={form.full_name}
              onChange={change('full_name')}
              placeholder="Иван Иванов"
              className={inputCls}
            />
          </div>

          <div>
            <label className="block text-sm text-gray-500 mb-1">Email</label>
            <input
              required
              type="email"
              value={form.email}
              onChange={change('email')}
              placeholder="example@mail.com"
              className={inputCls}
            />
          </div>

          <div>
            <label className="block text-sm text-gray-500 mb-1">Телефон</label>
            <input
              required
              value={form.phone}
              onChange={change('phone')}
              placeholder="+7 701 123 45 67"
              className={inputCls}
            />
          </div>

          <div>
            <label className="block text-sm text-gray-500 mb-1">Номер паспорта</label>
            <input
              required
              value={form.passport_num}
              onChange={change('passport_num')}
              placeholder="N12345678"
              className={inputCls}
            />
          </div>

          <div className="flex gap-3 pt-2">
            <button
              type="submit"
              disabled={submitting}
              className="bg-[#FF6D00] hover:bg-orange-600 text-white font-bold py-3 px-6 rounded-xl transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {submitting ? (
                <>
                  <span className="inline-block w-5 h-5 border-2 border-white/40 border-t-white rounded-full animate-spin"></span>
                  Сохраняем...
                </>
              ) : (
                'Сохранить профиль'
              )}
            </button>
            <button
              type="button"
              onClick={() => { setShowForm(false); setFormError(null) }}
              className="text-gray-500 hover:text-gray-700 transition font-medium"
            >
              Отмена
            </button>
          </div>

          {formError && (
            <div className="text-rose-500 text-sm">{formError}</div>
          )}
        </form>
      )}

      <h2 className="text-xl font-bold text-gray-900 mb-4">Мои билеты</h2>

      {!passenger ? (
        <div className="bg-white rounded-2xl shadow-sm p-10 text-center">
          <div className="text-6xl text-gray-200 mb-4">✈</div>
          <div className="text-gray-500">Сначала заполните профиль пассажира</div>
        </div>
      ) : (
        <>
          {hasPending && (
            <div className="bg-amber-50 border border-amber-100 rounded-xl px-4 py-2 text-amber-600 text-sm flex items-center gap-2 mb-4">
              <span className="animate-spin">⏳</span>
              PDF билеты генерируются, страница обновляется автоматически...
            </div>
          )}

          <div className="flex gap-2 mb-6 flex-wrap">
            {TABS.map((t) => (
              <button
                key={t.key}
                onClick={() => setTab(t.key)}
                className={
                  tab === t.key
                    ? 'bg-[#FF6D00] text-white rounded-lg px-4 py-2 text-sm font-semibold transition'
                    : 'text-gray-500 hover:text-[#FF6D00] rounded-lg px-4 py-2 text-sm transition'
                }
              >
                {t.label}
              </button>
            ))}
          </div>

          {loading && <Loader />}

          {!loading && error && (
            <div className="bg-rose-50 text-rose-600 rounded-2xl p-6 text-center">
              {error}
            </div>
          )}

          {!loading && !error && visible.length === 0 && (
            <div className="bg-white rounded-2xl shadow-sm p-10 text-center">
              <div className="text-6xl text-gray-200 mb-4">✈</div>
              <div className="text-gray-500 mb-4">У вас пока нет билетов</div>
              <Link
                to="/flights"
                className="text-[#FF6D00] font-semibold hover:underline"
              >
                Найти рейс →
              </Link>
            </div>
          )}

          {!loading && !error && visible.length > 0 && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {visible.map((t) => (
                <TicketCard
                  key={`${t.id}-${tick}`}
                  ticket={t}
                  onPay={(id) => {
                    const target = tickets.find((x) => x.id === id)
                    if (target) {
                      setPayError(null)
                      setPayingTicket(target)
                    }
                  }}
                />
              ))}
            </div>
          )}
        </>
      )}

      {payingTicket && (
        <div
          className="fixed inset-0 bg-black/40 backdrop-blur-sm z-50 flex items-center justify-center"
          onClick={closePayModal}
        >
          <div
            className="bg-white rounded-2xl shadow-xl p-8 max-w-md w-full mx-4"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="text-xl font-bold text-gray-900 mb-2">
              Подтверждение оплаты
            </h2>
            <p className="text-gray-500 text-sm mb-6">
              Вы оплачиваете билет на рейс {payingTicket.flight?.flight_number || `#${payingTicket.flight_id}`}
            </p>

            <div className="bg-gray-50 rounded-xl p-4 mb-6 space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-500">Маршрут</span>
                <span className="text-gray-900 font-medium">
                  {(payingTicket.flight?.origin || '???')} → {(payingTicket.flight?.destination || '???')}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">Место</span>
                <span className="text-gray-900 font-medium">
                  {payingTicket.seat?.seat_number || '—'} · {classLabel(payingTicket.seat?.class)}
                </span>
              </div>
              <div className="flex justify-between items-center pt-2 border-t border-gray-200">
                <span className="text-gray-500">Сумма</span>
                <span className="text-[#FF6D00] font-bold text-xl">
                  {payingTicket.seat?.price != null
                    ? `${payingTicket.seat.price.toLocaleString()} ₸`
                    : '0 ₸'}
                </span>
              </div>
            </div>

            <p className="text-xs text-gray-400 mb-6">
              После подтверждения билет считается оплаченным. PDF-билет будет
              сгенерирован автоматически.
            </p>

            {payError && (
              <div className="text-rose-500 text-sm mb-4 text-center">
                {payError}
              </div>
            )}

            <div className="flex gap-3">
              <button
                type="button"
                onClick={closePayModal}
                disabled={payLoading}
                className="flex-1 border border-gray-200 text-gray-600 hover:bg-gray-50 font-semibold py-3 rounded-xl transition disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Отмена
              </button>
              <button
                type="button"
                onClick={handlePay}
                disabled={payLoading}
                className="flex-1 bg-[#FF6D00] hover:bg-orange-600 text-white font-bold py-3 rounded-xl transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
              >
                {payLoading ? (
                  <>
                    <span className="inline-block w-5 h-5 border-2 border-white/40 border-t-white rounded-full animate-spin"></span>
                    Оплачиваем...
                  </>
                ) : (
                  <>
                    Оплатить {payingTicket.seat?.price != null
                      ? `${payingTicket.seat.price.toLocaleString()} ₸`
                      : ''}
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
