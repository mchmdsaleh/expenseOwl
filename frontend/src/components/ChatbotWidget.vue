<template>
  <div class="fixed z-50 flex flex-col transition-all duration-300" :class="[isOpen ? 'inset-0 md:inset-auto md:bottom-8 md:right-8 items-end' : 'bottom-6 left-6 md:left-auto md:right-8 md:bottom-8 items-start md:items-end']">
    <!-- Chat Window -->
    <div v-if="isOpen" class="flex flex-col w-full h-full md:w-96 md:h-[600px] bg-white dark:bg-slate-900 md:rounded-2xl shadow-2xl border-slate-200 dark:border-slate-800 overflow-hidden border">
      <!-- Header -->
      <div class="p-4 bg-gradient-to-r from-indigo-600 via-indigo-500 to-violet-600 text-white flex justify-between items-center shadow-lg">
        <div class="flex items-center space-x-3">
          <div class="bg-white/20 p-2 rounded-xl backdrop-blur-sm">
            <i class="fa-solid fa-robot text-xl"></i>
          </div>
          <div>
            <h3 class="font-bold text-sm tracking-wide">ExpenseOwl AI</h3>
            <p class="text-[10px] text-indigo-100 flex items-center mt-0.5">
              <span class="w-1.5 h-1.5 bg-green-400 rounded-full mr-1.5 animate-pulse"></span>
              Online & Ready to Help
            </p>
          </div>
        </div>
        <div class="flex items-center space-x-2">
          <button @click="sendInsightRequest" :disabled="isLoading" class="flex items-center space-x-1.5 bg-white/10 hover:bg-white/20 px-2.5 py-1 rounded-lg transition-all text-[10px] font-bold border border-white/10 disabled:opacity-50 active:scale-95 mr-1">
            <i class="fa-solid fa-wand-magic-sparkles text-[10px]"></i>
            <span class="hidden sm:inline">Insights</span>
          </button>
          <button @click="isOpen = false" class="hover:bg-white/10 p-2 rounded-full transition-all active:scale-90">
            <i class="fa-solid fa-xmark text-lg"></i>
          </button>
        </div>
      </div>

      <!-- Messages -->
      <div ref="messageContainer" class="flex-1 overflow-y-auto p-4 space-y-4 bg-slate-50/50 dark:bg-slate-900/50">
        <div v-for="(msg, idx) in chatHistory" :key="idx" :class="['flex', msg.role === 'user' ? 'justify-end' : 'justify-start']">
          <div :class="['max-w-[85%] rounded-2xl p-3 text-sm shadow-sm transition-all', msg.role === 'user' ? 'bg-indigo-600 text-white rounded-tr-none' : 'bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-200 border border-slate-100 dark:border-slate-700 rounded-tl-none']">
            <!-- Tool Calls (e.g. Preview Parsed Data) -->
            <div v-if="msg.tool_calls">
              <div v-for="call in msg.tool_calls" :key="call.id">
                <div v-if="call.function.name === 'preview_parsed_data'" class="space-y-3 mt-2">
                  <div class="flex items-center space-x-2 text-indigo-500 dark:text-indigo-400">
                    <i class="fa-solid fa-table-list text-xs"></i>
                    <p class="font-bold text-[10px] uppercase tracking-widest">Preview Data yang Diolah</p>
                  </div>
                  <div class="overflow-x-auto border border-slate-200 dark:border-slate-700 rounded-xl bg-white dark:bg-slate-900 shadow-inner">
                    <table class="min-w-full text-[10px] text-left">
                      <thead class="bg-slate-50 dark:bg-slate-800/50 text-slate-500">
                        <tr>
                          <th class="px-2 py-2">Name</th>
                          <th class="px-2 py-2">Category</th>
                          <th class="px-2 py-2">Amount</th>
                          <th class="px-2 py-2">Date</th>
                          <th class="px-2 py-2"></th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="(exp, eIdx) in getEditableExpenses(call.id, call.function.arguments)" :key="eIdx" class="border-t border-slate-100 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/30 transition-colors">
                          <td class="px-2 py-2">
                            <input v-model="exp.name" class="w-full min-w-[90px] bg-transparent border-none focus:ring-1 focus:ring-indigo-500 rounded px-1 py-0.5 text-[10px]" />
                          </td>
                          <td class="px-2 py-2">
                            <input v-model="exp.category" class="w-full min-w-[80px] bg-transparent border-none focus:ring-1 focus:ring-indigo-500 rounded px-1 py-0.5 text-[10px]" />
                          </td>
                          <td class="px-2 py-2">
                            <input :value="exp.amount" @input="setExpenseAmount(call.id, eIdx, $event.target.value)" class="w-full min-w-[90px] font-mono bg-transparent border-none focus:ring-1 focus:ring-indigo-500 rounded px-1 py-0.5 text-[10px]" />
                          </td>
                          <td class="px-2 py-2">
                            <input :value="toInputDate(exp.date)" @input="setExpenseDate(call.id, eIdx, $event.target.value)" type="date" class="w-full min-w-[120px] bg-transparent border-none focus:ring-1 focus:ring-indigo-500 rounded px-1 py-0.5 text-[10px]" />
                          </td>
                          <td class="px-2 py-2 text-right">
                            <button @click="removePreviewExpense(call.id, eIdx)" class="p-1 text-rose-400 hover:text-rose-600 transition-colors"><i class="fa-solid fa-trash-can"></i></button>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                  <div class="flex justify-between items-center px-1">
                    <button @click="addPreviewExpense(call.id, call.function.arguments)" class="flex items-center space-x-1 text-[10px] text-slate-500 hover:text-indigo-500 transition-colors font-medium">
                      <i class="fa-solid fa-plus"></i>
                      <span>Add Row</span>
                    </button>
                  </div>
                  <div class="flex justify-end space-x-2 pt-2 border-t border-slate-100 dark:border-slate-800">
                    <button @click="cancelAction(idx)" class="px-4 py-1.5 bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 rounded-lg text-xs font-bold hover:bg-slate-200 transition-all">Cancel</button>
                    <button @click="confirmBatch(call.id, call.function.arguments)" class="px-4 py-1.5 bg-gradient-to-r from-emerald-500 to-teal-600 text-white rounded-lg text-xs font-bold shadow-lg shadow-emerald-500/20 hover:scale-105 transition-all">Confirm & Save</button>
                  </div>
                </div>
                <!-- Other tools like add_expense, add_budget -->
                <div v-else-if="['add_expense', 'add_budget'].includes(call.function.name)" class="mt-2 p-3 bg-indigo-50 dark:bg-indigo-900/30 rounded-xl border border-indigo-100 dark:border-indigo-800 shadow-sm">
                   <div class="flex items-center space-x-2 text-indigo-600 dark:text-indigo-400 mb-2">
                     <i class="fa-solid fa-wand-magic-sparkles text-xs"></i>
                     <p class="text-[10px] font-bold uppercase tracking-widest">AI Recommendation</p>
                   </div>
                   <p class="text-xs font-medium text-slate-700 dark:text-slate-300">I'd like to perform action: <span class="text-indigo-600">{{ call.function.name }}</span></p>
                   <pre class="text-[10px] bg-white dark:bg-slate-900 p-2 mt-2 rounded-lg border border-indigo-100/50 dark:border-indigo-800/50 overflow-x-auto">{{ call.function.arguments }}</pre>
                   <div class="flex justify-end space-x-2 mt-4">
                    <button @click="cancelAction(idx)" class="px-4 py-1.5 bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 rounded-lg text-xs font-bold hover:bg-slate-200 transition-all">Ignore</button>
                    <button @click="confirmAction(call.function.name, call.function.arguments)" class="px-4 py-1.5 bg-indigo-600 text-white rounded-lg text-xs font-bold shadow-lg shadow-indigo-500/20 hover:scale-105 transition-all">Execute Action</button>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="whitespace-pre-wrap leading-relaxed">{{ msg.content }}</div>
          </div>
        </div>
        <div v-if="isLoading" class="flex justify-start">
          <div class="bg-white dark:bg-slate-800 p-3 rounded-2xl rounded-tl-none shadow-sm border border-slate-100 dark:border-slate-700">
            <div class="flex space-x-1.5">
              <div class="w-2 h-2 bg-indigo-400 rounded-full animate-bounce"></div>
              <div class="w-2 h-2 bg-indigo-400 rounded-full animate-bounce [animation-delay:0.2s]"></div>
              <div class="w-2 h-2 bg-indigo-400 rounded-full animate-bounce [animation-delay:0.4s]"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Attachment Preview -->
      <div v-if="attachments.length > 0" class="px-4 py-2 border-t border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-slate-900 flex flex-wrap gap-2">
        <div v-for="(file, fIdx) in attachments" :key="fIdx" class="flex items-center bg-white dark:bg-slate-800 border border-indigo-100 dark:border-indigo-900/50 px-2.5 py-1.5 rounded-lg text-[10px] text-slate-600 dark:text-slate-300 shadow-sm group">
          <i class="fa-solid fa-file-invoice text-indigo-500 mr-2"></i>
          <span class="truncate max-w-[120px] font-medium">{{ file.name }}</span>
          <button @click="removeAttachment(fIdx)" class="ml-2 text-slate-300 hover:text-rose-500 transition-colors"><i class="fa-solid fa-circle-xmark"></i></button>
        </div>
      </div>

      <!-- Input -->
      <div class="p-4 border-t border-slate-200 dark:border-slate-800 flex items-end space-x-2 bg-white dark:bg-slate-900">
        <label class="cursor-pointer p-2.5 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-xl transition-colors text-slate-500 group">
          <i class="fa-solid fa-paperclip text-lg group-hover:rotate-12 transition-transform"></i>
          <input type="file" multiple class="hidden" @change="handleFileUpload" accept=".csv,image/*" />
        </label>
        <textarea v-model="userInput" @keydown.enter.exact.prevent="sendMessage" placeholder="Ask me anything..." rows="1" class="flex-1 bg-slate-100 dark:bg-slate-800 border-none rounded-xl p-2.5 text-sm focus:ring-2 focus:ring-indigo-500 resize-none max-h-32 transition-all" ref="inputField"></textarea>
        <button @click="sendMessage" :disabled="isLoading || (!userInput.trim() && attachments.length === 0)" class="p-2.5 bg-gradient-to-r from-indigo-600 to-violet-600 text-white rounded-xl hover:shadow-lg hover:shadow-indigo-500/30 transition-all disabled:opacity-50 active:scale-95">
          <i class="fa-solid fa-paper-plane"></i>
        </button>
      </div>
    </div>

    <!-- Toggle Button -->
    <button 
      v-if="!isOpen"
      @click="toggleChat" 
      class="w-14 h-14 bg-gradient-to-br from-indigo-500 via-indigo-600 to-violet-600 text-white rounded-full shadow-2xl flex items-center justify-center hover:shadow-indigo-500/40 transition-all transform hover:scale-110 active:scale-95 border-2 border-white dark:border-slate-800 group relative"
    >
      <div class="absolute -top-1 -right-1 w-4 h-4 bg-rose-500 rounded-full border-2 border-white dark:border-slate-800 scale-0 group-hover:scale-100 transition-transform shadow-sm"></div>
      <i class="fa-solid fa-comment-dots text-2xl group-hover:rotate-12 transition-transform"></i>
    </button>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue';
