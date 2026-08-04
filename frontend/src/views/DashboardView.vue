<template>
  <section class="space-y-8 animate-in fade-in duration-1000 text-[var(--text-primary)] pb-20">
    <!-- Header: Period Selector -->
    <div class="flex flex-col gap-6 md:flex-row md:items-center md:justify-between px-2">
      <div>
        <h1 class="text-4xl font-black tracking-tighter">
          Halo, <span class="text-[var(--accent)]">{{ firstName || 'Bro' }}</span>!
        </h1>
        <p class="text-[10px] font-black text-[var(--text-secondary)] uppercase tracking-[0.4em] mt-2 flex items-center gap-2">
          <span class="inline-block h-2 w-2 rounded-full bg-emerald-500 animate-pulse"></span>
          Summary is ready
        </p>
      </div>

      <div class="flex items-center gap-2 md:gap-3">
        <button
          type="button"
          class="flex h-10 w-10 items-center justify-center rounded-2xl bg-[var(--bg-secondary)] border border-[var(--border)] shadow-2xl glass-card text-[var(--text-primary)] transition-all hover:bg-[var(--bg-elevated)] active:scale-90"
          :title="hideNumbers ? 'Show numbers' : 'Hide numbers'"
          :aria-pressed="hideNumbers"
          @click="toggleHideNumbers"
        >
          <i :class="hideNumbers ? 'fa-solid fa-eye-slash text-xs' : 'fa-solid fa-eye text-xs'"></i>
        </button>
        <div class="flex items-center gap-2 bg-[var(--bg-secondary)] p-1.5 rounded-3xl border border-[var(--border)] shadow-2xl glass-card text-[var(--text-primary)]">
          <button class="flex h-10 w-10 items-center justify-center rounded-2xl transition-all hover:bg-[var(--bg-elevated)] active:scale-90" @click="gotoPrevMonth">
            <i class="fa-solid fa-chevron-left text-xs"></i>
          </button>
          <div class="px-6 text-[10px] font-black uppercase tracking-[0.2em] min-w-[150px] text-center">{{ periodLabel }}</div>
          <button class="flex h-10 w-10 items-center justify-center rounded-2xl transition-all hover:bg-[var(--bg-elevated)] active:scale-90" @click="gotoNextMonth">
            <i class="fa-solid fa-chevron-right text-xs"></i>
          </button>
        </div>
      </div>
    </div>

    <!-- Premium Wallet Card -->
    <div class="relative overflow-hidden rounded-[24px] md:rounded-[48px] bg-[#121212] p-5 md:p-12 text-white shadow-2xl transition-all hover:shadow-indigo-500/10 group mx-1 md:mx-2">
      <div class="absolute -right-20 -top-20 h-80 w-80 rounded-full bg-[var(--accent)] opacity-10 blur-[100px] transition-opacity group-hover:opacity-20"></div>
      <div class="absolute -bottom-20 -left-20 h-80 w-80 rounded-full bg-violet-600 opacity-10 blur-[100px] transition-opacity group-hover:opacity-20"></div>

      <div class="relative z-10 flex flex-col h-full justify-between gap-8 md:gap-12 text-white">
        <div class="flex justify-between items-start">
          <div class="space-y-1 md:space-y-2 min-w-0">
            <p class="text-[10px] font-black uppercase tracking-[0.4em] opacity-60">Balance</p>
            <h2 class="text-2xl md:text-5xl font-black tracking-tight tabular-nums leading-none truncate">
              {{ formatCurrency(balance) }}
            </h2>
          </div>
          <div class="h-12 w-12 md:h-16 md:w-16 bg-white/10 rounded-2xl md:rounded-3xl flex items-center justify-center border border-white/20 backdrop-blur-xl shadow-inner text-white shrink-0 ml-2">
            <i class="fa-solid fa-wallet text-xl md:text-2xl text-white"></i>
          </div>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-6 md:gap-10 pt-6 md:pt-10 border-t border-white/10">
          <div class="group/stat">
            <p class="flex items-center gap-2 text-[10px] font-black uppercase tracking-[0.2em] text-emerald-300 mb-2">Inflow</p>
            <p class="text-lg md:text-2xl font-black tabular-nums text-white truncate">{{ formatCurrency(income) }}</p>
          </div>
          <div class="group/stat">
            <p class="flex items-center gap-2 text-[10px] font-black uppercase tracking-[0.2em] text-rose-300 mb-2">Outflow</p>
            <p class="text-lg md:text-2xl font-black tabular-nums text-white truncate">{{ formatCurrency(totalExpenses) }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Data Layout -->
    <div class="grid gap-8 lg:grid-cols-12 items-start px-2">
      <!-- Analytics Section -->
      <div class="lg:col-span-7">
        <div class="glass-card p-4 md:p-8 rounded-[20px] md:rounded-[32px] h-full flex flex-col shadow-xl min-h-[480px]">
           <div class="flex items-center justify-between mb-6">
              <h3 class="text-xl font-black tracking-tight text-[var(--text-primary)]">Spending Analysis</h3>
              <select v-model="dateFilter" class="bg-[var(--bg-elevated)] text-[10px] font-black uppercase tracking-widest px-4 py-2 rounded-2xl border border-[var(--border)] appearance-none cursor-pointer text-[var(--text-primary)]">
                 <option value="month">Monthly</option>
                 <option value="week">Weekly</option>
                 <option value="today">Today</option>
              </select>
           </div>

           <div v-if="!hasExpenseData" class="flex-1 flex items-center justify-center border-4 border-dashed border-[var(--border)] rounded-[20px] md:rounded-[40px] italic text-[var(--text-secondary)] font-bold uppercase tracking-widest opacity-30 text-[var(--text-primary)]">No Data</div>
           <div v-else class="grid grid-cols-1 lg:grid-cols-[260px_minmax(0,1fr)] items-start gap-6 md:gap-8 flex-1">
              <!-- Stabilized chart container -->
              <div class="relative h-56 w-56 md:h-64 md:w-64 mx-auto lg:mx-0 shrink-0 flex items-center justify-center">
                 <canvas ref="chartCanvas" class="relative z-10"></canvas>
                 <div class="absolute inset-0 flex flex-col items-center justify-center pointer-events-none z-0 text-[var(--text-primary)]">
                    <span class="text-[10px] font-bold uppercase text-[var(--text-secondary)] tracking-wide opacity-80">Total</span>
                    <span class="text-xl md:text-2xl font-black tabular-nums leading-none mt-2">{{ formatCurrency(totalExpenses) }}</span>
                 </div>
              </div>

              <div class="w-full space-y-2 max-h-[360px] overflow-y-auto pr-2 custom-scrollbar text-[var(--text-primary)]">
                <div
                  v-for="entry in legendEntries"
                  :key="entry.category"
                  class="p-3 rounded-xl bg-[var(--bg-elevated)]/30 border border-transparent hover:border-[var(--accent)]/30 transition-all cursor-pointer group"
                  :class="[entry.disabled && 'opacity-35 grayscale']"
                  @click="toggleCategory(entry.category)"
                >
                  <div class="flex items-center gap-2.5 min-w-0">
                    <div class="h-3.5 w-3.5 rounded-full shrink-0" :style="{ backgroundColor: entry.color }"></div>
                    <span class="text-sm font-black leading-tight break-words">{{ entry.category }}</span>
                  </div>
                  <div class="mt-1.5 flex items-end justify-between gap-3">
                    <span class="text-xs font-semibold text-[var(--text-secondary)]">{{ formatPercent(entry.percentage) }}</span>
                    <span class="text-base md:text-lg font-black tabular-nums leading-none tracking-tight whitespace-nowrap">{{ formatCurrency(entry.total) }}</span>
                  </div>
                </div>
              </div>
           </div>
        </div>
      </div>

      <!-- Budget Column -->
      <div class="lg:col-span-5 text-[var(--text-primary)]">
        <div v-if="budgetSummaries.length" class="glass-card p-6 md:p-10 rounded-[24px] md:rounded-[48px] h-full flex flex-col shadow-xl border-t-8 border-t-[var(--accent)] text-[var(--text-primary)]">
          <div class="flex items-start justify-between mb-12">
            <h3 class="text-2xl font-black tracking-tighter">Budgets</h3>
            <div v-if="overallBudgetStatus" :class="['text-[9px] font-black uppercase px-3 py-1 rounded-xl border-2', overallBudgetStatus.className]">{{ overallBudgetStatus.label }}</div>
          </div>

          <div class="space-y-10 flex-1 overflow-y-auto pr-2 custom-scrollbar mt-12">
            <div v-for="item in budgetSummaries" :key="item.id" class="space-y-4 group">
              <div class="flex items-end justify-between px-1 text-[var(--text-primary)]">
                <div class="flex flex-col">
                  <span class="text-lg font-black tracking-tight">{{ item.category }}</span>
                  <span class="text-[9px] font-black uppercase text-[var(--text-secondary)] mt-1 tracking-widest">{{ item.statusLabel }}</span>
                </div>
                <div class="text-right">
                   <div class="text-sm font-black">{{ formatCurrency(item.actual) }}</div>
                   <div class="text-[9px] font-bold text-[var(--text-secondary)] uppercase">of {{ formatCurrency(item.amount) }}</div>
                </div>
              </div>
              <div class="h-4 w-full rounded-full bg-[var(--bg-elevated)] p-1 border border-[var(--border)] overflow-hidden shadow-inner">
                <div
                  class="h-full rounded-full transition-all duration-1000 relative"
                  :class="[
                    item.status === 'over' ? 'bg-gradient-to-r from-rose-700 to-rose-400' : 
                    item.status === 'warning' ? 'bg-gradient-to-r from-orange-500 to-amber-300' : 
                    'bg-gradient-to-r from-emerald-600 to-teal-400'
                  ]"
                  :style="{ width: `${Math.min(item.percentage, 100)}%` }"
                >
                  <div class="absolute inset-0 bg-white/20 animate-pulse mix-blend-overlay"></div>
                </div>
              </div>
            </div>
          </div>
          <button @click="router.push('/settings')" class="w-full py-4 mt-10 rounded-3xl bg-[var(--bg-elevated)] font-black uppercase text-[10px] transition-all hover:bg-[var(--bg-secondary)] border border-[var(--border)] shadow-xl text-[var(--text-primary)]">Adjust Limits</button>
        </div>
        <div v-else class="glass-card p-6 md:p-12 rounded-[24px] md:rounded-[48px] flex flex-col items-center justify-center text-center gap-8 border-4 border-dashed border-[var(--border)] h-full opacity-60 text-[var(--text-primary)]">
           <i class="fa-solid fa-piggy-bank text-4xl text-[var(--accent)]"></i>
           <h3 class="text-xl font-black uppercase tracking-widest">No Budgets Set</h3>
           <button @click="router.push('/settings')" class="btn-primary px-8 h-12">Set Up Now</button>
        </div>
      </div>
    </div>

    <!-- Quick Add FAB -->
    <div class="fixed bottom-24 right-4 md:bottom-10 md:right-10 z-40 flex flex-col gap-3 md:gap-4">
       <button @click="handleOpenTyping" class="h-16 w-16 bg-white dark:bg-slate-800 text-[var(--accent)] rounded-full shadow-2xl flex items-center justify-center border border-[var(--border)] hover:scale-110 active:scale-95 transition-all"><i class="fa-solid fa-keyboard text-xl"></i></button>
       <button @click="handleOpenManual" class="h-20 w-20 bg-[var(--accent)] text-white rounded-[16px] md:rounded-[32px] shadow-2xl flex items-center justify-center shadow-indigo-500/40 hover:scale-110 active:scale-95 transition-all"><i class="fa-solid fa-plus text-3xl"></i></button>
    </div>

    <!-- Manual Form Overlay -->
    <div v-if="showExpenseForm" class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-xl px-4" @click.self="handleCloseAddPanels">
      <div class="w-full max-w-4xl glass-card rounded-[24px] md:rounded-[48px] p-6 md:p-12 shadow-2xl relative border-2 border-[var(--accent)] animate-in zoom-in duration-500 text-[var(--text-primary)]">
        <button @click="handleCloseAddPanels" class="absolute top-8 right-8 h-12 w-12 rounded-2xl bg-[var(--bg-elevated)] text-[var(--text-secondary)] flex items-center justify-center hover:text-rose-500 transition-colors shadow-lg"><i class="fa-solid fa-xmark text-xl"></i></button>
        <h3 class="text-3xl font-black tracking-tighter mb-10">New Transaction</h3>
        <form class="grid gap-10 md:grid-cols-2 text-[var(--text-primary)]" @submit.prevent="submitExpense">
          <div class="space-y-3"><label class="text-[11px] font-black uppercase text-[var(--text-secondary)]">Description</label><input v-model="form.name" type="text" class="input-modern w-full text-xl h-16 text-[var(--text-primary)]" placeholder="E.g. Coffee" required /></div>
          <div class="space-y-3"><label class="text-[11px] font-black uppercase text-[var(--text-secondary)]">Category</label><select v-model="form.category" class="input-modern w-full h-16 font-bold text-[var(--text-primary)]" required><option v-for="c in categories" :key="c" :value="c">{{ c }}</option></select></div>
          <div class="space-y-3"><label class="text-[11px] font-black uppercase text-[var(--text-secondary)]">Amount</label><div class="relative"><input :value="formattedAmount" @input="handleAmountInput" @blur="normalizeAmount" inputmode="decimal" class="input-modern w-full h-16 text-3xl font-black text-[var(--text-primary)] pl-6 tabular-nums" required /><span class="absolute right-6 top-1/2 -translate-y-1/2 text-xs font-black text-[var(--accent)] opacity-40 uppercase tracking-widest">IDR</span></div></div>
          <div class="space-y-3"><label class="text-[11px] font-black uppercase text-[var(--text-secondary)]">Date</label><input v-model="form.date" type="date" class="input-modern w-full h-16 font-bold text-[var(--text-primary)]" required /></div>
          <div class="md:col-span-2 pt-6"><button type="submit" class="btn-primary w-full h-20 text-2xl font-black shadow-2xl text-white">Confirm Record</button></div>
        </form>
      </div>
    </div>

    <!-- Smart Add Overlay -->
    <QuickAddExpenseCard v-if="showTypingForm" ref="typingCardRef" :card-class="'fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-3xl px-6'" :input-class="'input-modern w-full h-24 text-3xl md:text-4xl font-black text-[var(--text-primary)] px-4 md:px-10 shadow-2xl rounded-[20px] md:rounded-[48px] border-4 border-indigo-600 text-center placeholder:text-slate-700'" :primary-button-class="'hidden'" @added="handleQuickAddSuccess" @close="handleCloseAddPanels" />
  </section>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Chart, registerables } from 'chart.js';
