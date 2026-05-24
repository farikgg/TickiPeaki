import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

export default function HomePage() {
  const [origin, setOrigin] = useState('')
  const [destination, setDestination] = useState('')
  const navigate = useNavigate()

  const handleSearch = (e) => {
    e.preventDefault()
    const params = new URLSearchParams()
    if (origin) params.set('origin', origin)
    if (destination) params.set('destination', destination)
    navigate(`/flights${params.toString() ? `?${params.toString()}` : ''}`)
  }

  const popular = [
    { from: 'ALA', to: 'NQZ', emoji: '🏙️', label: 'Алматы → Астана', price: 'от 24 900 ₸' },
    { from: 'ALA', to: 'CIT', emoji: '🌇', label: 'Алматы → Шымкент', price: 'от 18 500 ₸' },
    { from: 'NQZ', to: 'ALA', emoji: '⛰️', label: 'Астана → Алматы', price: 'от 25 100 ₸' },
  ]

  const features = [
    { icon: '🔍', title: 'Быстрый поиск', desc: 'Найдите подходящий рейс за пару кликов' },
    { icon: '✅', title: 'Надёжные билеты', desc: 'Прямая интеграция с авиакомпаниями' },
    { icon: '📱', title: 'Удобно везде', desc: 'Работает на любом устройстве' },
  ]

  const inputCls = 'w-full pl-10 pr-4 py-3 bg-white border border-gray-200 focus:border-[#FF6D00] focus:ring-2 focus:ring-orange-100 rounded-xl outline-none transition text-[#1A1A1A] placeholder-gray-400'

  return (
    <div>
      <section className="bg-white rounded-3xl shadow-sm p-8 md:p-12 mb-10">
        <h1 className="text-3xl md:text-4xl font-bold text-gray-900 leading-tight">
          Летите туда, куда мечтаете
        </h1>
        <p className="text-gray-500 mt-3 text-base md:text-lg">
          Поиск лучших билетов по всем направлениям
        </p>

        <form
          onSubmit={handleSearch}
          className="mt-8 flex flex-col md:flex-row gap-3"
        >
          <div className="relative flex-1">
            <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">📍</span>
            <input
              value={origin}
              onChange={(e) => setOrigin(e.target.value.toUpperCase())}
              placeholder="Откуда"
              className={inputCls}
            />
          </div>
          <div className="relative flex-1">
            <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">📍</span>
            <input
              value={destination}
              onChange={(e) => setDestination(e.target.value.toUpperCase())}
              placeholder="Куда"
              className={inputCls}
            />
          </div>
          <button
            type="submit"
            className="bg-[#FF6D00] hover:bg-orange-600 px-8 py-3 rounded-xl text-white font-semibold transition"
          >
            Найти рейсы
          </button>
        </form>
      </section>

      <section className="mb-12">
        <h2 className="text-2xl font-bold text-gray-900 mb-5">Популярные направления</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {popular.map((p) => (
            <button
              key={`${p.from}-${p.to}`}
              onClick={() => navigate(`/flights?origin=${p.from}&destination=${p.to}`)}
              className="bg-white border border-gray-100 rounded-2xl p-5 text-left hover:border-[#FF6D00] hover:shadow-md transition cursor-pointer"
            >
              <div className="text-3xl mb-3">{p.emoji}</div>
              <div className="font-semibold text-gray-900">{p.label}</div>
              <div className="text-sm text-[#FF6D00] mt-1 font-medium">{p.price}</div>
            </button>
          ))}
        </div>
      </section>

      <section className="bg-[#FFF8F5] rounded-3xl p-8 md:p-10 mb-8">
        <h2 className="text-2xl font-bold text-gray-900 mb-8 text-center">
          Почему TickiPeaki?
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {features.map((f) => (
            <div key={f.title} className="text-center">
              <div className="text-4xl mb-3">{f.icon}</div>
              <div className="font-semibold text-gray-900 mb-1">{f.title}</div>
              <div className="text-sm text-gray-500">{f.desc}</div>
            </div>
          ))}
        </div>
      </section>
    </div>
  )
}
