<template>
  <div
    ref="containerRef"
    class="relative flex min-h-[44px] cursor-text flex-wrap items-center gap-2 rounded-2xl border border-[var(--border)] bg-[var(--bg-primary)] px-3 py-2 text-sm text-[var(--text-primary)]"
    @click="focusInput"
  >
    <div class="flex flex-wrap gap-2">
      <div
        v-for="tag in internalValue"
        :key="tag"
        class="flex items-center gap-2 rounded-full bg-[var(--accent)] px-3 py-1 text-xs font-medium text-white"
      >
        {{ tag }}
        <button
          type="button"
          class="transition hover:text-gray-200"
          @click.stop="removeTag(tag)"
        >
          ×
        </button>
      </div>
    </div>
    <input
      ref="inputRef"
      v-model="inputValue"
      type="text"
      :placeholder="placeholder"
      class="min-w-[120px] flex-1 border-none bg-transparent text-[var(--text-primary)] placeholder:text-[var(--text-secondary)] focus:outline-none"
      @focus="openDropdown"
      @input="onInput"
      @keydown.enter.prevent="createTag"
    />
    <div
      v-if="showDropdown"
      class="absolute left-0 top-full z-50 mt-2 w-full overflow-hidden rounded-2xl border border-[var(--border)] bg-[var(--bg-secondary)]/95 shadow-card backdrop-blur"
    >
      <button
        v-for="tag in filteredSuggestions"
        :key="tag"
        type="button"
        class="block w-full px-4 py-2 text-left text-sm text-[var(--text-primary)] transition hover:bg-[var(--accent)]/20"
        @click="selectTag(tag)"
      >
        {{ tag }}
      </button>
      <button
        v-if="canCreate"
        type="button"
        class="block w-full px-4 py-2 text-left text-sm font-semibold text-[var(--text-primary)] transition hover:bg-[var(--accent)]/20"
        @click="createTag"
      >
        + Create "{{ inputValue.trim() }}"
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';

const props = defineProps({
  modelValue: {
    type: Array,
    default: () => [],
  },
  suggestions: {
    type: Array,
    default: () => [],
  },
  placeholder: {
    type: String,
    default: '(optional)',
  },
});

const emit = defineEmits(['update:modelValue']);

const inputValue = ref('');
const showDropdown = ref(false);
const inputRef = ref(null);
const internalValue = ref([...props.modelValue]);
const containerRef = ref(null);

watch(
  () => props.modelValue,
  (value) => {
    internalValue.value = [...value];
  }
);

const filteredSuggestions = computed(() => {
  const value = inputValue.value.trim().toLowerCase();
  return props.suggestions
    .filter((tag) => !internalValue.value.includes(tag))
    .filter((tag) => tag.toLowerCase().includes(value));
});

const canCreate = computed(() => {
  const value = inputValue.value.trim();
  if (!value) return false;
  return !props.suggestions.map((tag) => tag.toLowerCase()).includes(value.toLowerCase());
});

function openDropdown() {
  showDropdown.value = true;
}

function closeDropdown() {
  showDropdown.value = false;
}

function focusInput() {
  inputRef.value?.focus();
}

function updateValue(value) {
  internalValue.value = value;
  emit('update:modelValue', value);
}

function selectTag(tag) {
  updateValue([...internalValue.value, tag]);
  inputValue.value = '';
  closeDropdown();
}

function createTag() {
  const value = inputValue.value.trim();
  if (!value) return;
  if (!internalValue.value.includes(value)) {
    updateValue([...internalValue.value, value]);
  }
  inputValue.value = '';
  closeDropdown();
}

function removeTag(tag) {
  updateValue(internalValue.value.filter((item) => item !== tag));
}

function onInput() {
  showDropdown.value = true;
}

function handleClickOutside(event) {
  if (!containerRef.value) return;
  if (!containerRef.value.contains(event.target)) {
    closeDropdown();
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>
