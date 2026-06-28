import forms from '@tailwindcss/forms'

/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: [
    './components/**/*.{vue,js,ts}',
    './layouts/**/*.vue',
    './pages/**/*.vue',
    './composables/**/*.{js,ts}',
    './plugins/**/*.{js,ts}',
    './utils/**/*.{js,ts}',
    './src/**/*.{vue,js,ts}',
    './app.vue',
    './error.vue',
  ],
  theme: {
    extend: {
      colors: {
        bg: '#F1F2F3',
        accent: {
          DEFAULT: '#fb7299',
          light: 'rgba(251, 114, 153, 0.12)',
          hover: 'rgba(251, 114, 153, 0.08)',
        },
        glass: {
          DEFAULT: 'var(--glass-bg)',
          solid: 'var(--glass-bg-solid)',
          hover: 'var(--glass-bg-hover)',
          border: 'var(--glass-border)',
        },
      },
      borderRadius: {
        glass: 'var(--radius-lg)',
      },
      backdropBlur: {
        glass: 'var(--glass-blur)',
        sidebar: 'var(--sidebar-blur)',
        tabbar: 'var(--tabbar-blur)',
      },
      boxShadow: {
        glass: 'var(--glass-shadow)',
        'glass-lg': 'var(--glass-shadow-lg)',
        'glass-sm': 'var(--glass-shadow-sm)',
        'accent-glow': 'var(--accent-glow)',
      },
      animation: {
        'fade-in': 'fadeIn 0.4s var(--ease-out) both',
        'slide-up': 'slideUp 0.4s var(--ease-out) both',
        'scale-in': 'scaleIn 0.25s var(--ease-out) both',
      },
      keyframes: {
        fadeIn: {
          from: { opacity: '0' },
          to: { opacity: '1' },
        },
        slideUp: {
          from: { opacity: '0', transform: 'translateY(12px)' },
          to: { opacity: '1', transform: 'translateY(0)' },
        },
        scaleIn: {
          from: { opacity: '0', transform: 'scale(0.95)' },
          to: { opacity: '1', transform: 'scale(1)' },
        },
      },
    },
    screens: {
      ssm: '0px',
      lm: { max: '640px' },
      ld: { max: '768px' },
      llg: { max: '1023px' },
      sm: '640px',
      smd: '641px',
      md: '768px',
      lg: '1024px',
      xl: '1280px',
      '2xl': '1536px',
    },
  },
  plugins: [forms],
}