import { sendChatMessage, apiFetch } from '../lib/api';

const isOpen = ref(false);
const isLoading = ref(false);
const userInput = ref('');
const attachments = ref([]);
const chatHistory = ref([
  { role: 'assistant', content: 'Halo! Saya AI ExpenseOwl. Ada yang bisa saya bantu hari ini? Anda bisa mengupload mutasi bank CSV atau foto nota untuk diolah.' }
]);
const messageContainer = ref(null);
const inputField = ref(null);
const editablePreviewExpenses = ref({});

const toggleChat = () => {
  isOpen.value = !isOpen.value;
  if (isOpen.value) {
    nextTick(() => inputField.value?.focus());
  }
};

const handleFileUpload = (e) => {
  const files = Array.from(e.target.files);
  attachments.value.push(...files);
};

const removeAttachment = (idx) => {
  attachments.value.splice(idx, 1);
};

const sendInsightRequest = () => {
  userInput.value = "Tolong berikan analisis singkat pengeluaran gw bulan ini dibanding budget. Kasih tau kalau ada yang boros ya.";
  sendMessage();
};

const sendMessage = async () => {
  if (isLoading.value || (!userInput.value.trim() && attachments.value.length === 0)) return;

  const content = userInput.value;
  const files = [...attachments.value];
  
  // Add to local history
  chatHistory.value.push({ 
    role: 'user', 
    content: content || (files.length > 0 ? `Uploaded ${files.length} file(s)` : '') 
  });
  
  userInput.value = '';
  attachments.value = [];
  isLoading.value = true;
  scrollToBottom();

  try {
    const apiMessages = chatHistory.value.map(m => ({ role: m.role, content: m.content }));
    const response = await sendChatMessage(apiMessages, files);
    
    if (response.choices && response.choices.length > 0) {
      chatHistory.value.push(response.choices[0].message);
    }
  } catch (err) {
    chatHistory.value.push({ role: 'assistant', content: `Maaf, terjadi kesalahan: ${err.message}` });
  } finally {
    isLoading.value = false;
    scrollToBottom();
  }
};

