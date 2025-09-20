<template>
  <section class="space-y-6">
    <div class="info-field">
      Version:
      <a href="https://github.com/tanq16/expenseowl/releases/latest" target="_blank" rel="noopener noreferrer">v4.0</a>
      <span class="separator">|</span>
      <a href="https://github.com/tanq16/expenseowl/blob/main/README.md" target="_blank" rel="noopener noreferrer">Documentation</a>
      <span class="separator">|</span>
      <a href="https://github.com/tanq16/expenseowl" target="_blank" rel="noopener noreferrer">GitHub</a>
    </div>

    <div class="form-container bg-surface shadow-card">
      <h2 align="center">Category Settings</h2>
      <div class="categories-list">
        <div v-for="(category, index) in categories" :key="category" class="category-item">
          <div class="category-handle-area">
            <span class="drag-handle"><i class="fa-solid fa-grip-lines"></i></span>
            <span>{{ category }}</span>
          </div>
          <div class="category-actions">
            <button type="button" class="nav-button" @click="moveCategory(index, -1)" :disabled="index === 0">
              <i class="fa-solid fa-arrow-up"></i>
            </button>
            <button type="button" class="nav-button" @click="moveCategory(index, 1)" :disabled="index === categories.length - 1">
              <i class="fa-solid fa-arrow-down"></i>
            </button>
            <button type="button" class="delete-button" @click="removeCategory(index)">
              <i class="fa-solid fa-times"></i>
            </button>
          </div>
        </div>
        <div class="category-input-container">
          <input v-model="newCategory" type="text" placeholder="Add new category" />
          <button type="button" class="nav-button" @click="addCategory">
            Add
          </button>
        </div>
        <button type="button" class="nav-button" @click="saveCategories">Save Categories</button>
        <div class="form-message" :class="categoryMessage.type" v-if="categoryMessage.text">{{ categoryMessage.text }}</div>
      </div>
    </div>

    <div class="settings-container">
      <div class="form-container half-width bg-surface shadow-card">
        <h2 align="center">Currency Settings</h2>
        <div class="currency-selector">
          <select v-model="currencyCode">
            <option v-for="code in currencyOptions" :key="code" :value="code">
              {{ code.toUpperCase() }} ({{ currencyBehaviors[code].symbol }})
            </option>
          </select>
          <button type="button" class="nav-button" @click="saveCurrency">Save</button>
        </div>
        <div class="form-message" :class="currencyMessage.type" v-if="currencyMessage.text">{{ currencyMessage.text }}</div>
      </div>
      <div class="form-container half-width bg-surface shadow-card">
        <h2 align="center">Start Date Settings</h2>
        <div class="start-date-manager">
          <input v-model.number="startDate" type="number" min="1" max="31" />
          <button type="button" class="nav-button" @click="saveStartDate">Save</button>
        </div>
        <div class="form-message" :class="startDateMessage.type" v-if="startDateMessage.text">{{ startDateMessage.text }}</div>
      </div>
    </div>

    <div class="settings-container">
      <div class="form-container half-width bg-surface shadow-card">
        <h2 align="center">Theme Settings</h2>
        <div class="theme-selector">
          <select v-model="theme" @change="applyTheme">
            <option value="system">System Default</option>
            <option value="light">Light</option>
            <option value="dark">Dark</option>
          </select>
        </div>
        <div class="form-message" :class="themeMessage.type" v-if="themeMessage.text">{{ themeMessage.text }}</div>
      </div>
      <div class="form-container half-width bg-surface shadow-card">
        <h2 align="center">Import/Export Data</h2>
        <div class="export-buttons">
          <div class="export-options">
            <a href="/export/csv" class="nav-button" download="expenses.csv">Export to CSV</a>
          </div>
          <div class="import-option">
            <label class="nav-button" for="csv-import-file">Import from CSV</label>
            <input id="csv-import-file" ref="csvImportRef" type="file" accept=".csv" hidden @change="(event) => handleImport(event, '/import/csv')" />
          </div>
          <div class="import-option">
            <label class="nav-button" for="csv-import-file-old">Import from ExpenseOwl v3.20-</label>
            <input id="csv-import-file-old" ref="csvImportOldRef" type="file" accept=".csv" hidden @change="(event) => handleImport(event, '/import/csvold')" />
          </div>
        </div>
        <div class="form-message" :class="importMessage.type" v-if="importMessage.text">{{ importMessage.text }}</div>
        <div class="import-summary" v-if="importSummary" style="display: block;">
          <h3>Import Summary</h3>
          <p>Total Processed: <span>{{ importSummary.totalProcessed }}</span></p>
          <p>Imported: <span>{{ importSummary.imported }}</span></p>
          <p>Skipped: <span>{{ importSummary.skipped }}</span></p>
          <p>New Categories: <span>{{ importSummary.newCategories }}</span></p>
        </div>
      </div>
    </div>

    <div class="form-container bg-surface shadow-card">
      <h2 align="center">Recurring Transactions</h2>
      <form class="expense-form recurring-expense-form" @submit.prevent="submitRecurring">
        <div class="form-group">
          <label for="recurringName">Name</label>
          <input id="recurringName" v-model="recurringForm.name" type="text" required />
        </div>
        <div class="form-group">
          <label for="recurringAmount">Amount</label>
          <input id="recurringAmount" v-model.number="recurringForm.amount" type="number" step="0.01" min="0.01" max="9000000000000000" required />
        </div>
        <div class="form-group">
          <label for="recurringCategory">Category</label>
          <select id="recurringCategory" v-model="recurringForm.category" required>
            <option value="" disabled>Select category</option>
            <option v-for="category in state.categories" :key="category" :value="category">{{ category }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>Tags</label>
          <TagInput v-model="recurringForm.tags" :suggestions="allTags" />
        </div>
        <div class="form-group">
          <label for="recurringInterval">Interval</label>
          <select id="recurringInterval" v-model="recurringForm.interval" required>
            <option value="daily">Daily</option>
            <option value="weekly">Weekly</option>
            <option value="monthly">Monthly</option>
            <option value="yearly">Yearly</option>
          </select>
        </div>
        <div class="form-group">
          <label for="recurringStartDate">Start Date</label>
          <input id="recurringStartDate" v-model="recurringForm.startDate" type="date" required />
        </div>
        <div class="form-group">
          <label for="recurringOccurrences">Occurrences (0 for indefinite)</label>
          <input id="recurringOccurrences" v-model.number="recurringForm.occurrences" type="number" min="0" required />
        </div>
        <div class="form-group form-group-checkbox">
          <label for="recurringReportGain">Report Gain</label>
          <input id="recurringReportGain" v-model="recurringForm.reportGain" type="checkbox" class="styled-checkbox" />
        </div>
        <button type="submit" class="nav-button">{{ recurringForm.submitLabel }}</button>
      </form>
      <div class="form-message" :class="recurringMessage.type" v-if="recurringMessage.text">{{ recurringMessage.text }}</div>

      <h3 align="center" style="margin-top: 2rem;">Existing Recurring Transactions</h3>
      <div v-if="state.recurringExpenses.length === 0" class="no-data">No recurring transactions configured.</div>
      <div v-else id="recurring-expenses-list">
        <div v-for="expense in state.recurringExpenses" :key="expense.id" class="recurring-item">
          <div>
            <strong>{{ expense.name }}</strong>
            <div class="recurring-meta">
              <span>{{ formatCurrency(expense.amount) }}</span>
              <span>• {{ expense.interval }}</span>
              <span>• Starts {{ formatDate(expense.start_date) }}</span>
              <span v-if="expense.occurrences && expense.occurrences > 0">• {{ expense.occurrences }} occurrences</span>
            </div>
            <div v-if="expense.tags && expense.tags.length" class="recurring-tags">Tags: {{ expense.tags.join(', ') }}</div>
          </div>
          <div class="recurring-actions">
            <button type="button" class="nav-button" @click="editRecurring(expense)">
              <i class="fa-solid fa-pen-to-square"></i>
            </button>
            <button type="button" class="delete-button" @click="deleteRecurring(expense)">
              <i class="fa-solid fa-trash-can"></i>
            </button>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue';
import TagInput from '../components/TagInput.vue';
import state, { loadInitialData, refreshExpenses, refreshRecurringExpenses } from '../stores/appState';
import { apiFetch } from '../lib/api';
import { currencyBehaviors, formatCurrency as formatCurrencyRaw } from '../lib/utils';

const currencyOptions = Object.keys(currencyBehaviors);

const categories = ref([]);
const newCategory = ref('');
const categoryMessage = ref({ text: '', type: '' });

const currencyCode = ref(state.currency);
const currencyMessage = ref({ text: '', type: '' });

const startDate = ref(state.startDate);
const startDateMessage = ref({ text: '', type: '' });

const theme = ref(localStorage.getItem('theme') || 'system');
const themeMessage = ref({ text: '', type: '' });

const importMessage = ref({ text: '', type: '' });
const importSummary = ref(null);
const csvImportRef = ref(null);
const csvImportOldRef = ref(null);

const recurringForm = ref(createRecurringForm());
const recurringMessage = ref({ text: '', type: '' });
const editingRecurringId = ref(null);

const allTags = computed(() => {
  const combined = new Set([...(state.tags || [])]);
  state.recurringExpenses.forEach((expense) => {
    (expense.tags || []).forEach((tag) => combined.add(tag));
  });
  return Array.from(combined);
});

watch(
  () => state.categories,
  (value) => {
    categories.value = [...value];
    if (!currencyOptions.includes(currencyCode.value)) {
      currencyCode.value = state.currency;
    }
  },
  { immediate: true }
);

watch(
  () => state.currency,
  (value) => {
    currencyCode.value = value;
  },
  { immediate: true }
);

watch(
  () => state.startDate,
  (value) => {
    startDate.value = value;
  },
  { immediate: true }
);

onMounted(async () => {
  await loadInitialData();
});

function createRecurringForm() {
  const today = new Date();
  const year = today.getFullYear();
  const month = String(today.getMonth() + 1).padStart(2, '0');
  const day = String(today.getDate()).padStart(2, '0');
  return {
    name: '',
    amount: null,
    category: '',
    tags: [],
    interval: 'monthly',
    startDate: `${year}-${month}-${day}`,
    occurrences: 2,
    reportGain: false,
    submitLabel: 'Add Recurring Transaction',
  };
}

function sanitizeCategory(value) {
  return value.replace(/[<>]/g, ' ').trim();
}

function addCategory() {
  const candidate = sanitizeCategory(newCategory.value);
  if (!candidate) {
    setCategoryMessage('Category name cannot be empty.', 'error');
    return;
  }
  if (categories.value.includes(candidate)) {
    setCategoryMessage('Category already exists.', 'error');
    return;
  }
  categories.value.push(candidate);
  newCategory.value = '';
}

function removeCategory(index) {
  categories.value.splice(index, 1);
}

function moveCategory(index, delta) {
  const newIndex = index + delta;
  if (newIndex < 0 || newIndex >= categories.value.length) return;
  const updated = [...categories.value];
  const [item] = updated.splice(index, 1);
  updated.splice(newIndex, 0, item);
  categories.value = updated;
}

function setCategoryMessage(text, type) {
  categoryMessage.value = { text, type };
  dismissAfter(() => (categoryMessage.value = { text: '', type: '' }));
}

async function saveCategories() {
  if (categories.value.length === 0) {
    setCategoryMessage('At least one category is required.', 'error');
    return;
  }
  try {
    const response = await apiFetch('/categories/edit', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(categories.value),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to save categories');
    }
    setCategoryMessage('Categories saved successfully.', 'success');
    state.categories = [...categories.value];
  } catch (error) {
    console.error('Failed to save categories', error);
    setCategoryMessage(error.message || 'Failed to save categories', 'error');
  }
}

function setCurrencyMessage(text, type) {
  currencyMessage.value = { text, type };
  dismissAfter(() => (currencyMessage.value = { text: '', type: '' }));
}

async function saveCurrency() {
  try {
    const response = await apiFetch('/currency/edit', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(currencyCode.value),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to save currency');
    }
    state.currency = currencyCode.value;
    setCurrencyMessage('Currency saved successfully.', 'success');
  } catch (error) {
    console.error('Failed to save currency', error);
    setCurrencyMessage(error.message || 'Failed to save currency', 'error');
  }
}

function setStartDateMessage(text, type) {
  startDateMessage.value = { text, type };
  dismissAfter(() => (startDateMessage.value = { text: '', type: '' }));
}

async function saveStartDate() {
  if (!Number.isInteger(startDate.value) || startDate.value < 1 || startDate.value > 31) {
    setStartDateMessage('Start date must be between 1 and 31.', 'error');
    return;
  }
  try {
    const response = await apiFetch('/startdate/edit', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(startDate.value),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to save start date');
    }
    state.startDate = startDate.value;
    setStartDateMessage('Start date saved successfully.', 'success');
  } catch (error) {
    console.error('Failed to save start date', error);
    setStartDateMessage(error.message || 'Failed to save start date', 'error');
  }
}

function applyTheme() {
  const value = theme.value;
  if (value === 'system') {
    document.documentElement.removeAttribute('data-theme');
  } else {
    document.documentElement.setAttribute('data-theme', value);
  }
  localStorage.setItem('theme', value);
  state.theme = value;
  themeMessage.value = { text: 'Theme updated.', type: 'success' };
  dismissAfter(() => (themeMessage.value = { text: '', type: '' }));
}

function setImportMessage(text, type) {
  importMessage.value = { text, type };
  if (text) {
    dismissAfter(() => (importMessage.value = { text: '', type: '' }));
  }
}

async function handleImport(event, endpoint) {
  const input = event.target;
  const file = input.files?.[0];
  if (!file) {
    return;
  }
  const formData = new FormData();
  formData.append('file', file);
  setImportMessage('Importing... this may take a while for large files.', '');
  importSummary.value = null;

  try {
    const response = await apiFetch(endpoint, {
      method: 'POST',
      body: formData,
    });
    const result = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(result.error || 'Failed to import CSV');
    }
    importSummary.value = {
      totalProcessed: result.total_processed ?? 0,
      imported: result.imported ?? 0,
      skipped: result.skipped ?? 0,
      newCategories: (result.new_categories || []).join(', ') || 'None',
    };
    setImportMessage('Import completed!', 'success');
    await refreshExpenses();
    await refreshRecurringExpenses();
  } catch (error) {
    console.error('Failed to import CSV', error);
    importSummary.value = null;
    setImportMessage(error.message || 'Failed to import CSV', 'error');
  } finally {
    input.value = '';
  }
}

