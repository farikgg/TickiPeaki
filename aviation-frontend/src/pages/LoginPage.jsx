import { useState } from 'react'
import { Link, useNavigate, useLocation, Navigate } from 'react-router-dom'
import { login as apiLogin } from '../api/client.js'
import { useAuth } from '../context/AuthContext.jsx'

export default function LoginPage() {
  const { login, isAuthenticated } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const from = location.state?.from || '/'

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [showPwd, setShowPwd] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  if (isAuthenticated) return <Navigate to="/" replace />

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    try {
      const res = await apiLogin({ username, password })
      const token = res.data?.token || res.data?.access_token || res.data?.data?.token
      if (!token) throw new Error('Сервер не вернул токен')
      login(token)
      navigate(from, { replace: true })
    } catch (err) {
      setError(
        err.response?.data?.message ||
        err.response?.data?.error ||
        err.message ||
        'Ошибка входа'
      )
    } finally {
      setLoading(false)
    }
  }

  const inputCls = 'w-full px-4 py-3 bg-white border border-gray-200 focus:border-[#FF6D00] focus:ring-2 focus:ring-orange-100 rounded-xl outline-none transition'

  return (
    <div className="max-w-md mx-auto mt-12 md:mt-16">
      <div className="bg-white rounded-2xl shadow-md p-8">
        <div className="text-center mb-8">
          <div className="text-4xl mb-2 text-[#FF6D00]">✈</div>
          <h1 className="text-2xl font-bold text-gray-900">Войти в SkyBook</h1>
          <p className="text-gray-500 text-sm mt-1">Добро пожаловать обратно</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm text-gray-500 mb-1">
              Имя пользователя
            </label>
            <input
              required
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="username"
              className={inputCls}
            />
          </div>

          <div>
            <label className="block text-sm text-gray-500 mb-1">Пароль</label>
            <div className="relative">
              <input
                required
                type={showPwd ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                className={`${inputCls} pr-10`}
              />
              <button
                type="button"
                onClick={() => setShowPwd((p) => !p)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                aria-label="Toggle password visibility"
              >
                {showPwd ? '🙈' : '👁'}
              </button>
            </div>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-[#FF6D00] hover:bg-orange-600 text-white font-semibold py-3 rounded-xl transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
          >
            {loading ? (
              <>
                <span className="inline-block w-5 h-5 border-2 border-white/40 border-t-white rounded-full animate-spin"></span>
                Входим...
              </>
            ) : (
              'Войти'
            )}
          </button>

          {error && (
            <div className="text-rose-500 text-sm text-center">{error}</div>
          )}
        </form>

        <div className="text-center text-sm text-gray-500 mt-6">
          Нет аккаунта?{' '}
          <Link
            to="/register"
            className="text-[#FF6D00] font-semibold hover:underline"
          >
            Зарегистрироваться →
          </Link>
        </div>
      </div>
    </div>
  )
}
