<template>
  <div :class="cardClass" ref="containerRef">
    <form class="space-y-4" @submit.prevent="handleSubmit">
      <div class="space-y-2">
        <label class="text-sm font-medium text-[var(--text-secondary)]" :for="inputId">
          Quick add (Name | Amount | Category | Date | Tags)
        </label>
        <textarea
          :id="inputId"
          v-model="typingInput"
          rows="3"
          :class="[inputClass, 'min-h-[96px]']"
          :placeholder="placeholder"
        ></textarea>
        <p class="text-xs text-[var(--text-secondary)]">
          Use <code>Name | Amount | Category</code> with optional <code>| Date</code> (<code>YYYY-MM-DD</code>) and <code>| Tags</code>.
          Prefix amount with <code>+</code> to record income.
        </p>
      </div>
      <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <button type="submit" :class="[primaryButtonClass, 'w-full sm:w-auto']">
          Save Expense
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
import { computed, ref } from 'vue';
import { apiFetch } from '../lib/api';
import { encryptPayload } from '../lib/encryption';
import { getISODateWithLocalTime } from '../lib/utils';

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
    default: 'Coffee | 4.50 | Food | 2024-05-15 | morning,work',
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

function normalizeDate(value) {
  if (!value) {
    const today = new Date();
    const year = today.getFullYear();
    const month = String(today.getMonth() + 1).padStart(2, '0');
    const day = String(today.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  }
  return value;
}

async function handleSubmit() {
  if (!typingInput.value.trim()) {
    setTypingMessage('Enter expense details using the format above.', 'error');
    return;
  }
  const segments = typingInput.value
    .split('|')
    .map((segment) => segment.trim())
    .filter(Boolean);
  if (segments.length < 3) {
    setTypingMessage('Use at least “Name | Amount | Category”.', 'error');
    return;
  }
  const [nameSegment, amountSegment, categorySegment, dateSegment = '', tagsSegment = ''] = segments;
  if (!nameSegment) {
    setTypingMessage('Name is required.', 'error');
    return;
  }
  if (!categorySegment) {
    setTypingMessage('Category is required.', 'error');
    return;
  }
  const numericAmount = Number(amountSegment.replace(/[^0-9.+-]/g, ''));
  if (!Number.isFinite(numericAmount) || numericAmount === 0) {
    setTypingMessage('Amount must be a non-zero number.', 'error');
    return;
  }
  const isGain = /^\s*\+/.test(amountSegment);
  const normalizedAmount = isGain ? Math.abs(numericAmount) : -Math.abs(numericAmount);
  const dateInput = normalizeDate(dateSegment);
  if (!/^\d{4}-\d{2}-\d{2}$/.test(dateInput)) {
    setTypingMessage('Date must use YYYY-MM-DD format.', 'error');
    return;
  }
  const tags = tagsSegment
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean);

  const payload = {
    name: nameSegment,
    category: categorySegment,
    amount: normalizedAmount,
    date: getISODateWithLocalTime(dateInput),
    tags,
  };

  try {
    await postExpense(payload);
    setTypingMessage('Expense added successfully!', 'success');
    typingInput.value = '';
    emits('added');
  } catch (error) {
    console.error('Failed quick add expense', error);
    setTypingMessage(error.message || 'Failed to add expense', 'error');
  }
}

defineExpose({
  reset() {
    typingInput.value = '';
    typingMessage.value = { text: '', type: '' };
  },
  getContainer() {
    return containerRef.value;
  },
});
</script>
