<template>
  <section class="space-y-6">
    <div class="flex flex-col items-center gap-3 md:flex-row md:items-center md:justify-between">
      <div v-if="showAll" class="min-w-[200px] text-center text-2xl font-bold">{{ periodLabel }}</div>
      <div v-else-if="dateFilter === 'month'" class="flex items-center justify-center gap-4">
        <button :class="iconButtonClass" @click="gotoPrevMonth">
          <i class="fa-solid fa-arrow-left"></i>
        </button>
        <div class="min-w-[200px] text-center text-2xl font-bold">{{ periodLabel }}</div>
        <button :class="iconButtonClass" @click="gotoNextMonth">
          <i class="fa-solid fa-arrow-right"></i>
        </button>
      </div>
      <div v-else class="min-w-[200px] text-center text-2xl font-bold">{{ periodLabel }}</div>
    </div>

    <div class="flex flex-col gap-2 text-sm text-[var(--text-secondary)] md:flex-row md:items-center md:justify-between">
      <div
        v-if="userDisplayName"
        class="inline-flex w-full items-center justify-center gap-2 self-start rounded-2xl border border-[var(--border)] bg-[var(--bg-secondary)]/70 px-4 py-2 text-[var(--text-primary)] md:w-auto md:justify-start md:self-auto"
      >
        <i class="fa-solid fa-circle-user text-[var(--accent)]"></i>
        <span>{{ userDisplayName }}</span>
      </div>
      <div class="flex items-center gap-2 self-end md:self-auto">
        <label for="showAllToggle" class="flex items-center gap-2">
          <input id="showAllToggle" v-model="showAll" type="checkbox" :class="checkboxClass" />
          Show All Transactions
        </label>
        <AddExpenseSpeedDial
          :manual-open="showExpenseForm"
          :typing-open="showTypingForm"
          :primary-button-class="primaryButtonClass"
          :speed-dial-button-class="speedDialButtonClass"
          @open-manual="handleOpenManual"
          @open-typing="handleOpenTyping"
          @close-all="handleCloseAddPanels"
        />
      </div>
    </div>

    <div v-if="showExpenseForm" :class="cardClass" ref="manualCardRef">
      <form class="grid gap-4 md:grid-cols-2" @submit.prevent="submitExpense">
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)]" for="name">Name</label>
          <input id="name" v-model="form.name" type="text" :class="inputClass" required />
        </div>
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)]" for="category">Category</label>
          <select id="category" v-model="form.category" :class="inputClass" required>
            <option value="" disabled>Choose category</option>
            <option v-for="category in state.categories" :key="category" :value="category">
              {{ category }}
            </option>
          </select>
        </div>
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)]">Tags</label>
          <TagInput v-model="form.tags" :suggestions="state.tags" />
        </div>
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)]" for="amount">Amount</label>
          <input
            id="amount"
            :value="formattedAmount"
            inputmode="decimal"
            :class="inputClass"
            required
            @input="handleAmountInput"
            @blur="normalizeAmount"
          />
        </div>
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)]" for="date">Date</label>
          <input id="date" v-model="form.date" type="date" :class="inputClass" required />
        </div>
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)] mb-2" for="reportGain">Report Gain</label>
          <label class="relative inline-flex h-6 w-12 cursor-pointer items-center">
            <input
              id="reportGain"
              v-model="form.reportGain"
              type="checkbox"
              class="peer sr-only"
            />
            <span class="absolute inset-0 rounded-full bg-[var(--border)] transition-colors duration-200 peer-checked:bg-[var(--accent)]"></span>
            <span class="absolute left-1 h-4 w-4 rounded-full bg-white transition-transform duration-200 peer-checked:translate-x-6"></span>
          </label>
        </div>
        <div class="md:col-span-2">
          <button type="submit" :class="[primaryButtonClass, 'w-full']">{{ form.submitLabel }}</button>
        </div>
      </form>
      <div
        v-if="formMessage.text"
        :class="[
          'mt-4 rounded-full px-4 py-2 text-center text-sm font-medium',
          formMessage.type === 'success'
            ? 'bg-emerald-500/20 text-emerald-200'
            : 'bg-rose-500/20 text-rose-200'
        ]"
      >
        {{ formMessage.text }}
      </div>
    </div>

    <QuickAddExpenseCard
      v-if="showTypingForm"
      ref="typingCardRef"
      :card-class="cardClass"
      :input-class="inputClass"
      :primary-button-class="primaryButtonClass"
      @switch-manual="handleOpenManual"
      @added="handleQuickAddSuccess"
    />

    <div>
      <div class="mb-5 flex flex-wrap items-center gap-2 md:gap-3">
        <!-- Date filter -->
        <div class="relative min-w-0">
          <select
            v-model="dateFilter"
            :class="selectClass"
            :disabled="showAll"
            aria-label="Filter transactions by period"
          >
            <option value="month">This Month</option>
            <option value="week">This Week</option>
            <option value="today">Today</option>
            <option value="yesterday">Yesterday</option>
            <option value="range">Custom Range</option>
          </select>
          <i
            class="fa-solid fa-chevron-down pointer-events-none absolute right-3 top-1/2 -translate-y-1/2
                  text-[var(--text-secondary)] text-[10px]"
            aria-hidden="true"
          ></i>
        </div>

        <div
          v-if="dateFilter === 'range'"
          class="flex w-full flex-wrap items-center gap-2 md:w-auto md:flex-nowrap"
        >
          <label :class="rangeInputWrapperClass">
            <i class="fa-solid fa-calendar-day shrink-0 text-sm text-[var(--text-secondary)]"></i>
            <div class="flex min-w-[140px] flex-1 flex-col gap-1">
              <span :class="rangeInputLabelClass">Start</span>
              <input v-model="rangeStart" type="date" :class="rangeDateInputClass" />
            </div>
          </label>
          <label :class="rangeInputWrapperClass">
            <i class="fa-solid fa-calendar-check shrink-0 text-sm text-[var(--text-secondary)]"></i>
            <div class="flex min-w-[140px] flex-1 flex-col gap-1">
              <span :class="rangeInputLabelClass">End</span>
              <input v-model="rangeEnd" type="date" :class="rangeDateInputClass" />
            </div>
          </label>
        </div>

        <div v-if="selectedCategory" :class="categoryFilterChipClass">
          <i class="fa-solid fa-filter text-xs text-[var(--accent)]"></i>
          <div class="flex flex-col leading-none">
            <span class="text-[10px] uppercase tracking-[0.18em] text-[var(--text-secondary)]">Category</span>
            <span class="text-sm font-medium text-[var(--text-primary)]">{{ selectedCategory }}</span>
          </div>
          <button type="button" class="text-xs text-[var(--text-secondary)] transition hover:text-[var(--text-primary)]" @click="clearCategoryFilter">
            Clear
          </button>
        </div>

        <!-- Sort option -->
        <div class="relative min-w-0">
          <select v-model="sortOption" :class="selectClass" aria-label="Sort transactions">
            <option v-for="option in sortChoices" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
          <i
            class="fa-solid fa-chevron-down pointer-events-none absolute right-3 top-1/2 -translate-y-1/2
                  text-[var(--text-secondary)] text-[10px]"
            aria-hidden="true"
          ></i>
        </div>
      </div>
      <p v-if="rangeValidationMessage" class="text-xs italic text-amber-300">{{ rangeValidationMessage }}</p>
      <div
        v-if="tableExpenses.length === 0"
        class="w-full rounded-3xl border border-dashed border-[var(--border)] bg-[var(--bg-secondary)]/60 py-12 text-center text-base italic text-[var(--text-secondary)]"
      >
        {{ emptyTableMessage }}
      </div>
      <div
        v-else
        class="relative overflow-x-auto rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)]/60 shadow-card"
      >
        <table class="w-full text-left text-sm text-[var(--text-secondary)]">
          <thead class="text-xs uppercase tracking-wide text-[var(--text-secondary)]">
            <tr class="bg-[var(--bg-primary)]/60">
              <th scope="col" class="px-6 py-3 font-semibold">Name</th>
              <th scope="col" class="px-6 py-3 font-semibold">Category</th>
              <th v-if="hasTags" scope="col" class="px-6 py-3 font-semibold">Tags</th>
              <th scope="col" class="px-6 py-3 font-semibold">Amount</th>
              <th scope="col" class="px-6 py-3 font-semibold">Date</th>
              <th scope="col" class="px-6 py-3"></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="expense in tableExpenses"
              :key="expense.id"
              class="border-t border-[var(--border)] bg-[var(--bg-secondary)]/40 text-[var(--text-primary)]"
            >
              <th scope="row" class="px-6 py-4 font-medium text-[var(--text-primary)]">{{ expense.name }}</th>
              <td class="px-6 py-4">{{ expense.category }}</td>
              <td v-if="hasTags" class="px-6 py-4 text-[var(--text-secondary)]">{{ (expense.tags || []).join(', ') }}</td>
              <td class="px-6 py-4 font-mono text-sm text-[var(--text-secondary)]">{{ formatCurrency(expense.amount) }}</td>
              <td class="px-6 py-4 text-[var(--text-secondary)]">{{ formatDate(expense.date) }}</td>
              <td class="px-6 py-4">
                <button :class="iconGhostButton" type="button" @click="editExpense(expense)">
                  <i class="fa-solid fa-pen-to-square"></i>
                </button>
                <button :class="iconDangerButton" type="button" @click="openDeleteModal(expense, $event)">
                  <i class="fa-solid fa-trash-can"></i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <transition name="fade">
      <div
        v-if="showDeleteModal"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 px-4"
        @click.self="closeDeleteModal"
      >
        <div class="w-full max-w-md rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)]/95 p-6 shadow-card backdrop-blur">
          <h3 class="text-lg font-semibold text-[var(--text-primary)]">Delete Expense</h3>
          <p class="mt-2 text-sm text-[var(--text-secondary)]">
            Are you sure you want to delete this expense? This action cannot be undone.
          </p>
          <div class="mt-6 flex justify-end gap-3">
            <button :class="primaryButtonClass" @click="closeDeleteModal">Cancel</button>
            <button :class="[primaryButtonClass, 'bg-rose-500 text-white hover:bg-rose-500/90']" @click="confirmDelete">
              Delete
            </button>
          </div>
        </div>
      </div>
    </transition>
  </section>
