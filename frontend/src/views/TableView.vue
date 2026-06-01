<template>
  <section class="space-y-8 animate-in fade-in duration-700">
    <!-- Period Navigation -->
    <div class="flex flex-col items-center gap-4 md:flex-row md:justify-between">
      <div v-if="showAll" class="text-2xl font-black tracking-tight text-[var(--text-primary)]">All Transactions</div>
      <div v-else-if="dateFilter === 'month'" class="flex items-center gap-1 bg-[var(--bg-secondary)] p-1 rounded-2xl border border-[var(--border)] shadow-sm">
        <button class="flex h-10 w-10 items-center justify-center rounded-xl transition-all hover:bg-[var(--bg-elevated)] active:scale-90" @click="gotoPrevMonth">
          <i class="fa-solid fa-chevron-left text-sm"></i>
        </button>
        <div class="px-6 text-lg font-bold tracking-tight text-[var(--text-primary)] text-center min-w-[180px]">{{ periodLabel }}</div>
        <button class="flex h-10 w-10 items-center justify-center rounded-xl transition-all hover:bg-[var(--bg-elevated)] active:scale-90" @click="gotoNextMonth">
          <i class="fa-solid fa-chevron-right text-sm"></i>
        </button>
      </div>
      <div v-else class="text-2xl font-black tracking-tight text-[var(--text-primary)]">{{ periodLabel }}</div>

      <div class="flex items-center gap-3">
        <div v-if="userDisplayName" class="hidden md:flex items-center gap-3 glass-card rounded-2xl px-4 py-2 text-sm font-semibold text-[var(--text-primary)]">
          <div class="h-6 w-6 bg-gradient-to-br from-indigo-500 to-violet-500 rounded-full flex items-center justify-center text-white text-[10px]">
             <i class="fa-solid fa-user"></i>
          </div>
          <span>{{ userDisplayName }}</span>
        </div>
        <AddExpenseSpeedDial
          :manual-open="showExpenseForm"
          :typing-open="showTypingForm"
          :primary-button-class="'btn-primary'"
          :speed-dial-button-class="'flex items-center gap-2 rounded-xl border border-[var(--border)] bg-[var(--bg-secondary)] px-4 py-2 text-xs font-bold transition-all hover:bg-[var(--accent)] hover:text-white'"
          @open-manual="handleOpenManual"
          @open-typing="handleOpenTyping"
          @close-all="handleCloseAddPanels"
        />
      </div>
    </div>

    <!-- Add/Edit Form -->
    <div v-if="showExpenseForm" id="addExpenseContainer" ref="manualCardRef" class="animate-in slide-in-from-top duration-500">
      <div class="glass-card p-6 md:p-8 rounded-3xl relative overflow-hidden">
        <div class="absolute top-0 left-0 w-2 h-full bg-[var(--accent)]"></div>
        <div class="flex items-center gap-3 mb-6 ml-2">
           <div class="h-10 w-10 bg-indigo-500/20 text-indigo-500 rounded-xl flex items-center justify-center">
              <i class="fa-solid" :class="editId ? 'fa-pen-to-square' : 'fa-file-circle-plus'"></i>
           </div>
           <h3 class="text-xl font-black tracking-tight text-[var(--text-primary)]">{{ editId ? 'Update Transaction' : 'Add Transaction' }}</h3>
        </div>

        <form class="grid gap-6 md:grid-cols-2 ml-2" @submit.prevent="submitExpense">
          <div class="space-y-2">
            <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1" for="name">Name</label>
            <input id="name" v-model="form.name" type="text" class="input-modern w-full" placeholder="Transaction description" required />
          </div>
          <div class="space-y-2">
            <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1" for="category">Category</label>
            <select id="category" v-model="form.category" class="input-modern w-full appearance-none" required>
              <option value="" disabled>Choose category</option>
              <option v-for="category in state.categories" :key="category" :value="category">
                {{ category }}
              </option>
            </select>
          </div>
          <div class="space-y-2">
            <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1">Tags</label>
            <TagInput v-model="form.tags" :suggestions="state.tags" />
          </div>
          <div class="space-y-2">
            <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1" for="amount">Amount</label>
            <div class="relative">
               <span class="absolute left-4 top-1/2 -translate-y-1/2 text-[var(--text-secondary)] font-bold text-sm">IDR</span>
               <input
                 id="amount"
                 :value="formattedAmount"
                 inputmode="decimal"
                 class="input-modern w-full pl-12 text-lg font-black tabular-nums text-[var(--text-primary)]"
                 required
                 @input="handleAmountInput"
                 @blur="normalizeAmount"
               />
            </div>
          </div>
          <div class="space-y-2">
            <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1" for="date">Date</label>
            <input id="date" v-model="form.date" type="date" class="input-modern w-full text-[var(--text-primary)]" required />
          </div>
          <div class="flex flex-col justify-end">
            <label class="flex items-center gap-3 cursor-pointer group mb-2">
              <input v-model="form.reportGain" type="checkbox" class="sr-only peer" />
              <div class="h-6 w-11 bg-slate-700 rounded-full peer peer-checked:bg-emerald-500 transition-colors relative after:content-[''] after:absolute after:top-1 after:left-1 after:h-4 after:w-4 after:bg-white after:rounded-full after:transition-transform peer-checked:after:translate-x-5"></div>
              <span class="text-sm font-bold text-[var(--text-secondary)] group-hover:text-[var(--text-primary)] transition-colors">Income / Gain</span>
            </label>
          </div>
          <div class="md:col-span-2 pt-4 flex gap-3">
            <button v-if="editId" type="button" class="btn-secondary px-8 py-3 rounded-2xl border border-[var(--border)] text-sm font-bold text-[var(--text-primary)]" @click="handleCloseAddPanels">Cancel</button>
            <button type="submit" class="btn-primary flex-1 shadow-indigo-500/20 shadow-2xl h-14 text-lg">{{ form.submitLabel }}</button>
          </div>
        </form>
        
        <transition name="fade">
          <div v-if="formMessage.text" class="mt-6 p-4 rounded-2xl text-center text-sm font-bold ml-2" :class="[formMessage.type === 'success' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-rose-500/10 text-rose-400 border border-rose-500/20']">
            {{ formMessage.text }}
          </div>
        </transition>
      </div>
    </div>

    <QuickAddExpenseCard
      v-if="showTypingForm"
      ref="typingCardRef"
      :card-class="'glass-card p-6 md:p-8 rounded-3xl animate-in slide-in-from-top duration-500'"
      :input-class="'input-modern w-full h-14 text-lg font-medium text-[var(--text-primary)]'"
      :primary-button-class="'btn-primary h-14 px-8'"
      @switch-manual="handleOpenManual"
      @added="handleQuickAddSuccess"
    />

    <!-- Filters & Table Section -->
    <div class="space-y-6">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div class="flex flex-wrap items-center gap-3">
          <!-- Date Filter -->
          <div class="relative group">
            <select v-model="dateFilter" :disabled="showAll" class="input-modern min-w-[150px] appearance-none pr-10 cursor-pointer text-[var(--text-primary)] disabled:opacity-30">
              <option value="month">Monthly</option>
              <option value="week">Weekly</option>
              <option value="today">Today</option>
              <option value="yesterday">Yesterday</option>
              <option value="range">Range</option>
            </select>
            <i class="fa-solid fa-angle-down absolute right-4 top-1/2 -translate-y-1/2 text-[var(--text-secondary)] pointer-events-none group-hover:text-[var(--accent)] transition-colors"></i>
          </div>

          <!-- Sort Filter -->
          <div class="relative group">
            <select v-model="sortOption" class="input-modern min-w-[150px] appearance-none pr-10 cursor-pointer text-[var(--text-primary)]">
              <option v-for="option in sortChoices" :key="option.value" :value="option.value">
                {{ option.label }}
              </option>
            </select>
            <i class="fa-solid fa-arrow-down-wide-short absolute right-4 top-1/2 -translate-y-1/2 text-[var(--text-secondary)] pointer-events-none group-hover:text-[var(--accent)] transition-colors"></i>
          </div>

          <!-- Show All Toggle -->
          <label class="flex items-center gap-3 cursor-pointer group bg-[var(--bg-secondary)] px-4 py-2.5 rounded-2xl border border-[var(--border)] shadow-sm">
            <input v-model="showAll" type="checkbox" class="sr-only peer" />
            <div class="h-5 w-9 bg-slate-700 rounded-full peer peer-checked:bg-[var(--accent)] transition-colors relative after:content-[''] after:absolute after:top-1 after:left-1 after:h-3 after:w-3 after:bg-white after:rounded-full after:transition-transform peer-checked:after:translate-x-4"></div>
            <span class="text-xs font-bold text-[var(--text-secondary)] group-hover:text-[var(--text-primary)] transition-colors">Show All</span>
          </label>
        </div>

        <div v-if="selectedCategory" class="flex items-center gap-3 bg-[var(--accent)]/10 text-[var(--accent)] px-4 py-2 rounded-2xl border border-[var(--accent)]/20 animate-in zoom-in duration-300">
          <i class="fa-solid fa-tag text-xs"></i>
          <span class="text-xs font-black uppercase tracking-tight">{{ selectedCategory }}</span>
          <button @click="clearCategoryFilter" class="hover:bg-[var(--accent)]/20 p-1 rounded-lg transition-colors"><i class="fa-solid fa-xmark"></i></button>
        </div>
      </div>

      <div v-if="dateFilter === 'range' && !showAll" class="flex flex-wrap items-center gap-3 animate-in fade-in duration-500">
          <div class="flex items-center gap-2 bg-[var(--bg-secondary)] px-4 py-2 rounded-2xl border border-[var(--border)] shadow-sm">
             <span class="text-[10px] font-bold text-[var(--text-secondary)] uppercase">From</span>
             <input v-model="rangeStart" type="date" class="bg-transparent text-sm font-bold text-[var(--text-primary)] focus:outline-none" />
          </div>
          <div class="flex items-center gap-2 bg-[var(--bg-secondary)] px-4 py-2 rounded-2xl border border-[var(--border)] shadow-sm">
             <span class="text-[10px] font-bold text-[var(--text-secondary)] uppercase">To</span>
             <input v-model="rangeEnd" type="date" class="bg-transparent text-sm font-bold text-[var(--text-primary)] focus:outline-none" />
          </div>
          <p v-if="rangeValidationMessage" class="text-xs font-bold text-rose-400 ml-2">{{ rangeValidationMessage }}</p>
      </div>

      <!-- Transactions List -->
      <div v-if="tableExpenses.length === 0" class="w-full py-24 text-center glass-card rounded-3xl border-dashed">
         <i class="fa-solid fa-folder-open text-4xl text-[var(--text-secondary)]/20 mb-4 block"></i>
         <p class="text-lg italic text-[var(--text-secondary)]">{{ emptyTableMessage }}</p>
      </div>

      <div v-else class="glass-card rounded-3xl overflow-hidden shadow-2xl transition-all border-none">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead>
              <tr class="bg-[var(--bg-elevated)]/50 text-[8px] md:text-[10px] font-black uppercase tracking-widest text-[var(--text-secondary)] border-b border-[var(--border)]">
                <th class="px-2 md:px-8 py-3 md:py-5">Name / Info</th>
                <th class="px-2 md:px-6 py-3 md:py-5">Category</th>
                <th v-if="hasTags" class="px-2 md:px-6 py-3 md:py-5">Tags</th>
                <th class="px-2 md:px-6 py-3 md:py-5">Amount</th>
                <th class="px-2 md:px-6 py-3 md:py-5">Date</th>
                <th class="px-2 md:px-8 py-3 md:py-5 text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-[var(--border)]">
              <tr v-for="expense in tableExpenses" :key="expense.id" class="group transition-colors hover:bg-[var(--bg-secondary)]/40">
                <td class="px-2 md:px-8 py-3 md:py-5 max-w-[120px] md:max-w-none">
                   <div class="flex flex-col min-w-0">
                      <span class="font-bold text-[var(--text-primary)] group-hover:text-[var(--accent)] transition-colors truncate">{{ expense.name }}</span>
                      <span class="text-[10px] font-medium text-[var(--text-secondary)] opacity-0 group-hover:opacity-100 transition-opacity truncate">ID: {{ expense.id.slice(0,8) }}</span>
                   </div>
                </td>
                <td class="px-2 md:px-6 py-3 md:py-5 max-w-[90px] md:max-w-none">
                   <span class="inline-block max-w-full px-2 md:px-3 py-1 rounded-full bg-[var(--bg-elevated)] text-[10px] font-black uppercase tracking-tight text-[var(--text-secondary)] border border-[var(--border)] truncate">{{ expense.category }}</span>
                </td>
                <td v-if="hasTags" class="px-3 md:px-6 py-3 md:py-5">
                   <div class="flex flex-wrap gap-1">
                      <span v-for="tag in expense.tags" :key="tag" class="text-[9px] font-bold text-[var(--accent)]">#{{ tag }}</span>
                   </div>
                </td>
                <td class="px-2 md:px-6 py-3 md:py-5">
                   <span class="font-black tabular-nums text-xs md:text-sm whitespace-nowrap" :class="expense.amount < 0 ? 'text-[var(--text-primary)]' : 'text-emerald-500'">{{ formatCurrency(expense.amount) }}</span>
                </td>
                <td class="px-2 md:px-6 py-3 md:py-5">
                   <span class="text-xs font-bold text-[var(--text-secondary)] whitespace-nowrap">{{ formatDate(expense.date) }}</span>
                </td>
                <td class="px-2 md:px-8 py-3 md:py-5 text-right space-x-1">
                  <button @click="editExpense(expense)" class="h-9 w-9 inline-flex items-center justify-center rounded-xl transition-all hover:bg-indigo-500/10 hover:text-indigo-500 active:scale-90 text-[var(--text-secondary)]" title="Edit">
                    <i class="fa-solid fa-pen-to-square text-sm"></i>
                  </button>
                  <button @click="openDeleteModal(expense, $event)" class="h-9 w-9 inline-flex items-center justify-center rounded-xl transition-all hover:bg-rose-500/10 hover:text-rose-500 active:scale-90 text-[var(--text-secondary)]" title="Delete">
                    <i class="fa-solid fa-trash-can text-sm"></i>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <transition name="fade">
      <div v-if="showDeleteModal" class="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-sm px-4" @click.self="closeDeleteModal">
        <div class="w-full max-w-sm glass-card rounded-3xl p-8 animate-in zoom-in duration-300">
          <div class="h-16 w-16 bg-rose-500/20 text-rose-500 rounded-2xl flex items-center justify-center mx-auto mb-6">
             <i class="fa-solid fa-triangle-exclamation text-3xl"></i>
          </div>
          <h3 class="text-xl font-black text-center tracking-tight text-[var(--text-primary)]">Delete Transaction?</h3>
          <p class="mt-3 text-sm text-[var(--text-secondary)] text-center leading-relaxed font-medium">
            This action will permanently remove this record. It's not reversible bro.
          </p>
          <div class="mt-8 grid grid-cols-2 gap-4">
            <button class="px-6 py-3 rounded-2xl bg-[var(--bg-elevated)] font-bold text-sm hover:brightness-110 transition-all text-[var(--text-primary)]" @click="closeDeleteModal">Nah, keep it</button>
            <button class="px-6 py-3 rounded-2xl bg-rose-600 text-white font-bold text-sm shadow-xl shadow-rose-600/20 hover:brightness-110 active:scale-95 transition-all" @click="confirmDelete">Delete it</button>
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
    const rect = element.getBoundingClientRect();
    const offset = Math.max(window.pageYOffset + rect.top - 80, 0);
    window.scrollTo({ top: offset, behavior: 'smooth' });
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
    const wasEditing = Boolean(editId.value);
    setFormMessage(wasEditing ? 'Expense updated successfully!' : 'Expense added successfully!', 'success');
    if (wasEditing) {
      handleCloseAddPanels();
    } else {
      resetForm();
    }
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

<style scoped>
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes zoomIn {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}

@keyframes slideInDown {
  from { transform: translateY(-20px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}

.animate-in {
  animation-fill-mode: both;
}
.fade-in { animation-name: fadeIn; }
.zoom-in { animation-name: zoomIn; }
.slide-in-from-top { animation-name: slideInDown; }

.fade-enter-active, .fade-leave-active {
  transition: all 0.3s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
  backdrop-filter: blur(0);
}

.btn-secondary {
  @apply bg-transparent hover:bg-[var(--bg-elevated)] transition-all active:scale-95;
}

::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}
::-webkit-scrollbar-track {
  background: transparent;
}
::-webkit-scrollbar-thumb {
  background: var(--border);
  border-radius: 10px;
}
::-webkit-scrollbar-thumb:hover {
  background: var(--text-secondary);
}
</style>
