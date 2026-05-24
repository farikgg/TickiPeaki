import { useState } from 'react'
import { NavLink, Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext.jsx'

export default function Navbar() {
  const { isAuthenticated, user, logout, hasPassenger } = useAuth()
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)

  const handleLogout = () => {
    logout()
    setOpen(false)
    navigate('/')
  }

  const linkClass = ({ isActive }) =>
    isActive
      ? 'text-[#FF6D00] font-semibold'
      : 'text-gray-600 hover:text-[#FF6D00] transition'

  const showDot = isAuthenticated && !hasPassenger

  return (
    <header className="sticky top-0 z-50 bg-white border-b border-gray-100 shadow-sm">
      <div className="max-w-6xl mx-auto px-4 h-16 flex items-center justify-between">
        <Link to="/" className="flex items-center gap-2 text-xl font-bold text-[#FF6D00]">
          <span>TickiPeaki</span>
        </Link>

        <nav className="hidden md:flex items-center gap-8">
          <NavLink to="/" end className={linkClass}>Главная</NavLink>
          <NavLink to="/flights" className={linkClass}>Рейсы</NavLink>
        </nav>

        <div className="hidden md:flex items-center gap-3">
          {isAuthenticated ? (
            <>
              <span className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-gray-100 text-gray-700 text-sm">
                <span>👤</span>
                <span>{user?.username || 'Пользователь'}</span>
              </span>
              <Link
                to="/profile"
                className="relative px-4 py-2 rounded-xl bg-[#FF6D00] hover:bg-orange-600 text-white text-sm font-semibold transition"
              >
                <span className="relative">
                  Личный кабинет
                  {showDot && (
                    <span className="absolute -top-1 -right-3 w-2 h-2 bg-[#FF6D00] rounded-full animate-pulse ring-2 ring-white" />
                  )}
                </span>
              </Link>
              <button
                onClick={handleLogout}
                className="text-sm text-gray-500 hover:text-red-500 transition"
              >
                Выйти
              </button>
            </>
          ) : (
            <>
              <Link
                to="/login"
                className="px-4 py-2 rounded-xl border border-[#FF6D00] text-[#FF6D00] hover:bg-orange-50 text-sm font-semibold transition"
              >
                Войти
              </Link>
              <Link
                to="/register"
                className="px-4 py-2 rounded-xl bg-[#FF6D00] hover:bg-orange-600 text-white text-sm font-semibold transition"
              >
                Регистрация
              </Link>
            </>
          )}
        </div>

        <button
          onClick={() => setOpen(o => !o)}
          className="md:hidden p-2 text-gray-600 text-xl"
          aria-label="Menu"
        >
          ☰
        </button>
      </div>

      {open && (
        <div className="md:hidden border-t border-gray-100 bg-white px-4 py-4 flex flex-col gap-3">
          <NavLink to="/" end className={linkClass} onClick={() => setOpen(false)}>Главная</NavLink>
          <NavLink to="/flights" className={linkClass} onClick={() => setOpen(false)}>Рейсы</NavLink>
          {isAuthenticated ? (
            <>
              <div className="flex items-center gap-2 text-sm text-gray-700">
                <span>👤</span>
                <span>{user?.username}</span>
              </div>
              <Link
                to="/profile"
                onClick={() => setOpen(false)}
                className="bg-[#FF6D00] text-white px-4 py-2 rounded-xl text-center font-semibold relative"
              >
                <span className="relative">
                  Личный кабинет
                  {showDot && (
                    <span className="absolute -top-1 -right-3 w-2 h-2 bg-white rounded-full animate-pulse" />
                  )}
                </span>
              </Link>
              <button onClick={handleLogout} className="text-left text-red-500">Выйти</button>
            </>
          ) : (
            <>
              <Link
                to="/login"
                onClick={() => setOpen(false)}
                className="border border-[#FF6D00] text-[#FF6D00] px-4 py-2 rounded-xl text-center font-semibold"
              >
                Войти
              </Link>
              <Link
                to="/register"
                onClick={() => setOpen(false)}
                className="bg-[#FF6D00] text-white px-4 py-2 rounded-xl text-center font-semibold"
              >
                Регистрация
              </Link>
            </>
          )}
        </div>
      )}
    </header>
  )
}