</template>

<script setup>
import { ref, computed, watch, onMounted, nextTick } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import TagInput from '../components/TagInput.vue';
import AddExpenseSpeedDial from '../components/AddExpenseSpeedDial.vue';
import QuickAddExpenseCard from '../components/QuickAddExpenseCard.vue';
import state, { loadInitialData, refreshExpenses } from '../stores/appState';
import { apiFetch } from '../lib/api';
import { encryptPayload } from '../lib/encryption';
import {
  formatMonth,
  getMonthExpenses,
  filterExpensesByRange,
  getISODateWithLocalTime,
  formatDateFromUTC,
  formatCurrency as formatCurrencyRaw,
  formatWeekRange,
  formatDayLabel,
  formatRangeLabel,
  getCycleAnchor,
} from '../lib/utils';

const route = useRoute();
const router = useRouter();
const currentDate = ref(new Date());
const monthCursor = ref(new Date());
const dateFilter = ref('month');
const sortOption = ref('dateDesc');
const showAll = ref(false);
const form = ref(createDefaultForm());
const editId = ref(null);
const formMessage = ref({ text: '', type: '' });
const showDeleteModal = ref(false);
const expenseToDelete = ref(null);
const rawAmount = ref('');
const rangeStart = ref(formatDateForInput(new Date()));
const rangeEnd = ref(formatDateForInput(new Date()));
const syncingFromRoute = ref(false);
const selectedCategory = ref('');
const showExpenseForm = ref(false);
const showTypingForm = ref(false);
const manualCardRef = ref(null);
const typingCardRef = ref(null);

