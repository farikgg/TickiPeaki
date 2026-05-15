import { useState, useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { getFlights } from '../api/client.js'
import FlightCard from '../components/FlightCard.jsx'
import Loader from '../components/Loader.jsx'

const LIMIT = 6

export default function FlightsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const navigate = useNavigate()

  const [origin, setOrigin] = useState(searchParams.get('origin') || '')
  const [destination, setDestination] = useState(searchParams.get('destination') || '')
  const [carrier, setCarrier] = useState(searchParams.get('carrier') || '')

  const page = Number(searchParams.get('page') || 1)

  const [flights, setFlights] = useState([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    let active = true
    setLoading(true)
    setError(null)

    const params = { page, limit: LIMIT }
    const o = searchParams.get('origin')
    const d = searchParams.get('destination')
    const c = searchParams.get('carrier')
    if (o) params.origin = o
    if (d) params.destination = d
    if (c) params.carrier = c

    getFlights(params)
      .then((res) => {
        if (!active) return
        const data = res.data
        const list = Array.isArray(data)
          ? data
          : (data.items || data.flights || data.data || [])
        const t = data.total ?? data.count ?? list.length
        setFlights(list)
        setTotal(t)
      })
      .catch((err) => {
        if (!active) return
        setError(err.response?.data?.message || err.message || 'Ошибка загрузки рейсов')
        setFlights([])
        setTotal(0)
      })
      .finally(() => {
        if (active) setLoading(false)
      })

    return () => { active = false }
  }, [searchParams])

  const handleSearch = (e) => {
    e.preventDefault()
    const next = new URLSearchParams()
    if (origin) next.set('origin', origin)
    if (destination) next.set('destination', destination)
    if (carrier) next.set('carrier', carrier)
    next.set('page', '1')
    setSearchParams(next)
  }

  const resetFilters = () => {
    setOrigin('')
    setDestination('')
    setCarrier('')
    setSearchParams({})
  }

  const totalPages = Math.max(1, Math.ceil(total / LIMIT))
  const goPage = (p) => {
    const next = new URLSearchParams(searchParams)
    next.set('page', String(p))
    setSearchParams(next)
  }

  const inputCls = 'flex-1 px-4 py-2.5 bg-white border border-gray-200 focus:border-[#FF6D00] focus:ring-2 focus:ring-orange-100 rounded-xl outline-none transition'

  return (
    <div>
      <form
        onSubmit={handleSearch}
        className="sticky top-16 z-40 bg-white rounded-2xl shadow-sm p-4 mb-6 flex flex-col md:flex-row gap-3 border border-gray-100"
      >
        <input
          value={origin}
          onChange={(e) => setOrigin(e.target.value.toUpperCase())}
          placeholder="Откуда"
          className={inputCls}
        />
        <input
          value={destination}
          onChange={(e) => setDestination(e.target.value.toUpperCase())}
          placeholder="Куда"
          className={inputCls}
        />
        <input
          value={carrier}
          onChange={(e) => setCarrier(e.target.value)}
          placeholder="Авиакомпания"
          className={inputCls}
        />
        <button
          type="submit"
          className="bg-[#FF6D00] hover:bg-orange-600 text-white font-semibold px-6 py-2.5 rounded-xl transition"
        >
          Найти
        </button>
      </form>

      {!loading && !error && (
        <div className="text-gray-500 mb-4">
          {total === 0
            ? 'Рейсы не найдены'
            : `${total} ${total === 1 ? 'рейс' : 'рейсов'} найдено`}
        </div>
      )}

      {loading && <Loader />}

      {!loading && error && (
        <div className="bg-rose-50 text-rose-600 rounded-2xl p-6 text-center">
          {error}
        </div>
      )}

      {!loading && !error && flights.length === 0 && (
        <div className="text-center py-16">
          <div className="text-7xl text-gray-200 mb-4">✈</div>
          <div className="text-lg text-gray-500 mb-4">
            Рейсы по вашему запросу не найдены
          </div>
          <button
            onClick={resetFilters}
            className="text-[#FF6D00] font-semibold hover:underline"
          >
            Сбросить фильтры
          </button>
        </div>
      )}

      {!loading && !error && flights.length > 0 && (
        <>
          <div className="grid grid-cols-1 gap-4">
            {flights.map((f) => (
              <FlightCard
                key={f.id}
                flight={f}
                onSelect={() => navigate(`/flights/${f.id}`)}
              />
            ))}
          </div>

          <div className="flex justify-center items-center gap-6 mt-10">
            <button
              onClick={() => goPage(page - 1)}
              disabled={page <= 1}
              className="text-[#FF6D00] font-semibold disabled:opacity-30 hover:underline disabled:no-underline"
            >
              ← Пред.
            </button>
            <span className="text-gray-500 text-sm">
              {page} / {totalPages}
            </span>
            <button
              onClick={() => goPage(page + 1)}
              disabled={page >= totalPages}
              className="text-[#FF6D00] font-semibold disabled:opacity-30 hover:underline disabled:no-underline"
            >
              След. →
            </button>
          </div>
        </>
      )}
    </div>
  )
}
