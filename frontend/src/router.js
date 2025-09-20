import { createRouter, createWebHistory } from 'vue-router';
import DashboardView from './views/DashboardView.vue';
import TableView from './views/TableView.vue';
import SettingsView from './views/SettingsView.vue';

const routes = [
  { path: '/', name: 'dashboard', component: DashboardView },
  { path: '/table', name: 'table', component: TableView },
  { path: '/settings', name: 'settings', component: SettingsView },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