const userDisplayName = computed(() => {
  const first = (state.user?.firstName || '').trim();
  const last = (state.user?.lastName || '').trim();
  const full = `${first} ${last}`.trim();
  return full || state.user?.email || '';
});

const periodLabel = computed(() => {
  if (showAll.value) {
    return 'All Transactions';
  }
  if (dateFilter.value === 'today') {
    return formatDayLabel(currentDate.value);
  }
  if (dateFilter.value === 'yesterday') {
    const yesterday = new Date(currentDate.value);
    yesterday.setDate(yesterday.getDate() - 1);
    return formatDayLabel(yesterday);
  }
  if (dateFilter.value === 'week') {
    return formatWeekRange(currentDate.value);
  }
  if (dateFilter.value === 'range') {
    return formatRangeLabel(rangeStart.value, rangeEnd.value);
  }
  return formatMonth(currentDate.value);
});

const filteredExpenses = computed(() => {
  let results;
  if (showAll.value) {
    results = [...state.expenses];
  } else if (dateFilter.value === 'month') {
    results = getMonthExpenses(state.expenses, currentDate.value, state.startDate, state.endOfMonth);
  } else if (dateFilter.value === 'range') {
    results = filterExpensesByRange(
      state.expenses,
      'range',
      currentDate.value,
      state.startDate,
      state.endOfMonth,
      { start: rangeStart.value, end: rangeEnd.value }
    );
  } else {
    results = filterExpensesByRange(state.expenses, dateFilter.value, currentDate.value, state.startDate, state.endOfMonth);
  }

  if (selectedCategory.value) {
    results = results.filter((expense) => normalizeCategory(expense.category) === selectedCategory.value);
  }

  return results;
});

