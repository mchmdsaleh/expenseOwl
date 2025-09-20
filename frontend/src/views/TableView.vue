<template>
  <section class="space-y-6">
    <div class="flex items-center justify-center gap-4" v-show="!showAll">
      <button :class="iconButtonClass" @click="gotoPrevMonth"><i class="fa-solid fa-arrow-left"></i></button>
      <div class="min-w-[200px] text-center text-2xl font-bold">{{ monthLabel }}</div>
      <button :class="iconButtonClass" @click="gotoNextMonth"><i class="fa-solid fa-arrow-right"></i></button>
    </div>

    <div class="flex items-center justify-end gap-3 text-sm text-[var(--text-secondary)]">
      <label for="showAllToggle" class="flex items-center gap-2">
        <input id="showAllToggle" v-model="showAll" type="checkbox" :class="checkboxClass" />
        Show All Transactions
      </label>
    </div>

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
            v-model.number="form.amount"
            type="number"
            step="0.01"
            min="0.01"
            max="9000000000000000"
            :class="inputClass"
            required
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

    <div>
      <div
        v-if="tableExpenses.length === 0"
        class="w-full rounded-3xl border border-dashed border-[var(--border)] bg-[var(--bg-secondary)]/60 py-12 text-center text-base italic text-[var(--text-secondary)]"
      >
        {{ showAll ? 'No transactions found' : 'No expenses recorded for this month' }}
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
import { ref, computed, watch, onMounted } from 'vue';
import TagInput from '../components/TagInput.vue';
import state, { loadInitialData, refreshExpenses } from '../stores/appState';
import { apiFetch } from '../lib/api';
import { formatMonth, getMonthExpenses, getISODateWithLocalTime, formatDateFromUTC, formatCurrency as formatCurrencyRaw } from '../lib/utils';

const currentDate = ref(new Date());
const showAll = ref(false);
const form = ref(createDefaultForm());
const editId = ref(null);
const formMessage = ref({ text: '', type: '' });
const showDeleteModal = ref(false);
const expenseToDelete = ref(null);

const monthLabel = computed(() => formatMonth(currentDate.value));

const tableExpenses = computed(() => {
  const base = showAll.value
    ? [...state.expenses].sort((a, b) => new Date(b.date) - new Date(a.date))
    : getMonthExpenses(state.expenses, currentDate.value, state.startDate);
  return base;
});

const hasTags = computed(() => tableExpenses.value.some((expense) => Array.isArray(expense.tags) && expense.tags.length > 0));

watch(
  () => state.expenses,
  () => {
    if (!showAll.value) {
      ensureCurrentMonthAvailable();
    }
  }
);

onMounted(async () => {
  await loadInitialData();
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

const iconGhostButton =
  'inline-flex h-9 w-9 items-center justify-center rounded-full border border-transparent text-[var(--text-secondary)] transition duration-150 hover:border-[var(--border)] hover:bg-[var(--bg-primary)]/60 hover:text-[var(--text-primary)]';

const iconDangerButton =
  'inline-flex h-9 w-9 items-center justify-center rounded-full border border-transparent text-[var(--text-secondary)] transition duration-150 hover:border-rose-500/40 hover:bg-rose-500/10 hover:text-rose-400';

function ensureCurrentMonthAvailable() {
  // noop placeholder for future adjustments
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
  editId.value = null;
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
  window.scrollTo({ top: 0, behavior: 'smooth' });
}

function toLocalDate(isoDate) {
  const date = new Date(isoDate);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
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
  const payload = {
    name: form.value.name,
    category: form.value.category,
    amount,
    date: getISODateWithLocalTime(form.value.date),
    tags: form.value.tags,
  };
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
    resetForm();
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
