<template>
  <div class="container flex min-h-screen flex-col gap-6">
    <header class="border-b border-[var(--border)] pb-4">
      <div class="nav-bar">
        <RouterLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="view-button"
          :class="{ active: route.path === link.to }"
          :data-tooltip="link.tooltip"
        >
          <i :class="link.icon"></i>
        </RouterLink>
        <button
          type="button"
          class="view-button"
          data-tooltip="Logout"
          @click="handleLogout"
        >
          <i class="fa-solid fa-right-from-bracket"></i>
        </button>
      </div>
    </header>
    <main class="flex-1">
      <RouterView />
    </main>
  </div>
</template>

<script setup>
import { useRoute, RouterLink, RouterView } from 'vue-router';
import { apiFetch } from './lib/api';

const route = useRoute();

const links = [
  { to: '/', icon: 'fa-solid fa-chart-pie', tooltip: 'Dashboard' },
  { to: '/table', icon: 'fa-solid fa-table', tooltip: 'Table View' },
  { to: '/settings', icon: 'fa-solid fa-gear', tooltip: 'Settings' },
];

async function handleLogout() {
  try {
    await apiFetch('/logout', { method: 'POST' });
  } catch (error) {
    console.error('Failed to log out', error);
  } finally {
    window.location.href = '/login';
  }
}
</script>