const tableExpenses = computed(() => sortExpenses(filteredExpenses.value));

const hasTags = computed(() => tableExpenses.value.some((expense) => Array.isArray(expense.tags) && expense.tags.length > 0));

const emptyTableMessage = computed(() => {
  if (showAll.value) {
    return 'No transactions found';
  }
  if (dateFilter.value === 'yesterday') {
    return 'No transactions recorded yesterday';
  }
  if (dateFilter.value === 'range') {
    return 'No transactions found in this range';
  }
  if (dateFilter.value === 'today') {
    return 'No transactions recorded today';
  }
  if (dateFilter.value === 'week') {
    return 'No expenses recorded this week';
  }
  return 'No expenses recorded for this month';
});

const formattedAmount = computed(() => {
  if (!rawAmount.value) return '';
  const numeric = Number(rawAmount.value.replace(/[^0-9.-]/g, '')) || 0;
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(numeric);
});

watch(
  () => state.expenses,
  () => {
    if (!showAll.value) {
      ensureCurrentMonthAvailable();
    }
  }
);

watch(
  () => [state.startDate, state.endOfMonth],
  () => {
    if (showAll.value) {
      alignCurrentCycle(monthCursor.value, { updateCurrent: false, updateCursor: true });
      return;
    }
    if (dateFilter.value === 'month') {
      alignCurrentCycle(monthCursor.value);
    } else {
      alignCurrentCycle(monthCursor.value, { updateCurrent: false, updateCursor: true });
    }
  }
);

watch(dateFilter, (next) => {
  if (syncingFromRoute.value || showAll.value) return;
  if (next === 'month') {
    alignCurrentCycle(monthCursor.value);
  } else {
    currentDate.value = new Date();
  }
});

watch(showAll, (value) => {
  if (syncingFromRoute.value) return;
  if (!value) {
    if (dateFilter.value === 'month') {
      alignCurrentCycle(monthCursor.value);
    } else {
      currentDate.value = new Date();
    }
  }
});

