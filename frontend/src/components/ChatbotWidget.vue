<template>
  <div class="fixed bottom-4 right-4 z-50 flex flex-col items-end">
    <!-- Chat Window -->
    <div v-if="isOpen" class="mb-4 flex flex-col w-80 md:w-96 h-[500px] bg-white dark:bg-slate-800 rounded-lg shadow-2xl border border-slate-200 dark:border-slate-700 overflow-hidden">
      <!-- Header -->
      <div class="p-4 bg-indigo-600 text-white flex justify-between items-center shadow-md">
        <div class="flex items-center space-x-2">
          <i class="fa-solid fa-owl text-xl"></i>
          <span class="font-bold">ExpenseOwl AI</span>
        </div>
        <button @click="isOpen = false" class="hover:text-indigo-200 transition-colors">
          <i class="fa-solid fa-xmark"></i>
        </button>
      </div>

      <!-- Messages -->
      <div ref="messageContainer" class="flex-1 overflow-y-auto p-4 space-y-4">
        <div v-for="(msg, idx) in chatHistory" :key="idx" :class="['flex', msg.role === 'user' ? 'justify-end' : 'justify-start']">
          <div :class="['max-w-[85%] rounded-lg p-3 text-sm shadow-sm', msg.role === 'user' ? 'bg-indigo-100 dark:bg-indigo-900 text-indigo-900 dark:text-indigo-100' : 'bg-slate-100 dark:bg-slate-700 text-slate-800 dark:text-slate-200']">
            <!-- Tool Calls (e.g. Preview Parsed Data) -->
            <div v-if="msg.tool_calls">
              <div v-for="call in msg.tool_calls" :key="call.id">
                <div v-if="call.function.name === 'preview_parsed_data'" class="space-y-2 mt-2">
                  <p class="font-semibold text-xs uppercase tracking-wider text-slate-500">Preview Data yang Diolah:</p>
                  <div class="overflow-x-auto border border-slate-300 dark:border-slate-600 rounded">
                    <table class="min-w-full text-[10px] text-left">
                      <thead class="bg-slate-200 dark:bg-slate-600">
                        <tr>
                          <th class="px-2 py-1">Name</th>
                          <th class="px-2 py-1">Cat</th>
                          <th class="px-2 py-1">Amount</th>
                          <th class="px-2 py-1">Date</th>
                          <th class="px-2 py-1"></th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="(exp, eIdx) in getEditableExpenses(call.id, call.function.arguments)" :key="eIdx" class="border-t border-slate-200 dark:border-slate-700">
                          <td class="px-2 py-1">
                            <input v-model="exp.name" class="w-full min-w-[90px] bg-white/80 dark:bg-slate-900 border border-slate-300 dark:border-slate-600 rounded px-1 py-0.5 text-[10px]" />
                          </td>
                          <td class="px-2 py-1">
                            <input v-model="exp.category" class="w-full min-w-[80px] bg-white/80 dark:bg-slate-900 border border-slate-300 dark:border-slate-600 rounded px-1 py-0.5 text-[10px]" />
                          </td>
                          <td class="px-2 py-1">
                            <input :value="exp.amount" @input="setExpenseAmount(call.id, eIdx, $event.target.value)" class="w-full min-w-[90px] font-mono bg-white/80 dark:bg-slate-900 border border-slate-300 dark:border-slate-600 rounded px-1 py-0.5 text-[10px]" />
                          </td>
                          <td class="px-2 py-1">
                            <input :value="toInputDate(exp.date)" @input="setExpenseDate(call.id, eIdx, $event.target.value)" type="date" class="w-full min-w-[120px] bg-white/80 dark:bg-slate-900 border border-slate-300 dark:border-slate-600 rounded px-1 py-0.5 text-[10px]" />
                          </td>
                          <td class="px-2 py-1 text-right">
                            <button @click="removePreviewExpense(call.id, eIdx)" class="px-1.5 py-0.5 bg-rose-500 text-white rounded text-[10px] hover:bg-rose-600 transition-colors">x</button>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                  <div class="flex justify-between items-center">
                    <button @click="addPreviewExpense(call.id, call.function.arguments)" class="px-2 py-1 bg-slate-500 text-white rounded text-[10px] hover:bg-slate-600 transition-colors">+ Row</button>
                  </div>
                  <div class="flex justify-end space-x-2 pt-2">
                    <button @click="confirmBatch(call.id, call.function.arguments)" class="px-3 py-1 bg-green-600 text-white rounded text-xs hover:bg-green-700 transition-colors">Confirm & Save</button>
                    <button @click="cancelAction(idx)" class="px-3 py-1 bg-slate-400 text-white rounded text-xs hover:bg-slate-500 transition-colors">Cancel</button>
                  </div>
                </div>
                <!-- Other tools like add_expense, add_budget -->
                <div v-else-if="['add_expense', 'add_budget'].includes(call.function.name)" class="mt-2 p-2 bg-indigo-50 dark:bg-indigo-800 rounded border border-indigo-200 dark:border-indigo-700">
                   <p class="text-xs">AI ingin melakukan aksi: <strong>{{ call.function.name }}</strong></p>
                   <pre class="text-[10px] bg-white dark:bg-slate-900 p-1 mt-1 rounded">{{ call.function.arguments }}</pre>
                   <div class="flex justify-end space-x-2 pt-2">
                    <button @click="confirmAction(call.function.name, call.function.arguments)" class="px-3 py-1 bg-indigo-600 text-white rounded text-xs hover:bg-indigo-700 transition-colors">Confirm</button>
                    <button @click="cancelAction(idx)" class="px-3 py-1 bg-slate-400 text-white rounded text-xs hover:bg-slate-500 transition-colors">Cancel</button>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="whitespace-pre-wrap">{{ msg.content }}</div>
          </div>
        </div>
        <div v-if="isLoading" class="flex justify-start">
          <div class="bg-slate-100 dark:bg-slate-700 p-3 rounded-lg shadow-sm">
            <div class="flex space-x-1">
              <div class="w-2 h-2 bg-slate-400 rounded-full animate-bounce"></div>
              <div class="w-2 h-2 bg-slate-400 rounded-full animate-bounce delay-75"></div>
              <div class="w-2 h-2 bg-slate-400 rounded-full animate-bounce delay-150"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Attachment Preview -->
      <div v-if="attachments.length > 0" class="px-4 py-2 border-t border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-900 flex flex-wrap gap-2">
        <div v-for="(file, fIdx) in attachments" :key="fIdx" class="flex items-center bg-indigo-100 dark:bg-indigo-900 px-2 py-1 rounded text-[10px] text-indigo-700 dark:text-indigo-200">
          <span class="truncate max-w-[100px]">{{ file.name }}</span>
          <button @click="removeAttachment(fIdx)" class="ml-2 hover:text-indigo-500"><i class="fa-solid fa-circle-xmark"></i></button>
        </div>
      </div>

      <!-- Input -->
      <div class="p-4 border-t border-slate-200 dark:border-slate-700 flex items-end space-x-2">
        <label class="cursor-pointer p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded transition-colors text-slate-500">
          <i class="fa-solid fa-paperclip text-lg"></i>
          <input type="file" multiple class="hidden" @change="handleFileUpload" accept=".csv,image/*" />
        </label>
        <textarea v-model="userInput" @keydown.enter.exact.prevent="sendMessage" placeholder="Ask me anything..." rows="1" class="flex-1 bg-slate-100 dark:bg-slate-900 border-none rounded-lg p-2 text-sm focus:ring-2 focus:ring-indigo-500 resize-none max-h-32 transition-all" ref="inputField"></textarea>
        <button @click="sendMessage" :disabled="isLoading || (!userInput.trim() && attachments.length === 0)" class="p-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-all disabled:opacity-50 shadow-md">
          <i class="fa-solid fa-paper-plane"></i>
        </button>
      </div>
    </div>

    <!-- Toggle Button -->
    <button @click="toggleChat" class="w-14 h-14 bg-indigo-600 text-white rounded-full shadow-2xl flex items-center justify-center hover:bg-indigo-700 transition-all transform hover:scale-110 active:scale-95 border-4 border-white dark:border-slate-800">
      <i :class="isOpen ? 'fa-solid fa-minus text-xl' : 'fa-solid fa-owl text-2xl'"></i>
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
.delay-75 { animation-delay: 75ms; }
.delay-150 { animation-delay: 150ms; }
</style>
