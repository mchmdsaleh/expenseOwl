<template>
  <div class="tags-input-container" @click="focusInput">
    <div class="selected-tags">
      <div v-for="tag in internalValue" :key="tag" class="tag-pill">
        {{ tag }}
        <span class="remove-tag" @click.stop="removeTag(tag)">×</span>
      </div>
    </div>
    <input
      ref="inputRef"
      v-model="inputValue"
      type="text"
      :placeholder="placeholder"
      @focus="openDropdown"
      @input="onInput"
      @keydown.enter.prevent="createTag"
    />
    <div v-if="showDropdown" class="tags-dropdown">
      <div
        v-for="tag in filteredSuggestions"
        :key="tag"
        @click="selectTag(tag)"
      >
        {{ tag }}
      </div>
      <div v-if="canCreate" class="new-tag" @click="createTag">
        + Create "{{ inputValue.trim() }}"
      </div>
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
  if (!event.composedPath().includes(inputRef.value?.parentElement)) {
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
