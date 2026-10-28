const PUBLIC_ROUTES = new Set(['/login', '/auth/callback'])

// nginx redirects prerendered routes to a trailing slash (/auth/callback/),
// so compare paths without it or the magic link lands on the login screen.
function normalizePath(path: string): string {
  return path.length > 1 ? path.replace(/\/+$/, '') : path
}

export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore()
  const path = normalizePath(to.path)

  if (!auth.initialized) {
    await auth.fetchMe()
  }

  if (!auth.isAuthenticated && !PUBLIC_ROUTES.has(path)) {
    return navigateTo('/login')
  }

  if (auth.isAuthenticated && path === '/login') {
    return navigateTo('/')
  }
})