function applyRouteFilters(query = route.query) {
  if (!query || typeof query !== 'object') return;

  const allowedFilters = new Set(['month', 'week', 'today', 'yesterday', 'range']);
  const parseBoolean = (value) => {
    if (typeof value === 'boolean') return value;
    if (typeof value !== 'string') return null;
    const normalized = value.toLowerCase();
    if (['1', 'true', 'yes'].includes(normalized)) return true;
    if (['0', 'false', 'no'].includes(normalized)) return false;
    return null;
  };
  const parseDate = (value) => {
    if (typeof value !== 'string' || !value) return null;
    const date = new Date(value);
    return Number.isNaN(date.getTime()) ? null : date;
  };

  const queryFilter = typeof query.filter === 'string' && allowedFilters.has(query.filter) ? query.filter : null;
  const querySort = typeof query.sort === 'string' && sortChoices.some((option) => option.value === query.sort)
    ? query.sort
    : null;
  const queryShowAll = parseBoolean(query.showAll);
  const anchorString = typeof query.cursor === 'string' && query.cursor ? query.cursor : typeof query.anchor === 'string' ? query.anchor : null;
  const anchorDate = parseDate(anchorString);

  syncingFromRoute.value = true;

  if (queryShowAll !== null) {
    showAll.value = queryShowAll;
  }

  if (queryFilter) {
    dateFilter.value = queryFilter;
  }

  if (queryFilter === 'range') {
    if (typeof query.start === 'string' && query.start) {
      rangeStart.value = query.start;
    }
    if (typeof query.end === 'string' && query.end) {
      rangeEnd.value = query.end;
    }
  }

  if (querySort) {
    sortOption.value = querySort;
  }

  syncingFromRoute.value = false;

  const categoryParam = typeof query.category === 'string' && query.category.trim() ? query.category.trim() : '';
  selectedCategory.value = categoryParam;

  if (showAll.value) return;

  if (anchorDate) {
    monthCursor.value = anchorDate;
  }

  if (queryFilter === 'month' && anchorDate) {
    alignCurrentCycle(anchorDate);
  } else if (anchorDate) {
    currentDate.value = anchorDate;
  }
}

function clearCategoryFilter() {
  selectedCategory.value = '';
  const nextQuery = { ...route.query };
  delete nextQuery.category;
  router.replace({ query: nextQuery }).catch(() => {});
}

onMounted(async () => {
  await loadInitialData();
  alignCurrentCycle(currentDate.value);
  applyRouteFilters(route.query);
});

watch(
  () => route.query,
  (query) => {
    applyRouteFilters(query);
  }
);

const iconButtonClass =
  'inline-flex h-11 w-11 items-center justify-center rounded-full border border-[var(--border)] bg-[var(--bg-secondary)] text-lg text-[var(--text-primary)] transition duration-150 ease-out hover:bg-[var(--accent)] hover:text-white hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40 focus:ring-offset-2 focus:ring-offset-[var(--bg-primary)]';

const primaryButtonClass =
  'inline-flex items-center justify-center gap-2 rounded-full border border-[var(--border)] bg-[var(--bg-secondary)] px-5 py-2 text-sm font-medium text-[var(--text-primary)] transition duration-150 ease-out hover:bg-[var(--accent)] hover:text-white hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40 focus:ring-offset-2 focus:ring-offset-[var(--bg-primary)] disabled:cursor-not-allowed disabled:opacity-50';

const speedDialButtonClass =
  'flex items-center gap-2 rounded-full border border-[var(--border)] bg-[var(--bg-secondary)]/90 px-4 py-2 text-sm font-medium text-[var(--text-primary)] transition duration-150 ease-out hover:bg-[var(--accent)] hover:text-white hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40 focus:ring-offset-2 focus:ring-offset-[var(--bg-primary)]';

const inputClass =
  'w-full rounded-xl border border-[var(--border)] bg-[var(--bg-primary)] px-4 py-2 text-[var(--text-primary)] placeholder:text-[var(--text-secondary)] focus:border-[var(--accent)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40';

const checkboxClass =
  'h-4 w-4 rounded border-[var(--border)] bg-[var(--bg-primary)] text-[var(--accent)] focus:ring-[var(--accent)]/60 focus:ring-offset-0';

const cardClass =
  'rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)]/80 p-6 shadow-card backdrop-blur';

const iconGhostButton =
  'inline-flex h-9 w-9 items-center justify-center rounded-full border border-transparent text-[var(--text-secondary)] transition duration-150 hover:border-[var(--border)] hover:bg-[var(--bg-primary)]/60 hover:text-[var(--text-primary)]';

const iconDangerButton =
  'inline-flex h-9 w-9 items-center justify-center rounded-full border border-transparent text-[var(--text-secondary)] transition duration-150 hover:border-rose-500/40 hover:bg-rose-500/10 hover:text-rose-400';

const selectClass =
  'w-full md:w-auto rounded-full border border-[var(--border)] bg-[var(--bg-secondary)] ' +
  'h-11 pl-4 pr-10 appearance-none text-sm text-[var(--text-primary)] ' +
  'focus:border-[var(--accent)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40 ' +
  'disabled:cursor-not-allowed disabled:opacity-60';