const parseToolArgs = (argsStr) => {
  try { return JSON.parse(argsStr); } catch (e) { return {}; }
};

const normalizeExpenseDate = (rawDate) => {
  if (!rawDate || typeof rawDate !== 'string') return rawDate;

  const dmySlash = rawDate.match(/^(\d{1,2})\/(\d{1,2})\/(\d{4})$/);
  if (dmySlash) {
    const [, dd, mm, yyyy] = dmySlash;
    return `${yyyy}-${mm.padStart(2, '0')}-${dd.padStart(2, '0')}T00:00:00Z`;
  }

  const dmyDash = rawDate.match(/^(\d{1,2})-(\d{1,2})-(\d{4})$/);
  if (dmyDash) {
    const [, dd, mm, yyyy] = dmyDash;
    return `${yyyy}-${mm.padStart(2, '0')}-${dd.padStart(2, '0')}T00:00:00Z`;
  }

  return rawDate;
};

const extractApiError = async (resp, fallback) => {
  try {
    const data = await resp.json();
    return data?.error || fallback;
  } catch (e) {
    return fallback;
  }
};

const cloneExpenses = (expenses) => JSON.parse(JSON.stringify(Array.isArray(expenses) ? expenses : []));

const getEditableExpenses = (callId, argsStr) => {
  if (!editablePreviewExpenses.value[callId]) {
    const { expenses } = parseToolArgs(argsStr);
    editablePreviewExpenses.value[callId] = cloneExpenses(expenses);
  }
  return editablePreviewExpenses.value[callId];
};

