import { useState, useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { getTickets } from '../api/client.js'
import TicketCard from '../components/TicketCard.jsx'
import Loader from '../components/Loader.jsx'
import { useAuth } from '../context/AuthContext.jsx'

const TABS = [
  { key: 'all', label: 'Все' },
  { key: 'reserved', label: 'Ожидает' },
  { key: 'paid', label: 'Оплачен' },
  { key: 'cancelled', label: 'Отменён' },
]

export default function ProfilePage() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  const [tickets, setTickets] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [tab, setTab] = useState('all')

  useEffect(() => {
    let active = true
    setLoading(true)
    setError(null)

    const params = { limit: 100 }
    if (user?.id) params.passenger_id = user.id

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
        setError(err.response?.data?.message || err.message || 'Не удалось загрузить билеты')
        setTickets([])
      })
      .finally(() => { if (active) setLoading(false) })

    return () => { active = false }
  }, [user])

  const handleLogout = () => {
    logout()
    navigate('/')
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
        <div>
          <div className="text-lg font-bold text-gray-900">
            {user?.username || 'Пользователь'}
          </div>
          {user?.role ? (
            <span className="inline-block mt-1 text-xs px-2 py-0.5 rounded-full bg-blue-50 text-blue-600 uppercase font-medium">
              {user.role}
            </span>
          ) : (
            <div className="text-sm text-gray-500">Пользователь SkyBook</div>
          )}
        </div>
      </div>

      <h2 className="text-xl font-bold text-gray-900 mb-4">Мои билеты</h2>

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
            <TicketCard key={t.id} ticket={t} />
          ))}
        </div>
      )}
    </div>
  )
}
