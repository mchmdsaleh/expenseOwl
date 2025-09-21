import { createRouter, createWebHistory } from 'vue-router';
import DashboardView from './views/DashboardView.vue';
import TableView from './views/TableView.vue';
import SettingsView from './views/SettingsView.vue';
import AuthView from './views/AuthView.vue';
import ProfileSettingsView from './views/ProfileSettingsView.vue';
import UserManagementView from './views/UserManagementView.vue';
import IntegrationsView from './views/IntegrationsView.vue';
import { getAuthToken } from './lib/api';
import state, { loadInitialData, loadSession } from './stores/appState';

const routes = [
  { path: '/', name: 'dashboard', component: DashboardView },
  { path: '/table', name: 'table', component: TableView },
  { path: '/settings', name: 'settings', component: SettingsView },
  { path: '/integrations', name: 'integrations', component: IntegrationsView },
  { path: '/profile', name: 'profile', component: ProfileSettingsView },
  { path: '/admin/users', name: 'admin-users', component: UserManagementView, meta: { requiresAdmin: true } },
  { path: '/login', name: 'login', component: AuthView },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach(async (to) => {
  const token = getAuthToken();
  if (!token && to.name !== 'login') {
    return {
      name: 'login',
      query: { redirect: to.fullPath },
    };
  }

  if (token && to.name === 'login') {
    const target = typeof to.query.redirect === 'string' && to.query.redirect ? to.query.redirect : '/';
    return { path: target };
  }

  if (token && !state.initialized && to.name !== 'login') {
    try {
      await loadInitialData();
    } catch (error) {
      console.error('Failed to load initial data', error);
    }
  }

  if (token && !state.user) {
    try {
      await loadSession();
    } catch (error) {
      console.error('Failed to load session', error);
    }
  }

  if (to.meta?.requiresAdmin && state.user?.role !== 'admin') {
    return { path: '/' };
  }

  return true;
});

export default router;