const addPreviewExpense = (callId, argsStr) => {
  const list = getEditableExpenses(callId, argsStr);
  list.push({ name: '', category: '', amount: 0, date: new Date().toISOString(), currency: 'IDR' });
};

const removePreviewExpense = (callId, index) => {
  const list = editablePreviewExpenses.value[callId] || [];
  list.splice(index, 1);
};

const setExpenseAmount = (callId, index, raw) => {
  const list = editablePreviewExpenses.value[callId];
  if (!list || !list[index]) return;
  const value = Number(raw);
  list[index].amount = Number.isFinite(value) ? value : 0;
};

const toInputDate = (value) => {
  if (!value) return '';
  const d = new Date(value);
  if (!Number.isNaN(d.getTime())) return d.toISOString().slice(0, 10);
  const normalized = normalizeExpenseDate(value);
  const nd = new Date(normalized);
  return Number.isNaN(nd.getTime()) ? '' : nd.toISOString().slice(0, 10);
};

const setExpenseDate = (callId, index, dateValue) => {
  const list = editablePreviewExpenses.value[callId];
  if (!list || !list[index]) return;
  list[index].date = dateValue ? `${dateValue}T00:00:00Z` : '';
};

const formatAmount = (amt) => {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(amt);
};

const formatDate = (dateStr) => {
  try { return new Date(dateStr).toLocaleDateString(); } catch (e) { return dateStr; }
};

const scrollToBottom = () => {
  nextTick(() => {
    if (messageContainer.value) {
      messageContainer.value.scrollTop = messageContainer.value.scrollHeight;
    }
  });
};

const confirmBatch = async (callId, argsStr) => {
  const fallbackExpenses = parseToolArgs(argsStr).expenses;
  const sourceExpenses = editablePreviewExpenses.value[callId] || fallbackExpenses;
  try {
    isLoading.value = true;
    const normalizedExpenses = Array.isArray(sourceExpenses)
      ? sourceExpenses.map((exp) => ({ ...exp, date: normalizeExpenseDate(exp?.date) }))
      : [];
    const resp = await apiFetch('/api/v1/expenses/batch', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(normalizedExpenses)
    });
    if (!resp.ok) throw new Error(await extractApiError(resp, 'Failed to save data'));
    
    chatHistory.value.push({ role: 'assistant', content: '✅ Berhasil menyimpan data transaksi!' });
  } catch (err) {
    chatHistory.value.push({ role: 'assistant', content: `❌ Error: ${err.message}` });
  } finally {
    isLoading.value = false;
    scrollToBottom();
  }
};

const confirmAction = async (tool, argsStr) => {
  const args = parseToolArgs(argsStr);
  let endpoint = '';
  if (tool === 'add_expense') endpoint = '/expense';
  if (tool === 'add_budget') endpoint = '/budget';

  try {
    isLoading.value = true;
    const resp = await apiFetch(endpoint, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(args)
    });
    if (!resp.ok) throw new Error('Action failed');
    chatHistory.value.push({ role: 'assistant', content: `✅ Berhasil melakukan aksi: ${tool}!` });
  } catch (err) {
    chatHistory.value.push({ role: 'assistant', content: `❌ Error: ${err.message}` });
  } finally {
    isLoading.value = false;
    scrollToBottom();
  }
};

const cancelAction = (idx) => {
  chatHistory.value.splice(idx, 1);
  chatHistory.value.push({ role: 'assistant', content: 'Aksi dibatalkan.' });
};
</script>

<style scoped>
.animate-bounce {
  animation: bounce 1s infinite;
}
@keyframes bounce {
  0%, 100% { transform: translateY(-25%); animation-timing-function: cubic-bezier(0.8,0,1,1); }
  50% { transform: none; animation-timing-function: cubic-bezier(0,0,0.2,1); }
}
</style>
