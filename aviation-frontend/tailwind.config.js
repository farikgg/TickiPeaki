/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx}'],
  theme: {
    extend: {
      colors: {
        brand: {
          DEFAULT: '#FF6D00',
          50: '#FFF8F5',
        },
      },
    },
  },
  plugins: [],
}
