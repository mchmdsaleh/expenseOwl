<template>
  <section class="space-y-6">
    <div class="flex flex-col items-center gap-3 md:flex-row md:items-center md:justify-between">
      <div v-if="dateFilter === 'month'" class="flex items-center justify-center gap-4">
        <button
          :class="[iconButtonClass, dateFilter !== 'month' && 'pointer-events-none opacity-50']"
          :disabled="dateFilter !== 'month'"
          @click="gotoPrevMonth"
        >
          <i class="fa-solid fa-arrow-left"></i>
        </button>
        <div class="min-w-[200px] text-center text-2xl font-bold">{{ periodLabel }}</div>
        <button
          :class="[iconButtonClass, dateFilter !== 'month' && 'pointer-events-none opacity-50']"
          :disabled="dateFilter !== 'month'"
          @click="gotoNextMonth"
        >
          <i class="fa-solid fa-arrow-right"></i>
        </button>
      </div>
      <div v-else class="min-w-[200px] text-center text-2xl font-bold">{{ periodLabel }}</div>
      <div
        v-if="userDisplayName"
        class="inline-flex w-full items-center justify-center gap-2 rounded-2xl border border-[var(--border)] bg-[var(--bg-secondary)]/70 px-4 py-2 text-sm font-medium text-[var(--text-primary)] md:w-auto md:justify-start"
      >
        <i class="fa-solid fa-circle-user text-[var(--accent)]"></i>
        <span>{{ userDisplayName }}</span>
      </div>
    </div>

    <div class="flex flex-wrap items-center justify-between gap-2">
      <div class="flex flex-wrap items-center gap-2">
        <div class="relative">
          <select v-model="dateFilter" :class="filterSelectClass">
            <option value="month">This Month</option>
            <option value="week">This Week</option>
            <option value="today">Today</option>
            <option value="yesterday">Yesterday</option>
            <option value="range">Custom Range</option>
          </select>
          <i class="fa-solid fa-chevron-down absolute right-4 top-1/2 -translate-y-1/2 text-[var(--text-secondary)] pointer-events-none text-xs"></i>
        </div>
        <div
          v-if="dateFilter === 'range'"
          class="flex w-full flex-wrap items-center gap-2 md:ml-4 md:w-auto md:flex-nowrap"
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
      </div>
      <div class="shrink-0">
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
    <p v-if="rangeValidationMessage" class="text-xs italic text-amber-300">{{ rangeValidationMessage }}</p>

    <div v-if="showExpenseForm" id="addExpenseContainer" ref="manualCardRef">
      <div :class="cardClass">
        <form class="grid gap-4 md:grid-cols-2" @submit.prevent="submitExpense">
          <div class="flex flex-col gap-2">
            <label class="text-sm font-medium text-[var(--text-secondary)]" for="name">Name</label>
            <input id="name" v-model="form.name" type="text" :class="inputClass" required />
          </div>
          <div class="flex flex-col gap-2">
            <label class="text-sm font-medium text-[var(--text-secondary)]" for="category">Category</label>
            <select id="category" v-model="form.category" :class="inputClass" required>
              <option value="" disabled>Choose category</option>
              <option v-for="category in categories" :key="category" :value="category">
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

    <div class="flex flex-col gap-6 rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)]/80 p-6 shadow-card backdrop-blur lg:flex-row">
      <div v-if="!hasExpenseData" class="w-full rounded-3xl border border-dashed border-[var(--border)] bg-[var(--bg-secondary)]/60 py-12 text-center text-base italic text-[var(--text-secondary)]">
        {{ emptyDashboardMessage }}
      </div>
      <template v-else>
        <div
          class="flex h-80 flex-1 cursor-pointer items-center justify-center"
          role="button"
          aria-label="View detailed transactions"
          @click="handleChartClick"
        >
          <canvas ref="chartCanvas"></canvas>
        </div>
        <div class="flex flex-1 flex-col gap-4">
          <div
            v-for="entry in legendEntries"
            :key="entry.category"
            :class="[
              legendItemClass,
              entry.disabled && 'opacity-40'
            ]"
            @click="toggleCategory(entry.category)"
          >
            <div class="h-4 w-4 rounded-md" :style="{ backgroundColor: entry.color }"></div>
            <div class="flex flex-1 items-center justify-between gap-3 text-sm text-[var(--text-secondary)]">
              <span>{{ entry.category }}<template v-if="entry.percentage !== null"> ({{ entry.percentage.toFixed(1) }}%)</template></span>
              <span class="font-mono text-sm text-[var(--text-secondary)]" v-if="entry.amount !== null">{{ entry.amountFormatted }}</span>
            </div>
          </div>
          <div class="mt-2 flex items-center justify-between rounded-2xl border border-[var(--border)] bg-[var(--bg-primary)]/60 px-4 py-3">
            <span class="text-sm font-medium text-[var(--text-secondary)]">Total:</span>
            <span class="font-mono text-base text-[var(--text-primary)]">{{ totalActiveFormatted }}</span>
          </div>
        </div>
      </template>
    </div>

    <div v-if="budgetSummaries.length" :class="cardClass">
      <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <h3 class="text-lg font-semibold text-[var(--text-primary)]">Budget Progress</h3>
        <div class="text-right text-sm text-[var(--text-secondary)]">
          <div class="font-mono text-base text-[var(--text-primary)]">{{ formatCurrency(totalBudgetRemaining) }} remaining</div>
          <div
            v-if="overallBudgetStatus"
            :class="['text-xs font-medium', overallBudgetStatus.className]"
          >
            {{ overallBudgetStatus.label }}
          </div>
        </div>
      </div>
      <div class="mt-5 space-y-4">
        <div
          v-for="item in budgetSummaries"
          :key="item.id"
          class="space-y-3 rounded-2xl border border-[var(--border)] bg-[var(--bg-primary)]/60 p-4"
        >
          <div class="flex items-center justify-between gap-3">
            <div class="text-sm font-semibold text-[var(--text-primary)]">{{ item.category }}</div>
            <div class="text-right text-xs text-[var(--text-secondary)]">
              <div class="font-mono text-sm text-[var(--text-primary)]">
                {{ formatCurrency(item.actual) }} / {{ formatCurrency(item.amount) }}
              </div>
              <div
                :class="[
                  'text-[11px] font-semibold uppercase tracking-wide',
                  item.status === 'over'
                    ? 'text-rose-400'
                    : item.status === 'warning'
                      ? 'text-amber-300'
                      : 'text-emerald-300'
                ]"
              >
                {{ item.statusLabel }}
              </div>
            </div>
          </div>
          <div class="h-2 w-full rounded-full bg-[var(--border)]/70">
            <div
              class="h-2 rounded-full transition-all"
              :class="item.status === 'over' ? 'bg-rose-500' : item.status === 'warning' ? 'bg-amber-400' : 'bg-emerald-500'"
              :style="{ width: `${Math.min(item.percentage, 100).toFixed(1)}%` }"
            ></div>
          </div>
          <div class="flex justify-between text-[11px] text-[var(--text-secondary)]">
            <span>Spent: {{ formatCurrency(item.actual) }}</span>
            <span>Remaining: {{ formatCurrency(item.remaining) }}</span>
          </div>
          <div class="flex flex-wrap gap-3 text-[11px] text-[var(--text-secondary)]">
            <span v-if="item.overrideAmount != null">Override: {{ formatCurrency(item.overrideAmount || 0) }}</span>
            <span v-if="item.adjustmentAmount">Adjustment: {{ formatCurrency(item.adjustmentAmount) }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="hasExpenseData" class="grid gap-4 md:grid-cols-3">
      <div :class="cashflowCardClass">
        <div class="text-sm font-medium text-[var(--text-secondary)]">Income</div>
        <div class="text-2xl font-bold text-emerald-400">{{ formatCurrency(income) }}</div>
      </div>
      <div :class="cashflowCardClass">
        <div class="text-sm font-medium text-[var(--text-secondary)]">Expenses</div>
        <div class="text-2xl font-bold text-rose-400">{{ formatCurrency(totalExpenses) }}</div>
      </div>
      <div :class="cashflowCardClass">
        <div class="text-sm font-medium text-[var(--text-secondary)]">Balance</div>
        <div
          class="text-2xl font-bold"
          :class="balance >= 0 ? 'text-emerald-400' : 'text-rose-400'"
        >
          {{ formatCurrency(balance) }}
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import { Chart, registerables } from 'chart.js';
import state, { loadInitialData, refreshExpenses, refreshBudgetSummaries } from '../stores/appState';
import TagInput from '../components/TagInput.vue';
import AddExpenseSpeedDial from '../components/AddExpenseSpeedDial.vue';
import QuickAddExpenseCard from '../components/QuickAddExpenseCard.vue';
import {
  formatMonth,
  getMonthExpenses,
  filterExpensesByRange,
  formatCurrency as formatCurrencyRaw,
  getISODateWithLocalTime,
  colorPalette,
  formatWeekRange,
  formatDayLabel,
  formatRangeLabel,
  getCycleAnchor,
} from '../lib/utils';
import { apiFetch } from '../lib/api';
import { encryptPayload } from '../lib/encryption';

Chart.register(...registerables);
Chart.defaults.color = '#b3b3b3';
Chart.defaults.borderColor = '#606060';
Chart.defaults.font.family = '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif';

const chartCanvas = ref(null);
let chartInstance = null;
const router = useRouter();

const currentDate = ref(new Date());
const monthCursor = ref(new Date());
const dateFilter = ref('month');
const showExpenseForm = ref(false);
const disabledCategories = ref(new Set());
const categoryColors = ref({});
const rangeStart = ref(formatDateForInput(daysAgo(6)));
const rangeEnd = ref(formatDateForInput(new Date()));

const form = ref(createDefaultForm());
const formMessage = ref({ text: '', type: '' });
const rawAmount = ref('');
const showTypingForm = ref(false);
const manualCardRef = ref(null);
const typingCardRef = ref(null);

const userDisplayName = computed(() => {
  const first = (state.user?.firstName || '').trim();
  const last = (state.user?.lastName || '').trim();
  const full = `${first} ${last}`.trim();
  return full || state.user?.email || '';
});

const iconButtonClass =
  'inline-flex h-11 w-11 items-center justify-center rounded-full border border-[var(--border)] bg-[var(--bg-secondary)] text-lg text-[var(--text-primary)] transition duration-150 ease-out hover:bg-[var(--accent)] hover:text-white hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40 focus:ring-offset-2 focus:ring-offset-[var(--bg-primary)]';

const primaryButtonClass =
  'inline-flex items-center justify-center gap-2 rounded-full border border-[var(--border)] bg-[var(--bg-secondary)] px-5 py-2 text-sm font-medium text-[var(--text-primary)] transition duration-150 ease-out hover:bg-[var(--accent)] hover:text-white hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40 focus:ring-offset-2 focus:ring-offset-[var(--bg-primary)] disabled:cursor-not-allowed disabled:opacity-50';

const inputClass =
  'w-full rounded-xl border border-[var(--border)] bg-[var(--bg-primary)] px-4 py-2 text-[var(--text-primary)] placeholder:text-[var(--text-secondary)] focus:border-[var(--accent)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40';

const checkboxClass =
  'h-4 w-4 rounded border-[var(--border)] bg-[var(--bg-primary)] text-[var(--accent)] focus:ring-[var(--accent)]/60 focus:ring-offset-0';

const filterSelectClass =
  'min-w-[180px] w-auto rounded-full border border-[var(--border)] bg-[var(--bg-secondary)] ' +
  'h-11 pl-4 pr-12 appearance-none text-sm text-[var(--text-primary)] ' +
  'focus:border-[var(--accent)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40';

const rangeInputWrapperClass =
  'flex w-full items-center gap-3 rounded-2xl border border-[var(--border)] bg-[var(--bg-secondary)]/70 px-4 py-3 ' +
  'shadow-sm transition duration-150 ease-out md:flex-1 focus-within:border-[var(--accent)] focus-within:bg-[var(--bg-primary)]/80 focus-within:shadow-lg';

const rangeInputLabelClass =
  'text-[10px] font-semibold uppercase tracking-[0.18em] text-[var(--text-secondary)]';

const rangeDateInputClass =
  'w-full appearance-none border-0 bg-transparent p-0 text-sm font-medium text-[var(--text-primary)] ' +
  'focus:outline-none focus:ring-0';

const cardClass =
  'rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)]/80 p-6 shadow-card backdrop-blur';

const legendItemClass =
  'flex items-center gap-4 rounded-2xl border border-transparent px-3 py-2 text-sm transition duration-150 ease-out hover:bg-[var(--bg-primary)]/60';

const cashflowCardClass =
  'flex flex-col items-center justify-center rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)]/80 px-6 py-6 text-center shadow-card backdrop-blur';

const speedDialButtonClass =
  'flex items-center gap-2 rounded-full border border-[var(--border)] bg-[var(--bg-secondary)]/90 px-4 py-2 text-sm font-medium text-[var(--text-primary)] transition duration-150 ease-out hover:bg-[var(--accent)] hover:text-white hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40 focus:ring-offset-2 focus:ring-offset-[var(--bg-primary)]';

const categories = computed(() => state.categories);
const baseBudgets = computed(() => state.budgets || []);
const monthlyBudgets = computed(() => {
  const summaries = Array.isArray(state.budgetSummaries) ? state.budgetSummaries : [];
  if (summaries.length > 0) {
    return summaries.map((summary) => ({
      id: summary.id,
      category: summary.category,
      baseAmount: summary.amount,
      effectiveAmount: summary.effectiveAmount ?? summary.amount,
      overrideAmount: summary.overrideAmount ?? null,
      adjustmentAmount: summary.adjustmentAmount ?? 0,
    }));
  }
  return baseBudgets.value.map((budget) => ({
    id: budget.id,
    category: budget.category,
    baseAmount: budget.amount,
    effectiveAmount: budget.amount,
    overrideAmount: null,
    adjustmentAmount: 0,
  }));
});
const formattedAmount = computed(() => {
  if (!rawAmount.value) return '';
  const numeric = Number(rawAmount.value.replace(/[^0-9.-]/g, '')) || 0;
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(numeric);
});

const monthExpenses = computed(() => getMonthExpenses(state.expenses, monthCursor.value, state.startDate, state.endOfMonth));
const displayedExpenses = computed(() => {
  if (dateFilter.value === 'month') {
    return monthExpenses.value;
  }
  if (dateFilter.value === 'range') {
    return filterExpensesByRange(
      state.expenses,
      'range',
      currentDate.value,
      state.startDate,
      state.endOfMonth,
      { start: rangeStart.value, end: rangeEnd.value }
    );
  }
  return filterExpensesByRange(state.expenses, dateFilter.value, currentDate.value, state.startDate, state.endOfMonth);
});
const hasExpenseData = computed(() => displayedExpenses.value.some((expense) => expense.amount < 0));

const income = computed(() => displayedExpenses.value.filter((exp) => exp.amount > 0).reduce((sum, exp) => sum + exp.amount, 0));
const totalExpenses = computed(() => displayedExpenses.value.filter((exp) => exp.amount < 0).reduce((sum, exp) => sum + Math.abs(exp.amount), 0));
const balance = computed(() => income.value - totalExpenses.value);

const periodLabel = computed(() => {
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

const emptyDashboardMessage = computed(() => {
  if (dateFilter.value === 'yesterday') {
    return 'No expenses recorded yesterday.';
  }
  if (dateFilter.value === 'range') {
    return 'No expenses recorded in this range.';
  }
  if (dateFilter.value === 'today') {
    return 'No expenses recorded today.';
  }
  if (dateFilter.value === 'week') {
    return 'No expenses recorded this week.';
  }
  return 'No expenses recorded this month.';
});

const legendEntries = computed(() => buildLegendEntries());
const totalActiveExpenses = computed(() => {
  return displayedExpenses.value
    .filter((exp) => {
      if (exp.amount >= 0) return false;
      const category = normalizeCategoryName(exp.category);
      return !disabledCategories.value.has(category);
    })
    .reduce((sum, exp) => sum + Math.abs(exp.amount), 0);
});
const totalActiveFormatted = computed(() => formatCurrency(totalActiveExpenses.value));

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

const budgetSummaries = computed(() => {
  if (!monthlyBudgets.value.length) return [];
  const totals = monthExpenses.value.reduce((acc, exp) => {
    if (exp.amount < 0) {
      const key = exp.category || 'Uncategorized';
      acc[key] = (acc[key] || 0) + Math.abs(exp.amount);
    }
    return acc;
  }, {});
  return monthlyBudgets.value
    .map((budget) => {
      const limit = budget.effectiveAmount ?? budget.baseAmount;
      const actual = totals[budget.category] || 0;
      const remaining = limit - actual;
      const percentage = limit > 0 ? Math.min(100, (actual / limit) * 100) : 0;
      let status = 'ok';
      if (remaining < 0) {
        status = 'over';
      } else if (percentage >= 80) {
        status = 'warning';
      }
      const statusLabel = status === 'over'
        ? 'Over budget'
        : status === 'warning'
          ? 'Approaching limit'
          : 'On track';
      return {
        id: budget.id,
        category: budget.category,
        amount: limit,
        actual,
        remaining,
        percentage,
        status,
        statusLabel,
        overrideAmount: budget.overrideAmount,
        adjustmentAmount: budget.adjustmentAmount,
      };
    })
    .sort((a, b) => a.category.localeCompare(b.category));
});

const totalBudgeted = computed(() => budgetSummaries.value.reduce((sum, item) => sum + item.amount, 0));
const totalBudgetActual = computed(() => budgetSummaries.value.reduce((sum, item) => sum + item.actual, 0));
const totalBudgetRemaining = computed(() => totalBudgeted.value - totalBudgetActual.value);
const overallBudgetStatus = computed(() => {
  if (!budgetSummaries.value.length || totalBudgeted.value === 0) return null;
  if (totalBudgetRemaining.value < 0) {
    return { label: 'Over budget', className: 'text-rose-400' };
  }
  const usageRatio = totalBudgetActual.value / totalBudgeted.value;
  if (usageRatio >= 0.8) {
    return { label: 'Approaching limit', className: 'text-amber-300' };
  }
  return { label: 'On track', className: 'text-emerald-300' };
});

async function loadBudgetsForCurrentMonth() {
  try {
    await refreshBudgetSummaries(monthCursor.value);
  } catch (error) {
    console.error('Failed to load monthly budgets', error);
  }
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

watch(
  () => state.expenses,
  () => {
    assignCategoryColors();
    updateChart();
  }
);

watch(displayedExpenses, () => {
  assignCategoryColors();
  updateChart();
});

watch(
  () => [state.startDate, state.endOfMonth],
  async () => {
    if (dateFilter.value === 'month') {
      alignCurrentCycle(monthCursor.value);
      await loadBudgetsForCurrentMonth();
      assignCategoryColors();
      updateChart();
    } else {
      alignCurrentCycle(monthCursor.value, { updateCurrent: false, updateCursor: true });
    }
  }
);

watch(disabledCategories, updateChart, { deep: true });

watch(
  dateFilter,
  async (next, prev) => {
    if (next === 'month') {
      alignCurrentCycle(monthCursor.value);
      await loadBudgetsForCurrentMonth();
    } else {
      currentDate.value = new Date();
    }
    if (next === 'range' && prev !== 'range') {
      rangeStart.value = formatDateForInput(daysAgo(6));
      rangeEnd.value = formatDateForInput(new Date());
    }
    disabledCategories.value = new Set();
    await nextTick();
    assignCategoryColors();
    updateChart();
  }
);

onMounted(async () => {
  await loadInitialData();
  alignCurrentCycle();
  await loadBudgetsForCurrentMonth();
  monthCursor.value = new Date(currentDate.value);
  assignCategoryColors();
  updateChart();
});

onBeforeUnmount(() => {
  if (chartInstance) {
    chartInstance.destroy();
    chartInstance = null;
  }
});

function daysAgo(amount) {
  const date = new Date();
  date.setDate(date.getDate() - amount);
  return date;
}

function formatDateForInput(date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
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

function resetForm() {
  form.value = createDefaultForm();
  rawAmount.value = '';
}

function setFormMessage(text, type) {
  formMessage.value = { text, type };
  if (text) {
    setTimeout(() => {
      formMessage.value = { text: '', type: '' };
    }, 3000);
  }
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

function normalizeCategoryName(category) {
  return (typeof category === 'string' && category.trim()) || 'Uncategorized';
}

function assignCategoryColors() {
  const colors = { ...categoryColors.value };
  const baseCategories = Array.isArray(state.categories) ? state.categories : [];
  const allCategories = Array.from(
    new Set([
      ...baseCategories.map((category) => normalizeCategoryName(category)),
      ...state.expenses.map((expense) => normalizeCategoryName(expense.category)),
    ]),
  );
  allCategories.forEach((category) => {
    if (!colors[category]) {
      const nextIndex = Object.keys(colors).length;
      colors[category] = colorPalette[nextIndex % colorPalette.length];
    }
  });
  categoryColors.value = colors;
}

async function gotoPrevMonth() {
  if (dateFilter.value !== 'month') return;
  const date = new Date(monthCursor.value);
  date.setMonth(date.getMonth() - 1);
  alignCurrentCycle(date);
  await loadBudgetsForCurrentMonth();
  await nextTick();
  assignCategoryColors();
  updateChart();
}

async function gotoNextMonth() {
  if (dateFilter.value !== 'month') return;
  const date = new Date(monthCursor.value);
  date.setMonth(date.getMonth() + 1);
  alignCurrentCycle(date);
  await loadBudgetsForCurrentMonth();
  await nextTick();
  assignCategoryColors();
  updateChart();
}

function calculateCategoryBreakdown(expenses) {
  const categoryTotals = {};
  let totalAmount = 0;
  expenses.forEach((exp) => {
    if (exp.amount < 0) {
      const amount = Math.abs(exp.amount);
      const category = normalizeCategoryName(exp.category);
      if (disabledCategories.value.has(category)) return;
      categoryTotals[category] = (categoryTotals[category] || 0) + amount;
      totalAmount += amount;
    }
  });
  return Object.entries(categoryTotals)
    .map(([category, total]) => ({
      category,
      total,
      percentage: totalAmount > 0 ? (total / totalAmount) * 100 : 0,
    }))
    .sort((a, b) => b.total - a.total);
}

function buildLegendEntries() {
  const breakdown = calculateCategoryBreakdown(displayedExpenses.value);
  const categoryMap = new Map(breakdown.map((item) => [item.category, item]));
  const currentMonthCategories = Array.from(
    new Set(
      displayedExpenses.value
        .filter((exp) => exp.amount < 0)
        .map((exp) => normalizeCategoryName(exp.category))
    )
  );
  currentMonthCategories.sort((a, b) => {
    const dataA = categoryMap.get(a);
    const dataB = categoryMap.get(b);
    if (dataA && dataB) return dataB.total - dataA.total;
    if (dataA) return -1;
    if (dataB) return 1;
    return a.localeCompare(b);
  });
  return currentMonthCategories.map((category) => {
    const entry = categoryMap.get(category);
    const disabled = disabledCategories.value.has(category);
    return {
      category,
      color: categoryColors.value[category] || '#4ECDC4',
      amount: entry ? entry.total : null,
      amountFormatted: entry ? formatCurrency(entry.total) : '',
      percentage: entry ? entry.percentage : null,
      disabled,
    };
  });
}

function formatCurrency(amount) {
  return formatCurrencyRaw(amount, state.currency);
}

function handleAmountInput(event) {
  const value = event.target.value;
  rawAmount.value = value.replace(/[^0-9.-]/g, '');
}

function normalizeAmount(event) {
  const numeric = Number(rawAmount.value.replace(/[^0-9.-]/g, '')) || 0;
  rawAmount.value = numeric === 0 ? '' : String(numeric);
  event.target.value = formattedAmount.value;
}

async function postExpense(body) {
  const payload = { ...body };
  const blob = await encryptPayload(payload);
  if (blob) {
    payload.blob = blob;
  }
  const response = await apiFetch('/expense', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.error || 'Failed to add expense');
  }
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
  const normalizedAmount = form.value.reportGain ? Math.abs(amount) : -Math.abs(amount);
  const payload = {
    name: form.value.name,
    category: form.value.category,
    amount: normalizedAmount,
    date: getISODateWithLocalTime(form.value.date),
    tags: form.value.tags,
  };
  try {
    await postExpense(payload);
    setFormMessage('Expense added successfully!', 'success');
    resetForm();
    await refreshExpenses();
  } catch (error) {
    console.error('Error adding expense', error);
    setFormMessage(error.message || 'Failed to add expense', 'error');
  }
}

function toggleCategory(category) {
  const next = new Set(disabledCategories.value);
  if (next.has(category)) {
    next.delete(category);
  } else {
    next.add(category);
  }
  disabledCategories.value = next;
}

function updateChart() {
  if (!chartCanvas.value) return;
  if (!hasExpenseData.value) {
    if (chartInstance) {
      chartInstance.destroy();
      chartInstance = null;
    }
    return;
  }
  const breakdown = calculateCategoryBreakdown(displayedExpenses.value);
  if (chartInstance) {
    chartInstance.destroy();
  }
  chartInstance = new Chart(chartCanvas.value, {
    type: 'doughnut',
    data: {
      labels: breakdown.map((item) => item.category),
      datasets: [
        {
          data: breakdown.map((item) => item.total),
          backgroundColor: breakdown.map((item) => categoryColors.value[item.category]),
          borderColor: '#1a1a1a',
          borderWidth: 1,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { display: false },
        tooltip: {
          callbacks: {
            label(context) {
              const value = context.raw;
              const total = context.dataset.data.reduce((sum, val) => sum + val, 0);
              const percentage = total > 0 ? ((value / total) * 100).toFixed(1) : 0;
              return `${context.label}: ${formatCurrency(value)} (${percentage}%)`;
            },
          },
        },
      },
    },
  });
}

function handleChartClick(event) {
  if (!hasExpenseData.value || !chartInstance) return;
  const query = buildTableNavigationQuery();
  const elements = chartInstance.getElementsAtEventForMode(
    event.nativeEvent ?? event,
    'nearest',
    { intersect: true },
    true
  );
  if (Array.isArray(elements) && elements.length > 0) {
    const { index } = elements[0];
    const category = chartInstance.data.labels?.[index];
    if (typeof category === 'string' && category.trim()) {
      query.category = category.trim();
    }
  }
  router.push({ name: 'table', query }).catch(() => {});
}

function buildTableNavigationQuery() {
  const query = { filter: dateFilter.value };
  if (dateFilter.value === 'range') {
    if (rangeStart.value) query.start = rangeStart.value;
    if (rangeEnd.value) query.end = rangeEnd.value;
  } else if (dateFilter.value === 'month') {
    query.cursor = formatDateForInput(monthCursor.value);
  } else {
    query.anchor = formatDateForInput(currentDate.value);
  }
  return query;
}
</script>
