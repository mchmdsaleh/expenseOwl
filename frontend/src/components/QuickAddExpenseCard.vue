<template>
  <div :class="cardClass" ref="containerRef">
    <form class="space-y-4" @submit.prevent="handleSubmit">
      <div class="space-y-2">
        <label class="text-sm font-medium text-[var(--text-secondary)]" :for="inputId">
          Quick add (describe expenses in plain text)
        </label>
        <textarea
          :id="inputId"
          v-model="typingInput"
          rows="3"
          :class="[inputClass, 'min-h-[96px]']"
          :placeholder="placeholder"
        ></textarea>
        <p class="text-xs text-[var(--text-secondary)]">
          Tulis satu atau beberapa transaksi seperti percakapan biasa. Contoh:
          <span class="italic">"kopi 12k di Kopi Kenangan t:ngantor"; "Bayar PLN 250 ribu tanggal 2024-05-02"</span>. Jika ada beberapa transaksi, pisahkan dengan titik koma (<code>;</code>). Tag opsional gunakan <code>t: tag1, tag2</code>.
        </p>
      </div>
      <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <button
          type="submit"
          :class="[primaryButtonClass, 'w-full sm:w-auto']"
          :disabled="isSubmitting"
        >
          {{ isSubmitting ? 'Processing…' : 'Save Expense' }}
        </button>
        <button
          type="button"
          :class="[primaryButtonClass, 'w-full sm:w-auto']"
          @click="$emit('switch-manual')"
        >
          Switch to Manual
        </button>
      </div>
    </form>
    <div
      v-if="typingMessage.text"
      :class="[
        'mt-4 rounded-full px-4 py-2 text-center text-sm font-medium',
        typingMessage.type === 'success'
          ? 'bg-emerald-500/20 text-emerald-200'
          : 'bg-rose-500/20 text-rose-200'
      ]"
    >
      {{ typingMessage.text }}
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { apiFetch } from '../lib/api';
import { encryptPayload } from '../lib/encryption';

const props = defineProps({
  cardClass: {
    type: String,
    default:
      'rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)]/80 p-6 shadow-card backdrop-blur',
  },
  inputClass: {
    type: String,
    default:
      'w-full rounded-xl border border-[var(--border)] bg-[var(--bg-primary)] px-4 py-2 text-[var(--text-primary)] placeholder:text-[var(--text-secondary)] focus:border-[var(--accent)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40',
  },
  primaryButtonClass: {
    type: String,
    default:
      'inline-flex items-center justify-center gap-2 rounded-full border border-[var(--border)] bg-[var(--bg-secondary)] px-5 py-2 text-sm font-medium text-[var(--text-primary)] transition duration-150 ease-out hover:bg-[var(--accent)] hover:text-white hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40 focus:ring-offset-2 focus:ring-offset-[var(--bg-primary)] disabled:cursor-not-allowed disabled:opacity-50',
  },
  placeholder: {
    type: String,
    default: 'kopi 12k di Kopi Kenangan t:ngantor; bayar PLN 250rb 2 Mei 2024',
  },
  inputId: {
    type: String,
    default: 'typingInput',
  },
});

const emits = defineEmits(['added', 'switch-manual']);

const typingInput = ref('');
const typingMessage = ref({ text: '', type: '' });
const containerRef = ref(null);
const isSubmitting = ref(false);

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

function setTypingMessage(text, type) {
  typingMessage.value = { text, type };
  if (text) {
    setTimeout(() => {
      typingMessage.value = { text: '', type: '' };
    }, 3000);
  }
}

async function handleSubmit() {
  if (isSubmitting.value) return;
  if (!typingInput.value.trim()) {
    setTypingMessage('Enter expense details using the format above.', 'error');
    return;
  }
  isSubmitting.value = true;
  try {
    const defaultDate = new Date().toISOString();
    const response = await apiFetch('/ai/parse-expense', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text: typingInput.value, defaultDate }),
    });
    if (response.status === 503) {
      setTypingMessage('AI expense parser is not configured. Please add expenses manually.', 'error');
      return;
    }
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to parse expenses');
    }
    const payload = await response.json().catch(() => ({}));
    const entriesRaw = Array.isArray(payload?.expenses)
      ? payload.expenses
      : Array.isArray(payload)
        ? payload
        : payload?.expenses != null
          ? [payload.expenses]
          : [];
    if (!entriesRaw.length) {
      setTypingMessage('No transactions detected. Please refine your message.', 'error');
      return;
    }
    let successCount = 0;
    const errors = [];
    for (const entry of entriesRaw) {
      try {
        const normalized = normalizeParsedEntry(entry, defaultDate);
        await postExpense(normalized);
        successCount += 1;
      } catch (err) {
        errors.push(err instanceof Error ? err.message : String(err));
      }
    }
    if (successCount > 0) {
      setTypingMessage(`Added ${successCount} transaction${successCount > 1 ? 's' : ''}.`, 'success');
      typingInput.value = '';
      emits('added');
    } else if (errors.length) {
      setTypingMessage(errors[0], 'error');
    } else {
      setTypingMessage('Unable to add transactions from the provided text.', 'error');
    }
  } catch (error) {
    console.error('Failed quick add expense', error);
    setTypingMessage(error.message || 'Failed to add expense', 'error');
  } finally {
    isSubmitting.value = false;
  }
}

defineExpose({
  reset() {
    typingInput.value = '';
    typingMessage.value = { text: '', type: '' };
    isSubmitting.value = false;
  },
  getContainer() {
    return containerRef.value;
  },
});

const CATEGORY_MAP = {
  food_drinks: 'Food',
  transport: 'Travel',
  fuel: 'Fuel',
  shopping: 'Shopping',
  bills_utilities: 'Utilities',
  entertainment: 'Entertainment',
  health_fitness: 'Healthcare',
  groceries: 'Groceries',
  personal_care: 'Personal Care',
  software_subscription: 'Software',
  misc: 'Miscellaneous',
};

function normalizeParsedEntry(entry, fallbackDate) {
  if (!entry || typeof entry !== 'object') {
    throw new Error('Received invalid expense from parser');
  }
  const amount = Number(entry.amount);
  if (!Number.isFinite(amount) || amount === 0) {
    throw new Error('Parsed transaction has an invalid amount');
  }
  const rawDate = typeof entry.date === 'string' && entry.date ? entry.date : fallbackDate;
  let parsedDate = new Date(rawDate);
  if (Number.isNaN(parsedDate.getTime())) {
    parsedDate = new Date(fallbackDate);
  }
  const isoDate = parsedDate.toISOString();
  const tags = Array.isArray(entry.tags) ? entry.tags.filter((tag) => tag && typeof tag === 'string') : [];
  const uniqueTags = [...new Set(tags.map((tag) => tag.trim()).filter(Boolean))];
  const category = mapCategoryName(entry.category);
  const payload = {
    name: entry.name || 'Quick add expense',
    category,
    amount: amount > 0 ? -Math.abs(amount) : amount,
    date: isoDate,
  };
  if (uniqueTags.length) {
    payload.tags = uniqueTags;
  }
  return payload;
}

function mapCategoryName(raw) {
  const key = String(raw || '').toLowerCase();
  if (CATEGORY_MAP[key]) {
    return CATEGORY_MAP[key];
  }
  const beautified = key
    .split(/[^a-z0-9]+/i)
    .filter(Boolean)
    .map((segment) => segment.charAt(0).toUpperCase() + segment.slice(1))
    .join(' ');
  return beautified || 'Miscellaneous';
}
</script>