import state, { loadInitialData, refreshExpenses, refreshBudgetSummaries } from '../stores/appState';
import QuickAddExpenseCard from '../components/QuickAddExpenseCard.vue';
import { formatMonth, getMonthExpenses, filterExpensesByRange, formatCurrency as formatCurrencyRaw, maskCurrency as maskCurrencyRaw, getISODateWithLocalTime, colorPalette, getCycleAnchor } from '../lib/utils';
import { apiFetch } from '../lib/api';
import { encryptPayload } from '../lib/encryption';

Chart.register(...registerables); 
Chart.defaults.color = '#8b949e'; 
Chart.defaults.borderColor = 'transparent';
Chart.defaults.font.family = 'Inter, sans-serif';

const chartCanvas = ref(null); 
let chartInstance = null; 
const route = useRoute();
const router = useRouter();
const hideNumbers = ref(false);
const NUMBER_VISIBILITY_QUERY_KEY = 'hideNumbers';
const NUMBER_VISIBILITY_STORAGE_KEY = 'expenseowl_dashboard_hide_numbers';

const currentDate = ref(new Date()); 
const monthCursor = ref(new Date()); 
const dateFilter = ref('month');
const showExpenseForm = ref(false); 
const showTypingForm = ref(false); 
const disabledCategories = ref(new Set());
const categoryColors = ref({}); 
const rangeStart = ref(new Date().toISOString().split('T')[0]); 
const rangeEnd = ref(new Date().toISOString().split('T')[0]);

