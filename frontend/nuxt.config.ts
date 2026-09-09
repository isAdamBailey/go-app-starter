import tailwindcss from '@tailwindcss/vite'

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({

  modules: ['@nuxt/eslint', '@pinia/nuxt', 'nuxt-security'],
  // SPA mode: this app is a thin client over the Go API and needs no SSR.
  ssr: false,
  devtools: { enabled: true },

  app: {
    head: {
      // TODO: replace with your app's name.
      title: 'App',
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
      ],
      meta: [
        { name: 'description', content: 'TODO: describe your app.' },
      ],
    },
  },
  css: ['~/assets/css/main.css'],

  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE
        || (process.env.NODE_ENV === 'production' ? '' : 'http://localhost:8080'),
    },
  },
  compatibilityDate: '2025-07-15',

  vite: {
    plugins: [tailwindcss()],
  },

  // Nuxt's ESLint module includes a stylistic/formatting layer, so there's
  // no separate Prettier setup to maintain (and no ESLint/Prettier rule
  // conflicts). This is the config the Nuxt team recommends as of
  // @nuxt/eslint v1: https://eslint.nuxt.com/packages/module#stylistic
  eslint: {
    config: {
      stylistic: true,
    },
  },

  // nuxt-security defaults are tuned for SSR HTML pages; this is an SSR-less
  // SPA shell with no third-party embeds, so CSP can stay strict. See
  // https://nuxt-security.vercel.app for per-app tuning as you add features
  // (e.g. an <img> source domain, an analytics script).
  security: {
    headers: {
      crossOriginEmbedderPolicy: 'unsafe-none',
    },
  },
})