const rangeInputWrapperClass =
  'flex w-full items-center gap-3 rounded-2xl border border-[var(--border)] bg-[var(--bg-secondary)]/70 px-4 py-3 ' +
  'shadow-sm transition duration-150 ease-out md:flex-1 focus-within:border-[var(--accent)] focus-within:bg-[var(--bg-primary)]/80 focus-within:shadow-lg';

const rangeInputLabelClass =
  'text-[10px] font-semibold uppercase tracking-[0.18em] text-[var(--text-secondary)]';

const rangeDateInputClass =
  'w-full appearance-none border-0 bg-transparent p-0 text-sm font-medium text-[var(--text-primary)] ' +
  'focus:outline-none focus:ring-0';

const categoryFilterChipClass =
  'inline-flex items-center gap-3 rounded-2xl border border-[var(--border)] bg-[var(--bg-secondary)]/80 px-4 py-2 ' +
  'shadow-sm backdrop-blur transition duration-150 ease-out';

const sortChoices = [
  { value: 'dateDesc', label: 'Date (Newest)' },
  { value: 'dateAsc', label: 'Date (Oldest)' },
  { value: 'amountDesc', label: 'Amount (High-Low)' },
  { value: 'amountAsc', label: 'Amount (Low-High)' },
  { value: 'nameAsc', label: 'Name (A-Z)' },
  { value: 'nameDesc', label: 'Name (Z-A)' },
];

const rangeValidationMessage = computed(() => {
  if (dateFilter.value !== 'range') return '';
  if (!rangeStart.value || !rangeEnd.value) {
    return 'Select both start and end dates';
  }
  if (new Date(rangeStart.value) > new Date(rangeEnd.value)) {
    return 'Start date must be before end date';
  }
  return '';
});

function ensureCurrentMonthAvailable() {
  // noop placeholder for future adjustments
}

function sortExpenses(expenses) {
  const sorted = [...expenses];
  switch (sortOption.value) {
    case 'dateAsc':
      sorted.sort((a, b) => new Date(a.date) - new Date(b.date));
      break;
    case 'amountDesc':
      sorted.sort((a, b) => Math.abs(b.amount) - Math.abs(a.amount));
      break;
    case 'amountAsc':
      sorted.sort((a, b) => Math.abs(a.amount) - Math.abs(b.amount));
      break;
    case 'nameAsc':
      sorted.sort((a, b) => (a.name || '').localeCompare(b.name || ''));
      break;
    case 'nameDesc':
      sorted.sort((a, b) => (b.name || '').localeCompare(a.name || ''));
      break;
    default:
      sorted.sort((a, b) => new Date(b.date) - new Date(a.date));
  }
  return sorted;
}

function alignCurrentCycle(baseDate = currentDate.value, { updateCurrent = true, updateCursor = true } = {}) {
  const aligned = getCycleAnchor(baseDate, state.startDate, state.endOfMonth);
  if (updateCursor) {
    monthCursor.value = new Date(aligned);
  }
  if (updateCurrent) {
    currentDate.value = new Date(aligned);
  }
  return new Date(aligned);
}

function normalizeCategory(category) {
  return (category && category.trim()) || 'Uncategorized';
}

function createDefaultForm() {
  const today = new Date();
  const year = today.getFullYear();
  const month = String(today.getMonth() + 1).padStart(2, '0');
  const day = String(today.getDate()).padStart(2, '0');
  return {
    name: '-',
    category: '',
    tags: [],
    amount: null,
    date: `${year}-${month}-${day}`,
    reportGain: false,
    submitLabel: 'Add Expense',
  };
}