const form = ref({ name: '', category: '', tags: [], amount: null, date: new Date().toISOString().split('T')[0], reportGain: false });
const formMessage = ref({ text: '', type: '' }); 
const rawAmount = ref(''); 
const typingCardRef = ref(null);

const firstName = computed(() => state.user?.firstName || ''); 
const categories = computed(() => state.categories);

const formattedAmount = computed(() => { 
  if (!rawAmount.value) return ''; 
  const numeric = Number(rawAmount.value.replace(/[^0-9.-]/g, '')) || 0;
  return new Intl.NumberFormat('en-US').format(numeric); 
});

function readHideNumbersPreference() {
  const raw = route.query[NUMBER_VISIBILITY_QUERY_KEY];
  const value = Array.isArray(raw) ? raw[0] : raw;
  if (typeof value === 'string') {
    const normalized = value.trim().toLowerCase();
    if (normalized === '') return true;
    if (['1', 'true', 'yes', 'on'].includes(normalized)) return true;
    if (['0', 'false', 'no', 'off'].includes(normalized)) return false;
  }
  if (typeof window !== 'undefined') {
    return localStorage.getItem(NUMBER_VISIBILITY_STORAGE_KEY) === '1';
  }
  return false;
}

function persistHideNumbersPreference(value) {
  hideNumbers.value = !!value;
  if (typeof window !== 'undefined') {
    localStorage.setItem(NUMBER_VISIBILITY_STORAGE_KEY, hideNumbers.value ? '1' : '0');
  }
}

