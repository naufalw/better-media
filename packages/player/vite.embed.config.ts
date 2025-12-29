import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  root:'.',
  build: {
    outDir: 'dist/embed',
    rollupOptions: {
      input: 'src/embed.tsx',
      output: {
        entryFileNames: 'player.js',
        assetFileNames: 'player.[ext]',
      },
    },
  },
})
