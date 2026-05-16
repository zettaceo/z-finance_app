/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        'zf-bg':     '#07090f',
        'zf-card':   '#0d1220',
        'zf-card2':  '#111a2e',
        'zf-green':  '#00e87a',
        'zf-green2': '#00c96b',
        'zf-blue':   '#3b82f6',
        'zf-purple': '#8b5cf6',
        'zf-gold':   '#f59e0b',
        'zf-red':    '#ef4444',
        'zf-t1':     '#f1f5f9',
        'zf-t2':     '#8b9bb4',
        'zf-t3':     '#4a5568',
      },
      fontFamily: {
        syne: ['Syne', 'sans-serif'],
        sans: ['DM Sans', 'system-ui', 'sans-serif'],
      },
      borderRadius: {
        'zf': '20px',
        'zf-sm': '12px',
      },
      animation: {
        'fade-in': 'fadeIn 0.4s ease forwards',
        'slide-up': 'slideUp 0.4s ease forwards',
        'slide-right': 'slideRight 0.35s ease forwards',
        'count-up': 'countUp 0.6s ease forwards',
        'pulse-green': 'pulseGreen 2s ease-in-out infinite',
        'spin-slow': 'spin 3s linear infinite',
      },
      keyframes: {
        fadeIn:     { from: { opacity: 0 }, to: { opacity: 1 } },
        slideUp:    { from: { opacity: 0, transform: 'translateY(20px)' }, to: { opacity: 1, transform: 'translateY(0)' } },
        slideRight: { from: { opacity: 0, transform: 'translateX(40px)' }, to: { opacity: 1, transform: 'translateX(0)' } },
        countUp:    { from: { opacity: 0, transform: 'translateY(8px)' }, to: { opacity: 1, transform: 'translateY(0)' } },
        pulseGreen: { '0%,100%': { boxShadow: '0 0 0 0 rgba(0,232,122,0.3)' }, '50%': { boxShadow: '0 0 0 8px rgba(0,232,122,0)' } },
      },
      backdropBlur: {
        'zf': '24px',
      },
    },
  },
  plugins: [],
}
