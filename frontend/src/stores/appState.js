import { reactive } from 'vue';
import { apiFetch } from '../lib/api';

const state = reactive({
  initialized: false,
  loading: false,
  user: null,
  expenses: [],
  categories: [],
  currency: 'usd',
  startDate: 1,
  endOfMonth: false,
  tags: [],
  recurringExpenses: [],
  budgetSummaries: [],
  budgetSummariesMonth: '',
  theme: typeof window !== 'undefined' ? (localStorage.getItem('theme') || 'system') : 'system',
});

function extractTags(expenses) {
  const tags = new Set();
  expenses.forEach((expense) => {
    if (Array.isArray(expense.tags)) {
      expense.tags.forEach((tag) => tags.add(tag));
    }
  });
  return Array.from(tags);
}

export async function loadSession() {
  const response = await apiFetch('/api/v1/session');
  if (!response.ok) throw new Error('Failed to fetch session');
  state.user = await response.json();
}

export async function loadInitialData() {
  if (state.initialized || state.loading) return;
  state.loading = true;
  try {
    await loadSession();

    const configResponse = await apiFetch('/config');
    if (!configResponse.ok) throw new Error('Failed to fetch configuration');
    const config = await configResponse.json();
    state.categories = config.categories || [];
    state.currency = config.currency || 'usd';
    state.startDate = config.startDate || 1;
    state.endOfMonth = !!config.endOfMonth;

    await refreshExpenses();
    await refreshRecurringExpenses();

    state.initialized = true;
  } finally {
    state.loading = false;
  }
}

export async function refreshExpenses() {
  const response = await apiFetch('/expenses');
  if (!response.ok) throw new Error('Failed to fetch expenses');
  const data = await response.json();
  state.expenses = Array.isArray(data)
    ? data
    : Array.isArray(data?.expenses)
      ? data.expenses
      : [];
  state.tags = extractTags(state.expenses);
}

export async function refreshRecurringExpenses() {
  const response = await apiFetch('/recurring-expenses');
  if (!response.ok) {
    state.recurringExpenses = [];
    return;
  }
  const data = await response.json();
  state.recurringExpenses = Array.isArray(data)
    ? data
    : Array.isArray(data?.expenses)
      ? data.expenses
      : [];
}

export async function refreshBudgetSummaries(date = new Date()) {
  const target = new Date(date);
  const key = formatMonthKey(target);
  const response = await apiFetch(`/budgets?month=${encodeURIComponent(key)}`);
  if (!response.ok) {
    state.budgetSummaries = [];
    state.budgetSummariesMonth = key;
    return;
  }
  const data = await response.json();
  state.budgetSummaries = Array.isArray(data) ? data : [];
  state.budgetSummariesMonth = key;
}

export function resetState() {
  state.initialized = false;
  state.loading = false;
  state.user = null;
  state.expenses = [];
  state.categories = [];
  state.currency = 'usd';
  state.startDate = 1;
  state.endOfMonth = false;
  state.tags = [];
  state.recurringExpenses = [];
  state.budgetSummaries = [];
  state.budgetSummariesMonth = '';
}

export function isAdmin() {
  return state.user?.role === 'admin';
}

export function addCategoryLocally(category) {
  if (!state.categories.includes(category)) {
    state.categories = [...state.categories, category];
  }
}

export function setUser(user) {
  state.user = user;
}

function formatMonthKey(date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  return `${year}-${month}`;
}

export default state;
