/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/templates/**/*.templ",
    "./web/**/*.html",
  ],
  theme: {
    extend: {
      colors: {
        brown: {
          50: '#fdf8f6',
          100: '#f2e8e5',
          200: '#eaddd7',
          300: '#e0cec7',
          400: '#d2bab0',
          500: '#bfa094',
          600: '#7f5539',
          700: '#6b4423',
          800: '#4a2c2a',
          900: '#3d2319',
        },
      },
    },
  },
  plugins: [],
}
