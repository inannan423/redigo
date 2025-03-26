/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./app/**/*.{js,jsx,ts,tsx,md,mdx}",
    "./app/layout.jsx",
    "./content/**/*.{md,mdx}",
    "./src/**/*.{js,jsx,ts,tsx,md,mdx}",
    "./node_modules/@heroui/theme/dist/components/(button|ripple|spinner).js"
  ],
    theme: {
      extend: {
        fontFamily: {
          sans: ['zsft-443', 'system-ui', '-apple-system', 'sans-serif'],
          mono: ['zsft-443', 'Menlo', 'Monaco', 'Consolas', 'monospace'],
        }
      }
    },
    darkMode: 'class',
  plugins: []
}