function syncHideNumbersPreference() {
  persistHideNumbersPreference(readHideNumbersPreference());
}

function toggleHideNumbers() {
  const nextValue = !hideNumbers.value;
  persistHideNumbersPreference(nextValue);
  const query = { ...route.query };
  if (nextValue) query[NUMBER_VISIBILITY_QUERY_KEY] = '1';
  else delete query[NUMBER_VISIBILITY_QUERY_KEY];
  router.replace({ path: route.path, query });
}

const displayedExpenses = computed(() => {
  if (dateFilter.value === 'month') return getMonthExpenses(state.expenses, monthCursor.value, state.startDate, state.endOfMonth);
  if (dateFilter.value === 'range') return filterExpensesByRange(state.expenses, 'range', currentDate.value, state.startDate, state.endOfMonth, { start: rangeStart.value, end: rangeEnd.value });
  return filterExpensesByRange(state.expenses, dateFilter.value, currentDate.value, state.startDate, state.endOfMonth);
});

const hasExpenseData = computed(() => displayedExpenses.value.some(e => e.amount < 0));
const income = computed(() => displayedExpenses.value.filter(e => e.amount > 0).reduce((s, e) => s + e.amount, 0));
const totalExpenses = computed(() => displayedExpenses.value.filter(e => e.amount < 0).reduce((s, e) => s + Math.abs(e.amount), 0));
const balance = computed(() => income.value - totalExpenses.value);

