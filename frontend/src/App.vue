<template>
  <RouterView v-if="isAuthRoute" />
  <div v-else class="mx-auto flex min-h-screen w-full max-w-6xl flex-col gap-8 px-4 md:px-6 pb-12 pt-6">
    <!-- Modern Floating Header -->
    <header
      class="sticky top-4 z-40 flex items-center justify-between gap-2 md:gap-4 glass-card rounded-3xl px-3 md:px-6 py-3 md:py-4 shadow-2xl transition-all duration-300"
    >
      <div class="flex items-center gap-6">
        <!-- Logo/Title -->
        <div class="hidden items-center gap-2 md:flex">
          <div class="bg-[var(--accent)] p-2 rounded-xl text-white shadow-lg">
            <i class="fa-solid fa-owl text-xl"></i>
          </div>
          <span class="text-xl font-black tracking-tight text-[var(--text-primary)]">Expense<span class="text-[var(--accent)]">Owl</span></span>
        </div>

        <nav class="flex items-center gap-1 md:gap-2">
          <RouterLink
            v-for="link in links"
            :key="link.to"
            :to="link.to"
            :title="link.tooltip"
            class="group relative flex h-10 w-10 md:h-12 md:w-12 items-center justify-center rounded-2xl transition-all duration-300 hover:bg-[var(--bg-elevated)]"
            :class="[route.path === link.to ? 'bg-[var(--accent)] text-white shadow-indigo-500/20 shadow-xl scale-110' : 'text-[var(--text-secondary)]']"
          >
            <i :class="[link.icon, 'text-base md:text-lg group-hover:scale-110 transition-transform']"></i>
            <span v-if="route.path === link.to" class="absolute -bottom-1 h-1 w-1 rounded-full bg-white"></span>
          </RouterLink>
        </nav>
      </div>

      <div class="flex items-center gap-1 md:gap-3">
        <RouterLink
          v-if="state.user"
          to="/profile"
          title="Profile"
          class="flex h-10 w-10 md:h-12 md:w-12 items-center justify-center rounded-2xl text-[var(--text-secondary)] transition-all duration-300 hover:bg-[var(--bg-elevated)] hover:text-[var(--text-primary)]"
          :class="[route.path === '/profile' && 'bg-[var(--bg-elevated)] text-[var(--text-primary)]']"
        >
          <i class="fa-solid fa-user-ninja text-base md:text-lg"></i>
        </RouterLink>
        <div class="hidden md:block h-8 w-px bg-[var(--border)] mx-1"></div>
        <button
          type="button"
          class="flex h-10 w-10 md:h-12 md:w-12 items-center justify-center rounded-2xl bg-rose-500/10 text-rose-500 transition-all duration-300 hover:bg-rose-500 hover:text-white hover:shadow-lg hover:shadow-rose-500/20 active:scale-95 shrink-0"
          title="Logout"
          @click="handleLogout"
        >
          <i class="fa-solid fa-power-off text-base md:text-lg"></i>
        </button>
      </div>
    </header>

    <main class="flex-1">
      <RouterView v-slot="{ Component }">
        <component :is="Component" :key="route.fullPath" />
      </RouterView>
    </main>
    <ChatbotWidget v-if="state.user" />
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useRoute, useRouter, RouterLink, RouterView } from 'vue-router';
import ChatbotWidget from './components/ChatbotWidget.vue';
import { apiFetch, clearAuthToken } from './lib/api';
import { clearCipher } from './lib/cipher';
import { resetEncryptionCache } from './lib/encryption';
import state, { resetState } from './stores/appState';

const route = useRoute();
const router = useRouter();

const links = computed(() => {
  const items = [
    { to: '/', icon: 'fa-solid fa-chart-pie', tooltip: 'Dashboard' },
    { to: '/table', icon: 'fa-solid fa-table', tooltip: 'Table View' },
    { to: '/settings', icon: 'fa-solid fa-gear', tooltip: 'Settings' },
  ];
  if (state.user?.role === 'admin') {
    items.push({ to: '/admin/users', icon: 'fa-solid fa-users-gear', tooltip: 'User Management' });
  }
  return items;
});

const navIconButton =
  'inline-flex h-12 w-12 items-center justify-center rounded-full border border-[var(--border)] bg-[var(--bg-secondary)] text-lg text-[var(--text-primary)] transition duration-150 ease-out hover:bg-[var(--accent)] hover:text-white hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40 focus:ring-offset-2 focus:ring-offset-[var(--bg-primary)]';

const isAuthRoute = computed(() => route.name === 'login');

async function handleLogout() {
  try {
    await apiFetch('/api/v1/user/logout', { method: 'POST' });
  } catch (error) {
    console.error('Failed to log out', error);
  } finally {
    clearAuthToken();
    clearCipher();
    resetEncryptionCache();
    resetState();
    router.push({ path: '/login' });
  }
}
</script>