function setRecurringMessage(text, type) {
  recurringMessage.value = { text, type };
  dismissAfter(() => (recurringMessage.value = { text: '', type: '' }));
}

function resetRecurringForm() {
  recurringForm.value = createRecurringForm();
  editingRecurringId.value = null;
}

async function submitRecurring() {
  if (!recurringForm.value.category) {
    setRecurringMessage('Please select a category.', 'error');
    return;
  }
  let amount = Number(recurringForm.value.amount || 0);
  if (Number.isNaN(amount) || amount === 0) {
    setRecurringMessage('Please enter a valid amount.', 'error');
    return;
  }
  if (!recurringForm.value.reportGain) {
    amount *= -1;
  }
  const payload = {
    name: recurringForm.value.name,
    amount,
    category: recurringForm.value.category,
    tags: recurringForm.value.tags,
    interval: recurringForm.value.interval,
    start_date: recurringForm.value.startDate,
    occurrences: recurringForm.value.occurrences,
  };
  const isEdit = Boolean(editingRecurringId.value);
  try {
    let response;
    if (isEdit) {
      const params = new URLSearchParams({ updateAll: 'true' });
      response = await apiFetch(`/recurring-expense/edit?id=${editingRecurringId.value}&${params.toString()}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
    } else {
      response = await apiFetch('/recurring-expense', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
    }
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to save recurring expense');
    }
    await refreshRecurringExpenses();
    setRecurringMessage(isEdit ? 'Recurring expense updated.' : 'Recurring expense added.', 'success');
    resetRecurringForm();
  } catch (error) {
    console.error('Failed to save recurring expense', error);
    setRecurringMessage(error.message || 'Failed to save recurring expense', 'error');
  }
}

function editRecurring(expense) {
  editingRecurringId.value = expense.id;
  recurringForm.value = {
    name: expense.name,
    amount: Math.abs(expense.amount),
    category: expense.category,
    tags: [...(expense.tags || [])],
    interval: expense.interval,
    startDate: expense.start_date?.slice(0, 10) || createRecurringForm().startDate,
    occurrences: expense.occurrences ?? 0,
    reportGain: expense.amount > 0,
    submitLabel: 'Update Recurring Transaction',
  };
  window.scrollTo({ top: 0, behavior: 'smooth' });
}

async function deleteRecurring(expense) {
  if (!confirm('Delete this recurring transaction (including future occurrences)?')) {
    return;
  }
  try {
    const params = new URLSearchParams({ removeAll: 'true' });
    const response = await apiFetch(`/recurring-expense/delete?id=${expense.id}&${params.toString()}`, {
      method: 'DELETE',
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to delete recurring expense');
    }
    await refreshRecurringExpenses();
    setRecurringMessage('Recurring expense deleted.', 'success');
  } catch (error) {
    console.error('Failed to delete recurring expense', error);
    setRecurringMessage(error.message || 'Failed to delete recurring expense', 'error');
  }
}

function formatCurrency(amount) {
  return formatCurrencyRaw(amount, state.currency);
}

function formatDate(date) {
  if (!date) return 'Unknown';
  const local = new Date(date);
  return local.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

function dismissAfter(callback) {
  setTimeout(() => {
    callback();
  }, 3000);
}
</script>
