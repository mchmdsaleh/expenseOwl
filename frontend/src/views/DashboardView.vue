<template>
  <section class="space-y-6">
    <div class="flex items-center justify-center gap-4">
      <button :class="iconButtonClass" @click="gotoPrevMonth"><i class="fa-solid fa-arrow-left"></i></button>
      <div class="min-w-[200px] text-center text-2xl font-bold">{{ monthLabel }}</div>
      <button :class="iconButtonClass" @click="gotoNextMonth"><i class="fa-solid fa-arrow-right"></i></button>
    </div>

    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div
        v-if="userDisplayName"
        class="inline-flex w-full items-center justify-center gap-2 rounded-2xl border border-[var(--border)] bg-[var(--bg-secondary)]/70 px-4 py-2 text-sm font-medium text-[var(--text-primary)] md:w-auto md:justify-start"
      >
        <i class="fa-solid fa-circle-user text-[var(--accent)]"></i>
        <span>{{ userDisplayName }}</span>
      </div>
      <div class="flex justify-end">
        <button :class="primaryButtonClass" @click="toggleExpenseForm">
          <i :class="showExpenseForm ? 'fa-solid fa-times' : 'fa-solid fa-plus'"></i>
          {{ showExpenseForm ? 'Close' : 'Add Expense' }}
        </button>
      </div>
    </div>

    <div v-if="showExpenseForm" id="addExpenseContainer">
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

    <div class="flex flex-col gap-6 rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)]/80 p-6 shadow-card backdrop-blur lg:flex-row">
      <div v-if="!hasExpenseData" class="w-full rounded-3xl border border-dashed border-[var(--border)] bg-[var(--bg-secondary)]/60 py-12 text-center text-base italic text-[var(--text-secondary)]">
        No expenses recorded this month.
      </div>
      <template v-else>
        <div class="flex h-80 flex-1 items-center justify-center">
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
import { Chart, registerables } from 'chart.js';
import state, { loadInitialData, refreshExpenses } from '../stores/appState';
import TagInput from '../components/TagInput.vue';
import { formatMonth, getMonthExpenses, formatCurrency as formatCurrencyRaw, getISODateWithLocalTime, colorPalette } from '../lib/utils';
import { apiFetch } from '../lib/api';

Chart.register(...registerables);
Chart.defaults.color = '#b3b3b3';
Chart.defaults.borderColor = '#606060';
Chart.defaults.font.family = '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif';

const chartCanvas = ref(null);
let chartInstance = null;

const currentDate = ref(new Date());
const showExpenseForm = ref(false);
const disabledCategories = ref(new Set());
const categoryColors = ref({});

const form = ref(createDefaultForm());
const formMessage = ref({ text: '', type: '' });
const rawAmount = ref('');

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

const cardClass =
  'rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)]/80 p-6 shadow-card backdrop-blur';

const legendItemClass =
  'flex items-center gap-4 rounded-2xl border border-transparent px-3 py-2 text-sm transition duration-150 ease-out hover:bg-[var(--bg-primary)]/60';

const cashflowCardClass =
  'flex flex-col items-center justify-center rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)]/80 px-6 py-6 text-center shadow-card backdrop-blur';

const categories = computed(() => state.categories);
const formattedAmount = computed(() => {
  if (!rawAmount.value) return '';
  const numeric = Number(rawAmount.value.replace(/[^0-9.-]/g, '')) || 0;
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(numeric);
});

const monthExpenses = computed(() => getMonthExpenses(state.expenses, currentDate.value, state.startDate));
const hasExpenseData = computed(() => monthExpenses.value.some((expense) => expense.amount < 0));

const income = computed(() => monthExpenses.value.filter((exp) => exp.amount > 0).reduce((sum, exp) => sum + exp.amount, 0));
const totalExpenses = computed(() => monthExpenses.value.filter((exp) => exp.amount < 0).reduce((sum, exp) => sum + Math.abs(exp.amount), 0));
const balance = computed(() => income.value - totalExpenses.value);

const monthLabel = computed(() => formatMonth(currentDate.value));

const legendEntries = computed(() => buildLegendEntries());
const totalActiveExpenses = computed(() => {
  return monthExpenses.value
    .filter((exp) => exp.amount < 0 && !disabledCategories.value.has(exp.category))
    .reduce((sum, exp) => sum + Math.abs(exp.amount), 0);
});
const totalActiveFormatted = computed(() => formatCurrency(totalActiveExpenses.value));

watch(
  () => state.expenses,
  () => {
    assignCategoryColors();
    updateChart();
  }
);

watch(monthExpenses, () => {
  assignCategoryColors();
  updateChart();
});

watch(disabledCategories, updateChart, { deep: true });

onMounted(async () => {
  await loadInitialData();
  assignCategoryColors();
  updateChart();
});

onBeforeUnmount(() => {
  if (chartInstance) {
    chartInstance.destroy();
    chartInstance = null;
  }
});

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

function toggleExpenseForm() {
  showExpenseForm.value = !showExpenseForm.value;
  if (!showExpenseForm.value) {
    resetForm();
  }
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

function assignCategoryColors() {
  const colors = { ...categoryColors.value };
  const allCategories = Array.from(
    new Set([
      ...state.categories,
      ...state.expenses.map((expense) => expense.category).filter(Boolean),
    ]),
  );
  allCategories.forEach((category) => {
    if (!category) return;
    if (!colors[category]) {
      const nextIndex = Object.keys(colors).length;
      colors[category] = colorPalette[nextIndex % colorPalette.length];
    }
  });
  categoryColors.value = colors;
}

function gotoPrevMonth() {
  const date = new Date(currentDate.value);
  date.setMonth(date.getMonth() - 1);
  currentDate.value = date;
  nextTick(() => {
    assignCategoryColors();
    updateChart();
  });
}

function gotoNextMonth() {
  const date = new Date(currentDate.value);
  date.setMonth(date.getMonth() + 1);
  currentDate.value = date;
  nextTick(() => {
    assignCategoryColors();
    updateChart();
  });
}

function calculateCategoryBreakdown(expenses) {
  const categoryTotals = {};
  let totalAmount = 0;
  expenses.forEach((exp) => {
    if (exp.amount < 0 && !disabledCategories.value.has(exp.category)) {
      const amount = Math.abs(exp.amount);
      categoryTotals[exp.category] = (categoryTotals[exp.category] || 0) + amount;
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
  const breakdown = calculateCategoryBreakdown(monthExpenses.value);
  const categoryMap = new Map(breakdown.map((item) => [item.category, item]));
  const currentMonthCategories = Array.from(
    new Set(
      monthExpenses.value
        .filter((exp) => exp.amount < 0)
        .map((exp) => exp.category)
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
  const body = {
    name: form.value.name,
    category: form.value.category,
    amount,
    date: getISODateWithLocalTime(form.value.date),
    tags: form.value.tags,
  };
  try {
    const response = await apiFetch('/expense', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to add expense');
    }
    setFormMessage('Expense added successfully!', 'success');
    resetForm();
    await refreshExpenses();
    showExpenseForm.value = false;
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
  const breakdown = calculateCategoryBreakdown(monthExpenses.value);
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
</script>
