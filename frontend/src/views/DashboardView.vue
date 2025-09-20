<template>
  <section class="space-y-6">
    <div class="month-navigation flex items-center justify-center gap-4">
      <button class="nav-button" @click="gotoPrevMonth"><i class="fa-solid fa-arrow-left"></i></button>
      <div class="current-month">{{ monthLabel }}</div>
      <button class="nav-button" @click="gotoNextMonth"><i class="fa-solid fa-arrow-right"></i></button>
    </div>

    <div class="table-controls flex justify-end">
      <button class="nav-button" @click="toggleExpenseForm">
        <i :class="showExpenseForm ? 'fa-solid fa-times' : 'fa-solid fa-plus'"></i>
        {{ showExpenseForm ? 'Close' : 'Add Expense' }}
      </button>
    </div>

    <div v-if="showExpenseForm" id="addExpenseContainer">
      <div class="form-container bg-surface shadow-card">
        <form class="expense-form" @submit.prevent="submitExpense">
          <div class="form-group">
            <label for="name">Name</label>
            <input id="name" v-model="form.name" type="text" required />
          </div>
          <div class="form-group">
            <label for="category">Category</label>
            <select id="category" v-model="form.category" required>
              <option value="" disabled>Choose category</option>
              <option v-for="category in categories" :key="category" :value="category">
                {{ category }}
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>Tags</label>
            <TagInput v-model="form.tags" :suggestions="state.tags" />
          </div>
          <div class="form-group">
            <label for="amount">Amount</label>
            <input id="amount" v-model.number="form.amount" type="number" step="0.01" min="0.01" max="9000000000000000" required />
          </div>
          <div class="form-group">
            <label for="date">Date</label>
            <input id="date" v-model="form.date" type="date" required />
          </div>
          <div class="form-group form-group-checkbox">
            <label for="reportGain">Report Gain</label>
            <input id="reportGain" v-model="form.reportGain" type="checkbox" class="styled-checkbox" />
          </div>
          <button type="submit" class="nav-button">{{ form.submitLabel }}</button>
        </form>
        <div class="form-message" :class="formMessage.type" v-if="formMessage.text">{{ formMessage.text }}</div>
      </div>
    </div>

    <div class="chart-container">
      <div v-if="!hasExpenseData" class="no-data">No expenses recorded this month.</div>
      <template v-else>
        <div class="chart-box">
          <canvas ref="chartCanvas"></canvas>
        </div>
        <div class="legend-box">
          <div
            v-for="entry in legendEntries"
            :key="entry.category"
            class="legend-item"
            :class="{ disabled: entry.disabled }"
            @click="toggleCategory(entry.category)"
          >
            <div class="color-box" :style="{ backgroundColor: entry.color }"></div>
            <div class="legend-text">
              <span>{{ entry.category }}<template v-if="entry.percentage !== null"> ({{ entry.percentage.toFixed(1) }}%)</template></span>
              <span class="amount" v-if="entry.amount !== null">{{ entry.amountFormatted }}</span>
            </div>
          </div>
          <div class="legend-total">
            <span>Total:</span>
            <span class="amount">{{ totalActiveFormatted }}</span>
          </div>
        </div>
      </template>
    </div>

    <div v-if="hasExpenseData" class="cashflow-container">
      <div class="cashflow-item income">
        <div class="cashflow-label">Income</div>
        <div class="cashflow-value">{{ formatCurrency(income) }}</div>
      </div>
      <div class="cashflow-item expenses">
        <div class="cashflow-label">Expenses</div>
        <div class="cashflow-value">{{ formatCurrency(totalExpenses) }}</div>
      </div>
      <div class="cashflow-item balance">
        <div class="cashflow-label">Balance</div>
        <div class="cashflow-value" :class="balance >= 0 ? 'positive' : 'negative'">{{ formatCurrency(balance) }}</div>
      </div>
    </div>
  </section>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
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

const categories = computed(() => state.categories);

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

watch(monthExpenses, updateChart);
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
}

function gotoNextMonth() {
  const date = new Date(currentDate.value);
  date.setMonth(date.getMonth() + 1);
  currentDate.value = date;
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

async function submitExpense() {
  if (!form.value.category) {
    setFormMessage('Please select a category', 'error');
    return;
  }
  let amount = Number(form.value.amount || 0);
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
