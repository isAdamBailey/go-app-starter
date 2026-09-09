<script setup lang="ts">
const auth = useAuthStore()

const email = ref('')
const status = ref<'idle' | 'sending' | 'sent' | 'error'>('idle')

async function onSubmit() {
  status.value = 'sending'
  try {
    await auth.requestMagicLink(email.value.trim())
    status.value = 'sent'
  }
  catch {
    status.value = 'error'
  }
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-surface px-4 py-6 text-ink sm:px-6 sm:py-10">
    <div class="mx-auto flex w-full max-w-3xl justify-center">
      <div class="w-full max-w-sm space-y-8">
        <div class="space-y-2 text-center">
          <!-- TODO: replace with your app's name -->
          <h1 class="text-2xl font-semibold tracking-tight">
            App
          </h1>
          <p class="text-sm text-muted">
            Sign in with your email
          </p>
        </div>

        <Transition
          name="fade"
          mode="out-in"
        >
          <form
            v-if="status !== 'sent'"
            class="space-y-3"
            @submit.prevent="onSubmit"
          >
            <div>
              <label
                for="email"
                class="block text-sm text-muted"
              >Email</label>
              <input
                id="email"
                v-model="email"
                type="email"
                required
                placeholder="you@example.com"
                autocomplete="email"
                class="mt-1 w-full rounded-md border border-line bg-panel px-3 py-2 text-sm text-ink placeholder:text-muted/60"
              >
            </div>

            <button
              type="submit"
              :disabled="status === 'sending'"
              class="w-full rounded-md bg-accent px-4 py-2 text-sm font-medium text-white transition-colors duration-150 hover:bg-accent-hover disabled:opacity-50"
            >
              {{ status === 'sending' ? 'Sending…' : 'Send sign-in link' }}
            </button>

            <p
              v-if="status === 'error'"
              class="text-sm text-danger"
            >
              Something went wrong. Please try again.
            </p>
          </form>

          <p
            v-else
            class="text-center text-sm text-muted"
          >
            If that email is allowed to sign in, a link has been sent. Check your
            inbox and click the link to continue.
          </p>
        </Transition>
      </div>
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 150ms ease-out;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@media (prefers-reduced-motion: reduce) {
  .fade-enter-active,
  .fade-leave-active {
    transition: none;
  }
}
</style>