const periodLabel = computed(() => {
  if (dateFilter.value === 'today') return 'Today';
  if (dateFilter.value === 'week') return 'This Week';
  if (dateFilter.value === 'range') return 'Range';
  return formatMonth(monthCursor.value);
});

const legendEntries = computed(() => {
  const breakdown = calculateCategoryBreakdown(displayedExpenses.value);
  return breakdown.map(item => ({
    ...item,
    color: categoryColors.value[item.category],
    amountFormatted: formatCurrency(item.total),
    disabled: disabledCategories.value.has(item.category)
  }));
});

const budgetSummaries = computed(() => {
  const t = getMonthExpenses(state.expenses, monthCursor.value, state.startDate, state.endOfMonth).reduce((a, e) => { if (e.amount < 0) { const k = e.category || 'Misc'; a[k] = (a[k] || 0) + Math.abs(e.amount); } return a; }, {});
  return (state.budgetSummaries || []).map(b => { 
    const a = t[b.category] || 0; 
    const l = b.effectiveAmount || b.amount; 
    const p = l > 0 ? (a / l) * 100 : 0; 
    let s = 'ok'; 
    if (a > l) s = 'over'; 
    else if (p > 80) s = 'warning'; 
    return { ...b, actual: a, amount: l, percentage: p, status: s, statusLabel: s.toUpperCase() }; 
  }).sort((a,b) => b.actual - a.actual);
});

