<template>
  <RouterView v-if="isAuthRoute" />
  <div
    v-else
    class="mx-auto flex min-h-screen w-full max-w-6xl flex-col gap-6 px-4 pb-10 pt-6"
  >
    <header class="flex flex-wrap items-center justify-between gap-4 border-b border-[var(--border)] pb-4">
      <nav class="flex flex-wrap items-center gap-3">
        <RouterLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          :title="link.tooltip"
          :class="[
            navIconButton,
            route.path === link.to && 'bg-[var(--accent)] text-white shadow-lg'
          ]"
        >
          <i :class="link.icon"></i>
        </RouterLink>
      </nav>
      <button
        type="button"
        :class="[navIconButton, 'bg-[var(--accent)] text-white hover:bg-[var(--accent)]/90']"
        title="Logout"
        @click="handleLogout"
      >
        <i class="fa-solid fa-right-from-bracket"></i>
      </button>
    </header>
    <main class="flex-1">
      <RouterView />
    </main>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useRoute, useRouter, RouterLink, RouterView } from 'vue-router';
import { apiFetch, clearAuthToken } from './lib/api';
import { resetState } from './stores/appState';

const route = useRoute();
const router = useRouter();

const links = [
  { to: '/', icon: 'fa-solid fa-chart-pie', tooltip: 'Dashboard' },
  { to: '/table', icon: 'fa-solid fa-table', tooltip: 'Table View' },
  { to: '/settings', icon: 'fa-solid fa-gear', tooltip: 'Settings' },
];

const navIconButton =
  'inline-flex h-12 w-12 items-center justify-center rounded-full border border-[var(--border)] bg-[var(--bg-secondary)] text-lg text-[var(--text-primary)] transition duration-150 ease-out hover:bg-[var(--accent)] hover:text-white hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40 focus:ring-offset-2 focus:ring-offset-[var(--bg-primary)]';

const isAuthRoute = computed(() => route.name === 'auth');

async function handleLogout() {
  try {
    await apiFetch('/api/v1/user/logout', { method: 'POST' });
  } catch (error) {
    console.error('Failed to log out', error);
  } finally {
    clearAuthToken();
    resetState();
    router.push({ path: '/auth' });
  }
}
</script>
