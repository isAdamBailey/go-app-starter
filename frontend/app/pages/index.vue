<script setup lang="ts">
const auth = useAuthStore()

async function onLogout() {
  await auth.logout()
  await navigateTo('/login')
}
</script>

<template>
  <div class="min-h-screen bg-surface px-4 py-6 text-ink sm:px-6 sm:py-10">
    <div class="mx-auto max-w-3xl space-y-6">
      <header class="flex items-center justify-between">
        <h1 class="text-xl font-semibold tracking-tight">
          App
        </h1>
        <button
          type="button"
          class="rounded-md border border-line px-3 py-1.5 text-sm text-muted hover:text-ink"
          @click="onLogout"
        >
          Sign out
        </button>
      </header>

      <main class="rounded-lg border border-line bg-panel p-6">
        <p class="text-sm text-muted">
          Signed in as <span class="font-medium text-ink">{{ auth.user?.email }}</span>.
        </p>
        <p class="mt-4 text-sm text-muted">
          <!-- TODO: replace this page with your app. -->
          Auth, sessions, and CSRF are wired up — build your app's real UI here.
        </p>
      </main>
    </div>
  </div>
</template>