const overallBudgetStatus = computed(() => { 
  const l = budgetSummaries.value.reduce((s,i) => s+i.amount, 0); 
  const a = budgetSummaries.value.reduce((s,i) => s+i.actual, 0); 
  if (l === 0) return null; 
  if (a > l) return { label: 'BONCOS BRO', className: 'text-rose-500 border-rose-500/30' }; 
  if (a > l * 0.8) return { label: 'WARNING', className: 'text-amber-500 border-amber-500/30' }; 
  return { label: 'ON TRACK', className: 'text-emerald-500 border-emerald-500/30' }; 
});

function calculateCategoryBreakdown(ex) { 
  const ts = {}; let o = 0; 
  ex.forEach(e => { 
    if (e.amount < 0) { 
      const a = Math.abs(e.amount); 
      const c = normalizeCategoryName(e.category); 
      if (!disabledCategories.value.has(c)) { 
        ts[c] = (ts[c] || 0) + a; 
        o += a; 
      } 
    } 
  }); 
  return Object.entries(ts).map(([category, total]) => ({ category, total, percentage: o > 0 ? (total/o)*100 : 0 })).sort((a,b) => b.total - a.total); 
}

function normalizeCategoryName(c) { return (typeof c === 'string' && c.trim()) || 'Misc'; }

function assignCategoryColors() { 
  const cs = { ...categoryColors.value }; 
  const cats = Array.from(new Set([...(state.categories || []), ...state.expenses.map(e => normalizeCategoryName(e.category))])); 
  cats.forEach((c, i) => { if (!cs[c]) cs[c] = colorPalette[i % colorPalette.length]; }); 
  categoryColors.value = cs; 
}

function alignCurrentCycle(d = currentDate.value) { 
  const a = getCycleAnchor(d, state.startDate, state.endOfMonth); 
  monthCursor.value = new Date(a); 
  return a; 
}

async function gotoPrevMonth() { 
  const d = new Date(monthCursor.value); d.setMonth(d.getMonth() - 1); 
  alignCurrentCycle(d); 
  await refreshBudgetSummaries(monthCursor.value); 
  updateChart(); 
}

async function gotoNextMonth() { 
  const d = new Date(monthCursor.value); d.setMonth(d.getMonth() + 1); 
  alignCurrentCycle(d); 
  await refreshBudgetSummaries(monthCursor.value); 
  updateChart(); 
}

function formatCurrency(amt) { return hideNumbers.value ? maskCurrencyRaw(amt, state.currency) : formatCurrencyRaw(amt, state.currency); }
function formatPercent(value) { return hideNumbers.value ? '\u2022\u2022\u2022' : `${Number(value || 0).toFixed(1)}%`; }
function handleAmountInput(e) { rawAmount.value = e.target.value.replace(/[^0-9.-]/g, ''); }
function normalizeAmount(e) { const n = Number(rawAmount.value) || 0; rawAmount.value = n === 0 ? '' : String(n); e.target.value = formattedAmount.value; }
function toggleCategory(c) { const n = new Set(disabledCategories.value); if (n.has(c)) n.delete(c); else n.add(c); disabledCategories.value = n; }

