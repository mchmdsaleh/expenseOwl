<template>
  <div class="relative">
    <div class="hidden items-center justify-end gap-2 md:flex">
      <div class="relative">
        <button
          v-if="!anyPanelOpen"
          ref="desktopTriggerRef"
          :class="primaryButtonClass"
          @click.stop="toggleMenu"
        >
          <i class="fa-solid fa-plus"></i>
          Add Expense
        </button>
        <button
          v-else
          :class="primaryButtonClass"
          @click.stop="emitCloseAll"
        >
          <i class="fa-solid fa-times"></i>
          Close
        </button>
        <transition name="fade">
          <div
            v-if="showMenu && !anyPanelOpen"
            ref="desktopMenuRef"
            class="absolute right-0 z-30 mt-2 w-44 space-y-2 rounded-2xl border border-[var(--border)] bg-[var(--bg-primary)]/95 p-3 shadow-lg backdrop-blur"
            @click.stop
          >
            <button type="button" :class="speedDialButtonClass" @click="handleOpenTyping">
              <i class="fa-solid fa-keyboard text-xs"></i>
              <span>Typing</span>
            </button>
            <button type="button" :class="speedDialButtonClass" @click="handleOpenManual">
              <i class="fa-solid fa-pen-to-square text-xs"></i>
              <span>Manual</span>
            </button>
          </div>
        </transition>
      </div>
    </div>

    <transition name="fade">
      <div
        v-if="showMenu && !anyPanelOpen"
        ref="mobileMenuRef"
        class="fixed bottom-24 right-6 z-40 flex flex-col gap-2 rounded-2xl border border-[var(--border)] bg-[var(--bg-primary)]/95 p-3 shadow-lg backdrop-blur md:hidden"
        @click.stop
      >
        <button type="button" :class="speedDialButtonClass" @click="handleOpenTyping">
          <i class="fa-solid fa-keyboard text-xs"></i>
          <span>Typing</span>
        </button>
        <button type="button" :class="speedDialButtonClass" @click="handleOpenManual">
          <i class="fa-solid fa-pen-to-square text-xs"></i>
          <span>Manual</span>
        </button>
      </div>
    </transition>

    <button
      ref="fabRef"
      class="fixed bottom-6 right-6 z-40 inline-flex h-14 w-14 items-center justify-center rounded-full bg-[var(--accent)] text-xl text-white shadow-2xl transition duration-150 ease-out hover:scale-105 focus:outline-none focus:ring-4 focus:ring-[var(--accent)]/30 md:hidden"
      @click.stop="toggleFab"
    >
      <i :class="anyPanelOpen ? 'fa-solid fa-times' : 'fa-solid fa-plus'"></i>
    </button>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';

const props = defineProps({
  manualOpen: {
    type: Boolean,
    default: false,
  },
  typingOpen: {
    type: Boolean,
    default: false,
  },
  primaryButtonClass: {
    type: String,
    default:
      'inline-flex items-center justify-center gap-2 rounded-full border border-[var(--border)] bg-[var(--bg-secondary)] px-5 py-2 text-sm font-medium text-[var(--text-primary)] transition duration-150 ease-out hover:bg-[var(--accent)] hover:text-white hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40 focus:ring-offset-2 focus:ring-offset-[var(--bg-primary)] disabled:cursor-not-allowed disabled:opacity-50',
  },
  speedDialButtonClass: {
    type: String,
    default:
      'flex items-center gap-2 rounded-full border border-[var(--border)] bg-[var(--bg-secondary)]/90 px-4 py-2 text-sm font-medium text-[var(--text-primary)] transition duration-150 ease-out hover:bg-[var(--accent)] hover:text-white hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40 focus:ring-offset-2 focus:ring-offset-[var(--bg-primary)]',
  },
});

const emits = defineEmits(['open-manual', 'open-typing', 'close-all']);

const showMenu = ref(false);
const desktopTriggerRef = ref(null);
const desktopMenuRef = ref(null);
const mobileMenuRef = ref(null);
const fabRef = ref(null);

const anyPanelOpen = computed(() => props.manualOpen || props.typingOpen);

watch(anyPanelOpen, (value) => {
  if (value) {
    showMenu.value = false;
  }
});

function toggleMenu() {
  showMenu.value = !showMenu.value;
}

function toggleFab() {
  if (anyPanelOpen.value) {
    emitCloseAll();
    return;
  }
  showMenu.value = !showMenu.value;
}

function handleOpenManual() {
  showMenu.value = false;
  emits('open-manual');
}

function handleOpenTyping() {
  showMenu.value = false;
  emits('open-typing');
}

function emitCloseAll() {
  showMenu.value = false;
  emits('close-all');
}

function handleDocumentClick(event) {
  if (!showMenu.value) return;
  const target = event.target;
  if (!(target instanceof Node)) return;
  const containers = [
    desktopTriggerRef.value,
    desktopMenuRef.value,
    mobileMenuRef.value,
    fabRef.value,
  ];
  const clickedInside = containers.some((element) => element && element.contains(target));
  if (!clickedInside) {
    showMenu.value = false;
  }
}

onMounted(() => {
  document.addEventListener('click', handleDocumentClick);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleDocumentClick);
});
</script>
