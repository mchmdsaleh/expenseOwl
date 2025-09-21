<template>
  <div class="flex min-h-screen items-center justify-center bg-[var(--bg-primary)] px-4">
    <div class="w-full max-w-md rounded-xl border border-[var(--border)] bg-[var(--bg-secondary)] p-8 shadow-xl">
      <h1 class="mb-6 text-center text-2xl font-semibold text-[var(--text-primary)]">
        {{ isLogin ? 'Sign In' : 'Create Account' }}
      </h1>
      <form class="space-y-4" @submit.prevent="handleSubmit">
        <div>
          <label class="mb-1 block text-sm font-medium text-[var(--text-secondary)]" for="email">Email</label>
          <input
            id="email"
            v-model="form.email"
            type="email"
            required
            class="w-full rounded-lg border border-[var(--border)] bg-[var(--bg-primary)] px-3 py-2 text-[var(--text-primary)] focus:border-[var(--accent)] focus:outline-none"
          />
        </div>
        <div class="grid gap-4" v-if="!isLogin">
          <div>
            <label class="mb-1 block text-sm font-medium text-[var(--text-secondary)]" for="firstName">First Name</label>
            <input
              id="firstName"
              v-model="form.firstName"
              type="text"
              required
              class="w-full rounded-lg border border-[var(--border)] bg-[var(--bg-primary)] px-3 py-2 text-[var(--text-primary)] focus:border-[var(--accent)] focus:outline-none"
            />
          </div>
          <div>
            <label class="mb-1 block text-sm font-medium text-[var(--text-secondary)]" for="lastName">Last Name</label>
            <input
              id="lastName"
              v-model="form.lastName"
              type="text"
              required
              class="w-full rounded-lg border border-[var(--border)] bg-[var(--bg-primary)] px-3 py-2 text-[var(--text-primary)] focus:border-[var(--accent)] focus:outline-none"
            />
          </div>
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium text-[var(--text-secondary)]" for="password">Password</label>
          <input
            id="password"
            v-model="form.password"
            type="password"
            required
            minlength="6"
            class="w-full rounded-lg border border-[var(--border)] bg-[var(--bg-primary)] px-3 py-2 text-[var(--text-primary)] focus:border-[var(--accent)] focus:outline-none"
          />
        </div>
        <p v-if="errorMessage" class="rounded-md bg-red-100 px-3 py-2 text-sm text-red-700">
          {{ errorMessage }}
        </p>
        <button
          type="submit"
          class="flex w-full items-center justify-center rounded-lg bg-[var(--accent)] px-4 py-2 font-semibold text-white transition hover:bg-[var(--accent)]/90"
          :disabled="submitting"
        >
          {{ submitting ? 'Please wait…' : isLogin ? 'Sign In' : 'Create Account' }}
        </button>
      </form>
      <p class="mt-6 text-center text-sm text-[var(--text-secondary)]">
        {{ isLogin ? "Don't have an account?" : 'Already have an account?' }}
        <button
          type="button"
          class="font-medium text-[var(--accent)] hover:underline"
          @click="toggleMode"
        >
          {{ isLogin ? 'Create one' : 'Sign in' }}
        </button>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { apiFetch, setAuthToken } from '../lib/api';
import { loadInitialData } from '../stores/appState';

const route = useRoute();
const router = useRouter();
const mode = ref('login');
const submitting = ref(false);
const errorMessage = ref('');

const form = ref({
  email: '',
  password: '',
  firstName: '',
  lastName: '',
});

const isLogin = computed(() => mode.value === 'login');

function toggleMode() {
  mode.value = isLogin.value ? 'signup' : 'login';
  errorMessage.value = '';
}

async function handleSubmit() {
  submitting.value = true;
  errorMessage.value = '';
  try {
    const endpoint = isLogin.value ? '/api/v1/user/login' : '/api/v1/user/signup';
    const payload = {
      email: form.value.email,
      password: form.value.password,
    };
    if (!isLogin.value) {
      payload.firstName = form.value.firstName;
      payload.lastName = form.value.lastName;
    }
    const response = await apiFetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    });
    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      errorMessage.value = data?.error || 'Authentication failed';
      return;
    }
    const data = await response.json();
    setAuthToken(data.token);
    await loadInitialData();
    const target = typeof route.query.redirect === 'string' && route.query.redirect ? route.query.redirect : '/';
    router.push(target);
  } catch (error) {
    console.error('Authentication error', error);
    errorMessage.value = 'Something went wrong. Please try again.';
  } finally {
    submitting.value = false;
  }
}
</script>
