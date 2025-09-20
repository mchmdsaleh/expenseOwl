/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        surface: 'var(--bg-secondary)',
        background: 'var(--bg-primary)',
        accent: 'var(--accent)',
        text: 'var(--text-primary)',
        muted: 'var(--text-secondary)',
        border: 'var(--border)',
      },
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
      },
      boxShadow: {
        card: '0 10px 25px rgba(8, 21, 40, 0.25)',
      },
    },
  },
  plugins: [],
};
