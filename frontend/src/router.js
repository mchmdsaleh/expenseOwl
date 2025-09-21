import { createRouter, createWebHistory } from 'vue-router';
import DashboardView from './views/DashboardView.vue';
import TableView from './views/TableView.vue';
import SettingsView from './views/SettingsView.vue';
import AuthView from './views/AuthView.vue';
import { getAuthToken } from './lib/api';
import state, { loadInitialData } from './stores/appState';

const routes = [
  { path: '/', name: 'dashboard', component: DashboardView },
  { path: '/table', name: 'table', component: TableView },
  { path: '/settings', name: 'settings', component: SettingsView },
  { path: '/auth', name: 'auth', component: AuthView },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach(async (to) => {
  const token = getAuthToken();
  if (!token && to.name !== 'auth') {
    return {
      name: 'auth',
      query: { redirect: to.fullPath },
    };
  }

  if (token && to.name === 'auth') {
    const target = typeof to.query.redirect === 'string' && to.query.redirect ? to.query.redirect : '/';
    return { path: target };
  }

  if (token && !state.initialized && to.name !== 'auth') {
    try {
      await loadInitialData();
    } catch (error) {
      console.error('Failed to load initial data', error);
    }
  }

  return true;
});

export default router;