async function submitExpense() {
  if (!form.value.category) return setFormMessage('Choose category', 'error'); 
  const amt = Number(rawAmount.value.replace(/[^0-9.-]/g, '')) || 0; if (amt === 0) return setFormMessage('Invalid amount', 'error');
  const p = { name: form.value.name, category: form.value.category, amount: form.value.reportGain ? Math.abs(amt) : -Math.abs(amt), date: getISODateWithLocalTime(form.value.date), tags: form.value.tags };
  try { 
    const b = await encryptPayload(p); if (b) p.blob = b; 
    const r = await apiFetch('/expense', { method: 'PUT', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(p) }); 
    if (!r.ok) throw new Error('Failed'); 
    setFormMessage('Success!', 'success'); resetForm(); await refreshExpenses(); handleCloseAddPanels(); 
  } catch (e) { setFormMessage(e.message, 'error'); }
}

function handleOpenManual() { showExpenseForm.value = true; }
function handleOpenTyping() { showTypingForm.value = true; }
function handleCloseAddPanels() { showExpenseForm.value = false; showTypingForm.value = false; resetForm(); }
function handleQuickAddSuccess() { refreshExpenses(); handleCloseAddPanels(); }
function setFormMessage(t, type) { formMessage.value = { text: t, type }; setTimeout(() => formMessage.value = { text: '', type: '' }, 3000); }

function updateChart() {
  if (!chartCanvas.value || !hasExpenseData.value) { if (chartInstance) chartInstance.destroy(); return; }
  const b = calculateCategoryBreakdown(displayedExpenses.value); 
  if (chartInstance) chartInstance.destroy();
  chartInstance = new Chart(chartCanvas.value, { 
    type: 'doughnut', 
    data: { 
      labels: b.map(i => i.category), 
      datasets: [{ 
        data: b.map(i => i.total), 
        backgroundColor: b.map(i => categoryColors.value[i.category]), 
        borderColor: 'transparent', 
        cutout: '74%', 
        borderRadius: 8, 
        spacing: 2,
        hoverOffset: 6 
      }] 
    }, 
    options: { 
      responsive: true, 
      maintainAspectRatio: false, 
      layout: { padding: 10 },
      animation: { duration: 900, easing: 'easeOutCubic' }, 
      plugins: { 
        legend: { display: false }, 
        tooltip: { 
          backgroundColor: '#000', 
          padding: 16, 
          titleFont: { size: 14, weight: '900' }, 
          bodyFont: { size: 12, weight: 'bold' }, 
          cornerRadius: 16, 
          displayColors: true,
          callbacks: { label: (c) => ` ${formatCurrency(c.raw)} (${hideNumbers.value ? '\u2022\u2022\u2022' : ((c.raw/b.reduce((s,i)=>s+i.total,0))*100).toFixed(1) + '%'})` } 
        } 
      } 
    } 
  });
}

watch(displayedExpenses, () => { assignCategoryColors(); updateChart(); });
watch(() => route.query[NUMBER_VISIBILITY_QUERY_KEY], syncHideNumbersPreference);
onMounted(async () => { syncHideNumbersPreference(); await loadInitialData(); alignCurrentCycle(); await refreshBudgetSummaries(monthCursor.value); assignCategoryColors(); updateChart(); });
onBeforeUnmount(() => { if (chartInstance) chartInstance.destroy(); });
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 3px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: var(--border); border-radius: 10px; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(30px); } to { opacity: 1; transform: translateY(0); } }
.animate-in { animation: fadeIn 1s cubic-bezier(0.16, 1, 0.3, 1) fill-mode-both; }
.fade-enter-active, .fade-leave-active { transition: all 0.4s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; transform: scale(0.95); }
</style>
