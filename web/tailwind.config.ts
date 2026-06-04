import type { Config } from 'tailwindcss';

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        graphite: 'var(--color-shell)',
        canvas: 'var(--color-bg)',
        panel: 'var(--color-panel)',
        muted: 'var(--color-panel-muted)',
        line: 'var(--color-border)',
        ink: 'var(--color-text)',
        subtle: 'var(--color-muted)',
        signal: 'rgb(var(--color-signal-rgb) / <alpha-value>)',
        ok: 'rgb(var(--color-ok-rgb) / <alpha-value>)',
        warn: 'rgb(var(--color-warn-rgb) / <alpha-value>)',
        fault: 'rgb(var(--color-fault-rgb) / <alpha-value>)',
      },
      fontFamily: {
        sans: ['Inter', 'Segoe UI', 'Arial', 'sans-serif'],
        mono: ['JetBrains Mono', 'Consolas', 'monospace'],
      },
    },
  },
  plugins: [],
} satisfies Config;