function formatDateForInput(date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function resetForm() {
  form.value = createDefaultForm();
  editId.value = null;
  rawAmount.value = '';
}

async function scrollToAddSection(target) {
  await nextTick();
  let element = null;
  if (target === 'manual') {
    element = manualCardRef.value;
  } else if (typingCardRef.value) {
    element =
      typeof typingCardRef.value.getContainer === 'function'
        ? typingCardRef.value.getContainer()
        : typingCardRef.value.$el ?? typingCardRef.value;
  }
  if (element && typeof element.scrollIntoView === 'function') {
    element.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }
}

function handleOpenManual() {
  resetForm();
  showTypingForm.value = false;
  showExpenseForm.value = true;
  scrollToAddSection('manual');
}

async function handleOpenTyping() {
  showExpenseForm.value = false;
  showTypingForm.value = true;
  await nextTick();
  typingCardRef.value?.reset?.();
  scrollToAddSection('typing');
}

function handleCloseAddPanels() {
  showExpenseForm.value = false;
  showTypingForm.value = false;
  resetForm();
  typingCardRef.value?.reset?.();
}

async function handleQuickAddSuccess() {
  await refreshExpenses();
}

function gotoPrevMonth() {
  if (showAll.value || dateFilter.value !== 'month') return;
  const date = new Date(monthCursor.value);
  date.setMonth(date.getMonth() - 1);
  alignCurrentCycle(date);
}

function gotoNextMonth() {
  if (showAll.value || dateFilter.value !== 'month') return;
  const date = new Date(monthCursor.value);
  date.setMonth(date.getMonth() + 1);
  alignCurrentCycle(date);
}

function setFormMessage(text, type) {
  formMessage.value = { text, type };
  if (text) {
    setTimeout(() => {
      formMessage.value = { text: '', type: '' };
    }, 3000);
  }
}

function formatCurrency(amount) {
  return formatCurrencyRaw(amount, state.currency);
}

function formatDate(date) {
  return formatDateFromUTC(date);
}

function editExpense(expense) {
  editId.value = expense.id;
  form.value = {
    name: expense.name,
    category: expense.category,
    tags: [...(expense.tags || [])],
    amount: Math.abs(expense.amount),
    date: toLocalDate(expense.date),
    reportGain: expense.amount > 0,
    submitLabel: 'Update Expense',
  };
  rawAmount.value = String(Math.abs(expense.amount));
  showTypingForm.value = false;
  showExpenseForm.value = true;
  scrollToAddSection('manual');
}

function toLocalDate(isoDate) {
  const date = new Date(isoDate);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function handleAmountInput(event) {
  rawAmount.value = event.target.value.replace(/[^0-9.-]/g, '');
}

function normalizeAmount(event) {
  const numeric = Number(rawAmount.value.replace(/[^0-9.-]/g, '')) || 0;
  rawAmount.value = numeric === 0 ? '' : String(numeric);
  event.target.value = formattedAmount.value;
}

async function submitExpense() {
  if (!form.value.category) {
    setFormMessage('Please select a category', 'error');
    return;
  }
  let amount = Number(rawAmount.value.replace(/[^0-9.-]/g, ''));
  if (Number.isNaN(amount) || amount === 0) {
    setFormMessage('Please enter a valid amount', 'error');
    return;
  }
  if (!form.value.reportGain) {
    amount *= -1;
  }
  const payload = {
    name: form.value.name,
    category: form.value.category,
    amount,
    date: getISODateWithLocalTime(form.value.date),
    tags: form.value.tags,
  };
  const blob = await encryptPayload(payload);
  if (blob) {
    payload.blob = blob;
  }
  const url = editId.value ? `/expense/edit?id=${editId.value}` : '/expense';
  try {
    const response = await apiFetch(url, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to save expense');
    }
    await refreshExpenses();
    setFormMessage(editId.value ? 'Expense updated successfully!' : 'Expense added successfully!', 'success');
    handleCloseAddPanels();
  } catch (error) {
    console.error('Failed to save expense', error);
    setFormMessage(error.message || 'Failed to save expense', 'error');
  }
}

function openDeleteModal(expense, event) {
  expenseToDelete.value = expense;
  if (event?.shiftKey) {
    confirmDelete();
    return;
  }
  showDeleteModal.value = true;
}

function closeDeleteModal() {
  showDeleteModal.value = false;
  expenseToDelete.value = null;
}

async function confirmDelete() {
  if (!expenseToDelete.value) return;
  try {
    const response = await apiFetch(`/expense/delete?id=${expenseToDelete.value.id}`, {
      method: 'DELETE',
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to delete expense');
    }
    await refreshExpenses();
    closeDeleteModal();
  } catch (error) {
    console.error('Failed to delete expense', error);
    setFormMessage(error.message || 'Failed to delete expense', 'error');
  }
}
</script>
