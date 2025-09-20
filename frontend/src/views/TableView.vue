<template>
  <section class="space-y-6">
    <div class="month-navigation flex items-center justify-center gap-4" v-show="!showAll">
      <button class="nav-button" @click="gotoPrevMonth"><i class="fa-solid fa-arrow-left"></i></button>
      <div class="current-month">{{ monthLabel }}</div>
      <button class="nav-button" @click="gotoNextMonth"><i class="fa-solid fa-arrow-right"></i></button>
    </div>

    <div class="table-controls flex items-center justify-end gap-2">
      <label for="showAllToggle">
        <input id="showAllToggle" v-model="showAll" type="checkbox" class="styled-checkbox" /> Show All Transactions
      </label>
    </div>

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
            <option v-for="category in state.categories" :key="category" :value="category">
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

    <div id="tableContainer">
      <div v-if="tableExpenses.length === 0" class="no-data">
        {{ showAll ? 'No transactions found' : 'No expenses recorded for this month' }}
      </div>
      <table v-else class="expense-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Category</th>
            <th v-if="hasTags" class="tags-column">Tags</th>
            <th>Amount</th>
            <th class="date-header">Date</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(expense, index) in tableExpenses" :key="expense.id">
            <td>{{ expense.name }}</td>
            <td>{{ expense.category }}</td>
            <td v-if="hasTags" class="tags-column">{{ (expense.tags || []).join(', ') }}</td>
            <td class="amount">{{ formatCurrency(expense.amount) }}</td>
            <td class="date-column">{{ formatDate(expense.date) }}</td>
            <td>
              <button class="edit-button" type="button" @click="editExpense(expense)">
                <i class="fa-solid fa-pen-to-square"></i>
              </button>
              <button class="delete-button" type="button" @click="openDeleteModal(expense, $event)">
                <i class="fa-solid fa-trash-can"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="modal" :class="{ active: showDeleteModal }" @click.self="closeDeleteModal">
      <div class="modal-content">
        <h3>Delete Expense</h3>
        <p>Are you sure you want to delete this expense? (cannot be undone)</p>
        <div class="modal-buttons">
          <button class="modal-button" @click="closeDeleteModal">Cancel</button>
          <button class="modal-button confirm" @click="confirmDelete">Delete</button>
        </div>
      </div>
    </div>
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
