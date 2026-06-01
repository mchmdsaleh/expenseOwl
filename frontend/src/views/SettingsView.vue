<template>
  <section class="space-y-10 animate-in fade-in duration-700">
    <!-- Category Settings -->
    <div class="glass-card p-6 md:p-8 rounded-3xl">
      <div class="flex items-center gap-3 mb-6">
         <div class="h-10 w-10 bg-indigo-500/20 text-indigo-500 rounded-xl flex items-center justify-center">
            <i class="fa-solid fa-tags text-xl"></i>
         </div>
         <h2 class="text-xl font-black tracking-tight text-[var(--text-primary)]">Categories</h2>
      </div>
      
      <div class="space-y-3">
        <div
          v-for="item in visibleCategories"
          :key="item.category"
          class="flex items-center justify-between gap-4 p-3 rounded-2xl bg-[var(--bg-elevated)] border border-[var(--border)] transition-all hover:border-[var(--accent)]/50 group"
        >
          <div class="flex items-center gap-4">
            <span class="text-[var(--text-secondary)] cursor-grab active:cursor-grabbing hover:text-[var(--text-primary)] transition-colors"><i class="fa-solid fa-grip-vertical"></i></span>
            <span class="font-bold text-sm text-[var(--text-primary)]">{{ item.category }}</span>
          </div>
          <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
            <button
              type="button"
              class="h-8 w-8 flex items-center justify-center rounded-lg hover:bg-[var(--bg-primary)] text-[var(--text-secondary)] disabled:opacity-20"
              @click="moveCategory(item.index, -1)"
              :disabled="item.index === 0"
            >
              <i class="fa-solid fa-chevron-up text-[10px]"></i>
            </button>
            <button
              type="button"
              class="h-8 w-8 flex items-center justify-center rounded-lg hover:bg-[var(--bg-primary)] text-[var(--text-secondary)] disabled:opacity-20"
              @click="moveCategory(item.index, 1)"
              :disabled="item.index === categories.length - 1"
            >
              <i class="fa-solid fa-chevron-down text-[10px]"></i>
            </button>
            <button type="button" class="h-8 w-8 flex items-center justify-center rounded-lg hover:bg-rose-500/10 text-rose-500" @click="removeCategory(item.index)">
              <i class="fa-solid fa-trash-can text-[10px]"></i>
            </button>
          </div>
        </div>
        
        <div v-if="canLoadMoreCategories" class="flex justify-center pt-2">
          <button type="button" class="text-xs font-bold text-[var(--accent)] hover:underline" @click="loadMoreCategories">Show all categories ({{ categories.length }})</button>
        </div>

        <div class="flex flex-col gap-4 mt-6 sm:flex-row">
          <input v-model="newCategory" type="text" placeholder="New category name..." class="input-modern flex-1" />
          <button type="button" class="btn-primary sm:w-auto px-8" @click="addCategory">Add</button>
        </div>

        <div class="flex flex-col gap-3 mt-4 sm:flex-row sm:items-center sm:justify-between border-t border-[var(--border)] pt-6">
          <button type="button" class="btn-primary bg-indigo-600 shadow-indigo-600/20 text-[var(--text-primary)]" @click="saveCategories">Save Changes</button>
          <transition name="fade">
            <div v-if="categoryMessage.text" :class="['rounded-xl px-4 py-2 text-sm font-bold', categoryMessage.type === 'success' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-rose-500/10 text-rose-400']">
              {{ categoryMessage.text }}
            </div>
          </transition>
        </div>
      </div>
    </div>

    <!-- AI Configuration -->
    <div class="grid gap-8 lg:grid-cols-2">
    <div class="glass-card p-6 md:p-8 rounded-3xl relative overflow-hidden">
        <div class="absolute top-0 right-0 p-4 opacity-5 pointer-events-none">
           <i class="fa-solid fa-microchip text-6xl text-[var(--accent)]"></i>
        </div>
        <div class="flex items-center gap-3 mb-6">
           <div class="h-10 w-10 bg-violet-500/20 text-violet-500 rounded-xl flex items-center justify-center">
              <i class="fa-solid fa-robot text-xl"></i>
           </div>
           <h2 class="text-xl font-black tracking-tight text-[var(--text-primary)]">AI Model</h2>
        </div>

        <div class="grid gap-6">
          <div class="space-y-2">
            <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1">Provider</label>
            <select v-model="aiConfig.provider" class="input-modern w-full appearance-none text-[var(--text-primary)]">
              <option value="openai">OpenAI</option>
              <option value="anthropic">Anthropic (Shim)</option>
              <option value="google">Google Gemini (Shim)</option>
              <option value="deepseek">DeepSeek</option>
              <option value="kimi">Kimi</option>
              <option value="qwen">Qwen / Alibaba</option>
              <option value="custom">Custom (OpenAI Compatible)</option>
            </select>
          </div>
          <div class="space-y-2">
            <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1">Model Name</label>
            <input v-model="aiConfig.model" type="text" class="input-modern w-full font-mono text-xs text-[var(--text-primary)]" placeholder="gpt-4o, deepseek-chat..." />
          </div>
          <div class="space-y-2">
            <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1">Base URL (Optional)</label>
            <input v-model="aiConfig.baseUrl" type="text" class="input-modern w-full font-mono text-xs text-[var(--text-primary)]" placeholder="https://api.openai.com/v1" />
          </div>
          <div class="space-y-2">
            <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1">API Key</label>
            <input v-model="aiConfig.apiKey" type="password" class="input-modern w-full text-[var(--text-primary)]" placeholder="sk-..." />
          </div>
          <div class="pt-2">
            <button type="button" class="btn-primary w-full bg-violet-600 shadow-violet-600/20 text-[var(--text-primary)]" @click="saveAIConfig">Save AI Config</button>
            <transition name="fade">
              <div v-if="aiConfigMessage.text" :class="['mt-4 rounded-xl px-4 py-2 text-center text-xs font-bold', aiConfigMessage.type === 'success' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-rose-500/10 text-rose-400']">
                {{ aiConfigMessage.text }}
              </div>
            </transition>
          </div>
        </div>
      </div>

      <div class="glass-card p-6 md:p-8 rounded-3xl flex flex-col h-full">
        <div class="flex items-center gap-3 mb-4">
           <div class="h-10 w-10 bg-indigo-500/20 text-indigo-500 rounded-xl flex items-center justify-center">
              <i class="fa-solid fa-brain text-xl"></i>
           </div>
           <h2 class="text-xl font-black tracking-tight text-[var(--text-primary)]">AI Context</h2>
        </div>
        <p class="text-xs text-[var(--text-secondary)] leading-relaxed mb-6 font-medium">
          Define custom rules for parsing bank mutations or receipt images. (e.g., "Always rename FamilyMart to Kopi")
        </p>
        <textarea
          v-model="aiContext"
          class="input-modern w-full flex-1 min-h-[250px] font-mono text-[10px] leading-relaxed resize-none p-4 text-[var(--text-primary)]"
          placeholder="Enter custom AI instructions here..."
        ></textarea>
        <div class="pt-6">
          <button type="button" class="btn-primary w-full text-[var(--text-primary)]" @click="saveAIContext">Save Context Rules</button>
          <transition name="fade">
            <div v-if="aiContextMessage.text" :class="['mt-4 rounded-xl px-4 py-2 text-center text-xs font-bold', aiContextMessage.type === 'success' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-rose-500/10 text-rose-400']">
              {{ aiContextMessage.text }}
            </div>
          </transition>
        </div>
      </div>
    </div>

    <!-- Budget Settings -->
    <div class="glass-card p-6 md:p-8 rounded-3xl">
      <div class="flex items-center gap-3 mb-8">
         <div class="h-10 w-10 bg-emerald-500/20 text-emerald-500 rounded-xl flex items-center justify-center">
            <i class="fa-solid fa-sack-dollar text-xl"></i>
         </div>
         <h2 class="text-xl font-black tracking-tight text-[var(--text-primary)]">Budget Limits</h2>
      </div>

      <form class="grid gap-6 md:grid-cols-2 bg-[var(--bg-elevated)] p-6 rounded-2xl border border-[var(--border)]" @submit.prevent="submitBudget">
        <div class="space-y-2">
          <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1" for="budgetCategory">Category</label>
          <select id="budgetCategory" v-model="budgetForm.category" class="input-modern w-full appearance-none text-[var(--text-primary)]" required>
            <option value="" disabled>Select category</option>
            <option v-for="category in budgetCategoryOptions" :key="category" :value="category">
              {{ category }}
            </option>
          </select>
        </div>
        <div class="space-y-2">
          <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1" for="budgetAmount">Monthly Limit</label>
          <div class="relative">
             <span class="absolute left-4 top-1/2 -translate-y-1/2 text-[var(--text-secondary)] font-bold text-sm">IDR</span>
             <input
               id="budgetAmount"
               :value="formattedBudgetAmount"
               inputmode="decimal"
               class="input-modern w-full pl-12 text-lg font-black tabular-nums text-[var(--text-primary)]"
               required
               @input="handleBudgetAmountInput"
               @blur="normalizeBudgetAmount"
             />
          </div>
        </div>
        <div class="md:col-span-2 flex items-center justify-between gap-4 pt-2">
          <div class="flex gap-2">
            <button type="submit" class="btn-primary min-w-[140px] text-[var(--text-primary)]">{{ budgetForm.submitLabel }}</button>
            <button v-if="budgetForm.id" type="button" class="px-6 py-2.5 rounded-2xl border border-[var(--border)] text-sm font-bold text-[var(--text-primary)] hover:bg-[var(--bg-primary)] transition-all" @click="resetBudgetForm">Cancel</button>
          </div>
          <transition name="fade">
            <div v-if="budgetMessage.text" :class="['rounded-xl px-4 py-2 text-sm font-bold', budgetMessage.type === 'success' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-rose-500/10 text-rose-400']">
              {{ budgetMessage.text }}
            </div>
          </transition>
        </div>
      </form>

      <div class="mt-10 space-y-6">
        <div v-if="budgetList.length" class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between px-2 text-[var(--text-primary)]">
          <div class="flex flex-col">
            <h3 class="text-sm font-black uppercase tracking-widest text-[var(--text-secondary)]">Active Overrides</h3>
            <p class="text-[10px] font-medium text-[var(--text-secondary)]">Adjusting limits for {{ budgetMonthLabel }}</p>
          </div>
          <div class="flex items-center gap-3">
            <input id="budgetMonth" type="month" v-model="budgetMonth" class="input-modern text-xs font-bold h-10 text-[var(--text-primary)]" @change="handleBudgetMonthChange" />
          </div>
        </div>

        <div v-if="budgetList.length === 0" class="py-12 text-center border-2 border-dashed border-[var(--border)] rounded-3xl italic text-[var(--text-secondary)] text-sm font-medium">
          No budgets configured yet.
        </div>
        
        <div v-else class="grid gap-6">
          <div
            v-for="budget in budgetList"
            :key="budget.id"
            class="p-6 rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)] shadow-sm hover:shadow-xl transition-all"
          >
            <div class="flex items-start justify-between mb-6">
              <div class="space-y-1">
                <div class="text-lg font-black tracking-tight text-[var(--text-primary)]">{{ budget.category }}</div>
                <div class="flex items-center gap-2 text-[10px] font-bold uppercase tracking-wider text-[var(--text-secondary)]">
                   Base: <span class="text-[var(--text-primary)] font-black tabular-nums">{{ formatCurrency(budget.amount) }}</span>
                </div>
              </div>
              <button type="button" class="h-10 w-10 flex items-center justify-center rounded-xl bg-rose-500/10 text-rose-500 hover:bg-rose-500 hover:text-white transition-all active:scale-90 shadow-sm" @click="openDeleteBudget(budget)">
                <i class="fa-solid fa-trash-can text-sm"></i>
              </button>
            </div>

            <div class="grid gap-6 md:grid-cols-2">
               <!-- Override -->
               <div class="space-y-3 bg-[var(--bg-elevated)] p-4 rounded-2xl relative overflow-hidden">
                  <div class="text-[9px] font-black uppercase tracking-widest text-[var(--text-secondary)]">Month Override</div>
                  <div class="flex gap-2">
                     <input v-model="budgetForms[budget.id].override" class="input-modern flex-1 h-10 text-xs font-bold bg-[var(--bg-primary)] text-[var(--text-primary)]" placeholder="New limit..." />
                     <button @click="saveBudgetOverride(budget)" class="h-10 px-4 rounded-xl bg-[var(--accent)] text-white text-[10px] font-black uppercase transition-all hover:brightness-110 active:scale-95 shadow-lg shadow-indigo-500/20">Save</button>
                  </div>
                  <div v-if="budgetSummaryMap[budget.id]?.overrideId" class="flex items-center justify-between mt-2">
                     <span class="text-[10px] font-bold text-amber-400">Active: {{ formatCurrency(budgetSummaryMap[budget.id].overrideAmount || 0) }}</span>
                     <button @click="clearBudgetOverride(budget)" class="text-[9px] font-black text-rose-400 hover:underline">Reset</button>
                  </div>
               </div>
               <!-- Adjustment -->
               <div class="space-y-3 bg-[var(--bg-elevated)] p-4 rounded-2xl">
                  <div class="text-[9px] font-black uppercase tracking-widest text-[var(--text-secondary)]">Adjustment (±)</div>
                  <div class="flex gap-2">
                     <input v-model="budgetForms[budget.id].adjustment" class="input-modern flex-1 h-10 text-xs font-bold bg-[var(--bg-primary)] text-[var(--text-primary)]" placeholder="e.g. +50, -20..." />
                     <button @click="saveBudgetAdjustment(budget)" class="h-10 px-4 rounded-xl bg-[var(--accent)] text-white text-[10px] font-black uppercase transition-all hover:brightness-110 active:scale-95 shadow-lg shadow-indigo-500/20">Apply</button>
                  </div>
                  <div v-if="budgetSummaryMap[budget.id]?.adjustmentId" class="flex items-center justify-between mt-2">
                     <span class="text-[10px] font-bold text-emerald-400">Active: {{ budgetSummaryMap[budget.id].adjustmentAmount > 0 ? '+' : '' }}{{ formatCurrency(budgetSummaryMap[budget.id].adjustmentAmount || 0) }}</span>
                     <button @click="clearBudgetAdjustment(budget)" class="text-[9px] font-black text-rose-400 hover:underline">Reset</button>
                  </div>
               </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Layout Configs -->
    <div class="grid gap-8 md:grid-cols-2 lg:grid-cols-3">
      <!-- Currency -->
      <div class="glass-card p-6 rounded-3xl">
        <div class="flex items-center gap-3 mb-6">
           <div class="h-8 w-8 bg-amber-500/20 text-amber-500 rounded-lg flex items-center justify-center">
              <i class="fa-solid fa-coins text-sm"></i>
           </div>
           <h2 class="text-base font-black tracking-tight text-[var(--text-primary)]">Currency</h2>
        </div>
        <div class="flex gap-2">
          <select v-model="currencyCode" class="input-modern flex-1 h-11 text-sm appearance-none pr-8 text-[var(--text-primary)]">
            <option v-for="code in currencyOptions" :key="code" :value="code">
              {{ code.toUpperCase() }} ({{ currencyBehaviors[code].symbol }})
            </option>
          </select>
          <button type="button" class="btn-primary h-11 px-4 text-[var(--text-primary)]" @click="saveCurrency">Set</button>
        </div>
        <transition name="fade">
          <div v-if="currencyMessage.text" class="mt-4 text-xs font-bold text-center text-emerald-400">{{ currencyMessage.text }}</div>
        </transition>
      </div>

      <!-- Start Date -->
      <div class="glass-card p-6 rounded-3xl">
        <div class="flex items-center gap-3 mb-6">
           <div class="h-8 w-8 bg-sky-500/20 text-sky-500 rounded-lg flex items-center justify-center">
              <i class="fa-solid fa-calendar-day text-sm"></i>
           </div>
           <h2 class="text-base font-black tracking-tight text-[var(--text-primary)]">Cycle Start</h2>
        </div>
        <div class="flex gap-2 mb-4">
          <input v-model.number="startDate" type="number" min="1" max="31" class="input-modern flex-1 h-11 text-[var(--text-primary)]" :disabled="endOfMonth" />
          <button type="button" class="btn-primary h-11 px-4 text-[var(--text-primary)]" @click="saveStartDate" :disabled="endOfMonth">Set</button>
        </div>
        <label class="flex items-center gap-3 cursor-pointer group p-2 rounded-xl hover:bg-[var(--bg-elevated)] transition-all">
          <input type="checkbox" class="sr-only peer" v-model="endOfMonth" />
          <div class="h-5 w-9 bg-slate-700 rounded-full peer peer-checked:bg-[var(--accent)] transition-colors relative after:content-[''] after:absolute after:top-1 after:left-1 after:h-3 after:w-3 after:bg-white after:rounded-full after:transition-transform peer-checked:after:translate-x-4"></div>
          <span class="text-xs font-bold text-[var(--text-secondary)] group-hover:text-[var(--text-primary)] transition-colors">End of Month Alignment</span>
        </label>
        <div class="mt-4 text-right" v-if="endOfMonth !== state.endOfMonth">
           <button @click="saveEndOfMonth" class="text-[10px] font-black text-[var(--accent)] uppercase tracking-widest hover:underline">Confirm Alignment Change</button>
        </div>
      </div>

      <!-- Theme -->
      <div class="glass-card p-6 rounded-3xl">
        <div class="flex items-center gap-3 mb-6">
           <div class="h-8 w-8 bg-indigo-500/20 text-indigo-500 rounded-lg flex items-center justify-center">
              <i class="fa-solid fa-palette text-sm"></i>
           </div>
           <h2 class="text-base font-black tracking-tight text-[var(--text-primary)]">Theme</h2>
        </div>
        <div class="relative group">
           <select v-model="theme" @change="applyTheme" class="input-modern w-full h-11 appearance-none pr-10 text-sm text-[var(--text-primary)]">
             <option value="system">System Default</option>
             <option value="light">Modern Light</option>
             <option value="dark">Modern Dark</option>
           </select>
           <i class="fa-solid fa-angle-down absolute right-4 top-1/2 -translate-y-1/2 text-[var(--text-secondary)] pointer-events-none group-hover:text-[var(--accent)] transition-colors"></i>
        </div>
        <transition name="fade">
          <div v-if="themeMessage.text" class="mt-4 text-xs font-bold text-center text-indigo-400">{{ themeMessage.text }}</div>
        </transition>
      </div>
    </div>

    <!-- Recurring Transactions -->
    <div class="glass-card p-6 md:p-8 rounded-3xl" ref="recurringCardRef">
      <div class="flex items-center gap-3 mb-8">
         <div class="h-10 w-10 bg-indigo-500/20 text-indigo-500 rounded-xl flex items-center justify-center">
            <i class="fa-solid fa-arrows-spin text-xl"></i>
         </div>
         <h2 class="text-xl font-black tracking-tight text-[var(--text-primary)]">Recurring Transactions</h2>
      </div>

      <form class="grid gap-6 md:grid-cols-3 bg-[var(--bg-elevated)] p-8 rounded-3xl border border-[var(--border)]" @submit.prevent="submitRecurring">
        <div class="space-y-2">
          <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1">Name</label>
          <input v-model="recurringForm.name" type="text" class="input-modern w-full text-[var(--text-primary)]" placeholder="Netflix, Rent, etc..." required />
        </div>
        <div class="space-y-2">
          <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1">Amount</label>
          <input
            :value="formattedRecurringAmount"
            inputmode="decimal"
            class="input-modern w-full font-black text-lg tabular-nums text-[var(--text-primary)]"
            required
            @input="handleRecurringAmountInput"
            @blur="normalizeRecurringAmount"
          />
        </div>
        <div class="space-y-2">
          <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1">Category</label>
          <select v-model="recurringForm.category" class="input-modern w-full appearance-none text-[var(--text-primary)]" required>
            <option value="" disabled>Select category</option>
            <option v-for="category in state.categories" :key="category" :value="category">{{ category }}</option>
          </select>
        </div>
        <div class="space-y-2">
          <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1">Interval</label>
          <select v-model="recurringForm.interval" class="input-modern w-full appearance-none text-[var(--text-primary)]" required>
            <option value="daily">Daily</option>
            <option value="weekly">Weekly</option>
            <option value="monthly">Monthly</option>
            <option value="yearly">Yearly</option>
          </select>
        </div>
        <div class="space-y-2">
          <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1">Start Date</label>
          <input v-model="recurringForm.startDate" type="date" class="input-modern w-full text-[var(--text-primary)]" required />
        </div>
        <div class="space-y-2">
          <label class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)] ml-1">Limit Occurrences</label>
          <input v-model.number="recurringForm.occurrences" type="number" min="0" class="input-modern w-full text-[var(--text-primary)]" />
        </div>
        <div class="md:col-span-3 flex flex-wrap items-center justify-between gap-4 pt-4 border-t border-[var(--border)]">
          <label class="flex items-center gap-3 cursor-pointer group">
            <input v-model="recurringForm.reportGain" type="checkbox" class="sr-only peer" />
            <div class="h-5 w-9 bg-slate-700 rounded-full peer peer-checked:bg-emerald-500 transition-colors relative after:content-[''] after:absolute after:top-1 after:left-1 after:h-3 after:w-3 after:bg-white after:rounded-full after:transition-transform peer-checked:after:translate-x-4"></div>
            <span class="text-xs font-bold text-[var(--text-secondary)] group-hover:text-[var(--text-primary)] transition-colors uppercase tracking-widest">Income / Gain</span>
          </label>
          <div class="flex gap-2">
             <button v-if="editingRecurringId" type="button" @click="resetRecurringForm" class="btn-secondary px-8 py-3 rounded-2xl border border-[var(--border)] text-sm font-bold text-[var(--text-primary)]">Cancel</button>
             <button type="submit" class="btn-primary px-12 py-3 min-w-[200px] text-base text-[var(--text-primary)]">{{ recurringForm.submitLabel }}</button>
          </div>
        </div>
      </form>
      
      <transition name="fade">
        <div v-if="recurringMessage.text" :class="['mt-6 p-4 rounded-2xl text-center text-sm font-bold', recurringMessage.type === 'success' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-rose-500/10 text-rose-400']">
          {{ recurringMessage.text }}
        </div>
      </transition>

      <div class="mt-12">
        <h3 class="text-sm font-black uppercase tracking-widest text-[var(--text-secondary)] mb-6 text-center">Active Automations</h3>
        <div v-if="state.recurringExpenses.length === 0" class="py-10 text-center glass-card rounded-3xl border-dashed italic text-[var(--text-secondary)] text-sm">
           No recurring tasks found.
        </div>
        <div v-else class="grid gap-4">
          <div
            v-for="expense in state.recurringExpenses"
            :key="expense.id"
            class="flex flex-col gap-4 p-5 rounded-2xl bg-[var(--bg-elevated)] border border-[var(--border)] sm:flex-row sm:items-center sm:justify-between group hover:border-[var(--accent)]/40 transition-all shadow-sm"
          >
            <div class="flex items-center gap-4">
               <div class="h-10 w-10 rounded-xl bg-[var(--bg-primary)] flex items-center justify-center text-[var(--accent)] group-hover:scale-110 transition-transform">
                  <i class="fa-solid fa-repeat"></i>
               </div>
               <div class="space-y-0.5">
                  <div class="font-black text-[var(--text-primary)] tracking-tight">{{ expense.name }}</div>
                  <div class="flex items-center gap-2 text-[10px] font-bold text-[var(--text-secondary)] uppercase">
                    <span>{{ formatCurrency(expense.amount) }}</span>
                    <span>•</span>
                    <span class="text-[var(--accent)]">{{ expense.interval }}</span>
                    <span>•</span>
                    <span>Started {{ formatDate(expense.startDate) }}</span>
                  </div>
               </div>
            </div>
            <div class="flex items-center gap-2">
              <button type="button" class="h-9 w-9 flex items-center justify-center rounded-lg hover:bg-[var(--bg-primary)] text-[var(--text-secondary)] hover:text-indigo-400 transition-all" @click="editRecurring(expense)">
                <i class="fa-solid fa-pen-to-square"></i>
              </button>
              <button type="button" class="h-9 w-9 flex items-center justify-center rounded-lg hover:bg-rose-500/10 text-rose-500 transition-all" @click="openDeleteRecurring(expense)">
                <i class="fa-solid fa-trash-can"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Import/Export -->
    <div class="glass-card p-6 md:p-8 rounded-3xl">
      <div class="flex items-center gap-3 mb-8">
         <div class="h-10 w-10 bg-sky-500/20 text-sky-500 rounded-xl flex items-center justify-center">
            <i class="fa-solid fa-file-export text-xl"></i>
         </div>
         <h2 class="text-xl font-black tracking-tight text-[var(--text-primary)]">Data Transfer</h2>
      </div>

      <div class="grid gap-8 md:grid-cols-2">
         <!-- Export -->
         <div class="space-y-6">
            <h3 class="text-xs font-black uppercase tracking-widest text-[var(--text-secondary)] px-2">Export Data</h3>
            <div class="space-y-4 bg-[var(--bg-elevated)] p-6 rounded-3xl border border-[var(--border)]">
               <label class="flex items-center gap-3 cursor-pointer group">
                  <input v-model="exportAll" type="checkbox" class="sr-only peer" />
                  <div class="h-5 w-9 bg-slate-700 rounded-full peer peer-checked:bg-[var(--accent)] transition-colors relative after:content-[''] after:absolute after:top-1 after:left-1 after:h-3 after:w-3 after:bg-white after:rounded-full after:transition-transform peer-checked:after:translate-x-4"></div>
                  <span class="text-xs font-bold text-[var(--text-secondary)] group-hover:text-[var(--text-primary)] transition-colors">Everything (Lifetime)</span>
               </label>

               <div v-if="!exportAll" class="space-y-4 animate-in fade-in duration-300">
                  <select v-model="exportDateFilter" class="input-modern w-full h-11 text-xs font-bold text-[var(--text-primary)]">
                    <option value="month">This Month</option>
                    <option value="week">This Week</option>
                    <option value="today">Today</option>
                    <option value="range">Custom Range</option>
                  </select>
                  <div v-if="exportDateFilter === 'range'" class="grid grid-cols-2 gap-2">
                     <input v-model="exportRangeStart" type="date" class="input-modern text-[10px] font-bold text-[var(--text-primary)]" />
                     <input v-model="exportRangeEnd" type="date" class="input-modern text-[10px] font-bold text-[var(--text-primary)]" />
                  </div>
               </div>
               
               <button @click="exportCsv" :disabled="exportingCsv" class="btn-primary w-full h-12 text-sm shadow-indigo-500/10 text-[var(--text-primary)]">
                  {{ exportingCsv ? 'Preparing...' : 'Generate CSV Download' }}
               </button>
            </div>
         </div>

         <!-- Import -->
         <div class="space-y-6">
            <h3 class="text-xs font-black uppercase tracking-widest text-[var(--text-secondary)] px-2">Import Data</h3>
            <div class="flex flex-col items-center justify-center gap-4 bg-[var(--bg-elevated)] p-6 rounded-3xl border border-dashed border-[var(--border)] group hover:border-[var(--accent)] transition-colors h-full min-h-[200px]">
               <i class="fa-solid fa-cloud-arrow-up text-4xl text-[var(--text-secondary)] group-hover:text-[var(--accent)] transition-colors"></i>
               <label for="csv-import-file" class="cursor-pointer text-center">
                  <span class="text-sm font-bold text-[var(--text-primary)] block mb-1">Click to upload CSV</span>
                  <span class="text-[10px] font-medium text-[var(--text-secondary)]">Standard ExpenseOwl Format</span>
               </label>
               <input id="csv-import-file" type="file" accept=".csv" hidden @change="(event) => handleImport(event, '/import/csv')" />
            </div>
         </div>
      </div>

      <transition name="fade">
        <div v-if="importSummary" class="mt-8 p-6 rounded-3xl bg-[var(--bg-elevated)] border border-[var(--border)] animate-in slide-in-from-bottom duration-500">
          <div class="flex items-center gap-3 mb-4">
             <i class="fa-solid fa-circle-check text-emerald-500"></i>
             <h4 class="font-black text-sm tracking-tight text-[var(--text-primary)]">Import Successful</h4>
          </div>
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 text-center text-[var(--text-primary)]">
             <div class="p-3 rounded-2xl bg-[var(--bg-primary)]">
                <div class="text-[10px] font-bold text-[var(--text-secondary)] uppercase mb-1">Total</div>
                <div class="text-lg font-black">{{ importSummary.totalProcessed }}</div>
             </div>
             <div class="p-3 rounded-2xl bg-[var(--bg-primary)]">
                <div class="text-[10px] font-bold text-[var(--text-secondary)] uppercase mb-1 text-emerald-400">Saved</div>
                <div class="text-lg font-black text-emerald-400">{{ importSummary.imported }}</div>
             </div>
             <div class="p-3 rounded-2xl bg-[var(--bg-primary)]">
                <div class="text-[10px] font-bold text-[var(--text-secondary)] uppercase mb-1 text-rose-400">Skipped</div>
                <div class="text-lg font-black text-rose-400">{{ importSummary.skipped }}</div>
             </div>
             <div class="p-3 rounded-2xl bg-[var(--bg-primary)]">
                <div class="text-[10px] font-bold text-[var(--text-secondary)] uppercase mb-1">New Cats</div>
                <div class="text-lg font-black truncate px-1" :title="importSummary.newCategories">{{ importSummary.newCategories }}</div>
             </div>
          </div>
        </div>
      </transition>

      <transition name="fade">
        <div v-if="importMessage.text || exportMessage.text" class="mt-6 p-4 rounded-2xl text-center text-xs font-bold" :class="[ (importMessage.type === 'success' || exportMessage.type === 'success') ? 'bg-emerald-500/10 text-emerald-400' : 'bg-rose-500/10 text-rose-400']">
          {{ importMessage.text || exportMessage.text }}
        </div>
      </transition>
    </div>
  </section>

  <!-- Modals -->
  <transition name="fade">
    <div v-if="showDeleteBudget" class="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-sm px-4" @click.self="closeDeleteBudget">
      <div class="w-full max-w-sm glass-card rounded-3xl p-8 animate-in zoom-in duration-300">
        <h3 class="text-xl font-black text-center tracking-tight text-[var(--text-primary)]">Delete Budget?</h3>
        <p class="mt-3 text-sm text-[var(--text-secondary)] text-center leading-relaxed font-medium">
          Remove budget for <span class="text-[var(--text-primary)] font-bold">{{ budgetToDelete?.category }}</span>? This clears all monthly settings.
        </p>
        <div class="mt-8 grid grid-cols-2 gap-4">
          <button class="px-6 py-3 rounded-2xl bg-[var(--bg-elevated)] font-bold text-sm text-[var(--text-primary)]" @click="closeDeleteBudget">Keep it</button>
          <button class="px-6 py-3 rounded-2xl bg-rose-600 text-white font-bold text-sm" @click="confirmDeleteBudget">Delete</button>
        </div>
      </div>
    </div>
  </transition>

  <transition name="fade">
    <div v-if="showDeleteRecurring" class="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-sm px-4" @click.self="closeDeleteRecurring">
      <div class="w-full max-w-sm glass-card rounded-3xl p-8 animate-in zoom-in duration-300">
        <h3 class="text-xl font-black text-center tracking-tight text-[var(--text-primary)]">Stop Automation?</h3>
        <p class="mt-3 text-sm text-[var(--text-secondary)] text-center leading-relaxed font-medium">
          Future occurrences for this transaction will be deleted.
        </p>
        <div class="mt-8 grid grid-cols-2 gap-4">
          <button class="px-6 py-3 rounded-2xl bg-[var(--bg-elevated)] font-bold text-sm text-[var(--text-primary)]" @click="closeDeleteRecurring">Nevermind</button>
          <button class="px-6 py-3 rounded-2xl bg-rose-600 text-white font-bold text-sm" @click="confirmDeleteRecurring">Stop it</button>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted, nextTick } from 'vue';
import TagInput from '../components/TagInput.vue';
import state, { loadInitialData, refreshExpenses, refreshRecurringExpenses, refreshBudgets, refreshBudgetSummaries } from '../stores/appState';
import { apiFetch, getAIContext, updateAIContext, getAIConfig, updateAIConfig } from '../lib/api';
import { encryptPayload } from '../lib/encryption';
import { currencyBehaviors, formatCurrency as formatCurrencyRaw, getISODateWithLocalTime, formatMonth as formatMonthLabel } from '../lib/utils';

const primaryButtonClass = 'btn-primary';

const categories = ref([]);
const newCategory = ref('');
const categoryMessage = ref({ text: '', type: '' });
const categoryDisplayCount = ref(5);

const currencyCode = ref(state.currency);
const currencyMessage = ref({ text: '', type: '' });

const startDate = ref(state.startDate);
const startDateMessage = ref({ text: '', type: '' });
const endOfMonth = ref(state.endOfMonth || false);
const endOfMonthMessage = ref({ text: '', type: '' });

const theme = ref(localStorage.getItem('theme') || 'system');
const themeMessage = ref({ text: '', type: '' });

const aiContext = ref('');
const aiContextMessage = ref({ text: '', type: '' });

const aiConfig = ref({
  provider: 'openai',
  model: 'gpt-4o',
  baseUrl: '',
  apiKey: '',
});
const aiConfigMessage = ref({ text: '', type: '' });

const importMessage = ref({ text: '', type: '' });
const exportMessage = ref({ text: '', type: '' });
const exportingCsv = ref(false);
const exportDateFilter = ref('month');
const exportRangeStart = ref('');
const exportRangeEnd = ref('');
const exportAll = ref(false);
const importSummary = ref(null);
const csvImportRef = ref(null);

const budgetForm = ref(createBudgetForm());
const budgetMessage = ref({ text: '', type: '' });
const rawBudgetAmount = ref('');
const budgetMonth = ref(formatBudgetMonthKey(new Date()));
const budgetForms = reactive({});

const recurringForm = ref(createRecurringForm());
const recurringMessage = ref({ text: '', type: '' });
const editingRecurringId = ref(null);
const showDeleteRecurring = ref(false);
const expenseToDelete = ref(null);
const rawRecurringAmount = ref('');
const showDeleteBudget = ref(false);
const budgetToDelete = ref(null);

const visibleCategories = computed(() =>
  categories.value.map((category, index) => ({ category, index })).slice(0, categoryDisplayCount.value)
);

const canLoadMoreCategories = computed(
  () => categoryDisplayCount.value < categories.value.length
);

const allTags = computed(() => {
  const combined = new Set([...(state.tags || [])]);
  state.recurringExpenses.forEach((expense) => {
    (expense.tags || []).forEach((tag) => combined.add(tag));
  });
  return Array.from(combined);
});

const exportRangeValidationMessage = computed(() => {
  if (exportAll.value || exportDateFilter.value !== 'range') {
    return '';
  }
  if (!exportRangeStart.value || !exportRangeEnd.value) {
    return 'Select both start and end dates';
  }
  if (new Date(exportRangeStart.value) > new Date(exportRangeEnd.value)) {
    return 'Start date must be before end date';
  }
  return '';
});

const budgetList = computed(() => {
  if (!Array.isArray(state.budgets)) return [];
  return [...state.budgets].sort((a, b) => a.category.localeCompare(b.category));
});

const budgetCategoryOptions = computed(() => [...state.categories].sort((a, b) => a.localeCompare(b)));
const budgetSummaryMap = computed(() => {
  const map = {};
  const summaries = Array.isArray(state.budgetSummaries) ? state.budgetSummaries : [];
  summaries.forEach((summary) => {
    map[summary.id] = summary;
  });
  return map;
});
const budgetMonthLabel = computed(() => {
  try {
    const date = parseBudgetMonthKey(budgetMonth.value);
    return formatMonthLabel(date);
  } catch (error) {
    return budgetMonth.value;
  }
});

const recurringCardRef = ref(null);
const formattedRecurringAmount = computed(() => {
  if (!rawRecurringAmount.value) return '';
  const numeric = Number(rawRecurringAmount.value.replace(/[^0-9.-]/g, '')) || 0;
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(numeric);
});

const formattedBudgetAmount = computed(() => {
  if (!rawBudgetAmount.value) return '';
  const numeric = Number(rawBudgetAmount.value.replace(/[^0-9.-]/g, '')) || 0;
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(numeric);
});

watch(
  () => state.categories,
  (value) => {
    categories.value = [...value];
    if (!currencyOptions.includes(currencyCode.value)) {
      currencyCode.value = state.currency;
    }
    ensureCategoryDisplayBounds();
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

watch(
  () => state.endOfMonth,
  (value, oldValue) => {
    endOfMonth.value = value;
    if (oldValue !== undefined && value !== oldValue) {
      loadBudgetSummaries();
    }
  },
  { immediate: true }
);

watch(
  () => state.budgets,
  (budgets) => {
    ensureBudgetFormEntries(Array.isArray(budgets) ? budgets : []);
  },
  { immediate: true }
);

watch(
  () => state.budgetSummaries,
  () => {
    syncBudgetFormsFromSummaries();
  }
);

watch(budgetMonth, (value, oldValue) => {
  if (value !== oldValue) {
    loadBudgetSummaries();
  }
});

onMounted(async () => {
  await loadInitialData();
  await loadBudgetSummaries();
  try {
    const [context, config] = await Promise.all([
      getAIContext(),
      getAIConfig()
    ]);
    aiContext.value = context;
    if (config) {
      aiConfig.value = { ...aiConfig.value, ...config };
    }
  } catch (err) {
    console.error('Failed to load AI settings', err);
  }
});

async function saveAIContext() {
  try {
    await updateAIContext(aiContext.value);
    aiContextMessage.value = { text: 'AI context saved successfully.', type: 'success' };
    dismissAfter(() => (aiContextMessage.value = { text: '', type: '' }));
  } catch (err) {
    aiContextMessage.value = { text: err.message || 'Failed to save AI context', type: 'error' };
    dismissAfter(() => (aiContextMessage.value = { text: '', type: '' }));
  }
}

async function saveAIConfig() {
  try {
    await updateAIConfig(aiConfig.value);
    aiConfigMessage.value = { text: 'AI config saved successfully.', type: 'success' };
    dismissAfter(() => (aiConfigMessage.value = { text: '', type: '' }));
  } catch (err) {
    aiConfigMessage.value = { text: err.message || 'Failed to save AI config', type: 'error' };
    dismissAfter(() => (aiConfigMessage.value = { text: '', type: '' }));
  }
}

function createBudgetForm() {
  return {
    id: null,
    category: '',
    submitLabel: 'Add Budget',
  };
}

function formatBudgetMonthKey(date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  return `${year}-${month}`;
}

function parseBudgetMonthKey(key) {
  if (typeof key !== 'string' || key.length < 7) {
    throw new Error('Invalid month value');
  }
  const [yearRaw, monthRaw] = key.split('-');
  const year = Number(yearRaw);
  const month = Number(monthRaw);
  if (!Number.isFinite(year) || !Number.isFinite(month) || month < 1 || month > 12) {
    throw new Error('Invalid month value');
  }
  return new Date(year, month - 1, 1);
}

async function loadBudgetSummaries() {
  try {
    const target = parseBudgetMonthKey(budgetMonth.value);
    await refreshBudgetSummaries(target);
    syncBudgetFormsFromSummaries();
  } catch (error) {
    console.error('Failed to load budget summaries', error);
    setBudgetMessage(error.message || 'Failed to load budget details.', 'error');
  }
}

function ensureBudgetFormEntries(budgets) {
  const ids = new Set();
  budgets.forEach((budget) => {
    ids.add(budget.id);
    if (!budgetForms[budget.id]) {
      budgetForms[budget.id] = { override: '', adjustment: '' };
    }
  });
  Object.keys(budgetForms).forEach((id) => {
    if (!ids.has(id)) {
      delete budgetForms[id];
    }
  });
}

function syncBudgetFormsFromSummaries() {
  const summaries = Array.isArray(state.budgetSummaries) ? state.budgetSummaries : [];
  const ids = new Set();
  summaries.forEach((summary) => {
    ids.add(summary.id);
    if (!budgetForms[summary.id]) {
      budgetForms[summary.id] = { override: '', adjustment: '' };
    }
    budgetForms[summary.id].override = summary.overrideAmount != null ? formatEditableAmount(summary.overrideAmount) : '';
    budgetForms[summary.id].adjustment = summary.adjustmentId ? formatEditableAmount(summary.adjustmentAmount) : '';
  });
  Object.keys(budgetForms).forEach((id) => {
    if (!ids.has(id)) {
      if (!budgetForms[id]) {
        budgetForms[id] = { override: '', adjustment: '' };
      }
      budgetForms[id].override = budgetForms[id].override ?? '';
      budgetForms[id].adjustment = budgetForms[id].adjustment ?? '';
    }
  });
}

function formatEditableAmount(value) {
  if (typeof value !== 'number' || Number.isNaN(value)) {
    return '';
  }
  const fixed = value.toFixed(2);
  return fixed.replace(/\.00$/, '');
}

function handleBudgetMonthChange(event) {
  if (event?.target?.value) {
    budgetMonth.value = event.target.value;
  }
}

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

function handleBudgetAmountInput(event) {
  rawBudgetAmount.value = event.target.value.replace(/[^0-9.-]/g, '');
}

function normalizeBudgetAmount(event) {
  const numeric = Number(rawBudgetAmount.value.replace(/[^0-9.-]/g, '')) || 0;
  rawBudgetAmount.value = numeric === 0 ? '' : String(numeric);
  event.target.value = formattedBudgetAmount.value;
}

function setBudgetMessage(text, type) {
  budgetMessage.value = { text, type };
  dismissAfter(() => (budgetMessage.value = { text: '', type: '' }));
}

function getBudgetForm(budgetId) {
  if (!budgetForms[budgetId]) {
    budgetForms[budgetId] = { override: '', adjustment: '' };
  }
  return budgetForms[budgetId];
}

function resetBudgetForm() {
  budgetForm.value = createBudgetForm();
  rawBudgetAmount.value = '';
}

async function submitBudget() {
  if (!budgetForm.value.category) {
    setBudgetMessage('Please choose a category.', 'error');
    return;
  }
  const amount = Number(rawBudgetAmount.value.replace(/[^0-9.-]/g, '')) || 0;
  if (amount <= 0) {
    setBudgetMessage('Enter a valid amount greater than zero.', 'error');
    return;
  }
  const payload = {
    category: budgetForm.value.category,
    amount,
    currency: state.currency,
  };
  const isEdit = Boolean(budgetForm.value.id);
  const url = isEdit
    ? `/budget/edit?id=${encodeURIComponent(budgetForm.value.id)}`
    : '/budget';
  try {
    const response = await apiFetch(url, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to save budget');
    }
    await refreshBudgets();
    await loadBudgetSummaries();
    setBudgetMessage(isEdit ? 'Budget updated.' : 'Budget added.', 'success');
    resetBudgetForm();
  } catch (error) {
    console.error('Failed to save budget', error);
    setBudgetMessage(error.message || 'Failed to save budget', 'error');
  }
}

function editBudget(budget) {
  budgetForm.value = {
    id: budget.id,
    category: budget.category,
    submitLabel: 'Update Budget',
  };
  rawBudgetAmount.value = String(budget.amount);
}

function openDeleteBudget(budget) {
  budgetToDelete.value = budget;
  showDeleteBudget.value = true;
}

function closeDeleteBudget() {
  showDeleteBudget.value = false;
  budgetToDelete.value = null;
}

async function confirmDeleteBudget() {
  if (!budgetToDelete.value) {
    return;
  }
  const target = budgetToDelete.value;
  try {
    const response = await apiFetch(`/budget/delete?id=${encodeURIComponent(target.id)}`, {
      method: 'DELETE',
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to delete budget');
    }
    await refreshBudgets();
    await loadBudgetSummaries();
    setBudgetMessage('Budget removed.', 'success');
    if (budgetForm.value.id === target.id) {
      resetBudgetForm();
    }
  } catch (error) {
    console.error('Failed to delete budget', error);
    setBudgetMessage(error.message || 'Failed to delete budget', 'error');
  } finally {
    closeDeleteBudget();
  }
}

function cleanAmountInput(value) {
  if (value === null || value === undefined) return NaN;
  const numeric = Number(String(value).replace(/[^0-9.-]/g, ''));
  return Number.isNaN(numeric) ? NaN : numeric;
}

async function saveBudgetOverride(budget) {
  const form = getBudgetForm(budget.id);
  const raw = String(form.override ?? '').trim();
  if (!raw) {
    setBudgetMessage('Enter an override amount before saving.', 'error');
    return;
  }
  const amount = cleanAmountInput(raw);
  if (!Number.isFinite(amount) || amount <= 0) {
    setBudgetMessage('Override must be a positive number.', 'error');
    return;
  }
  try {
    const response = await apiFetch('/budget/override', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ budgetId: budget.id, month: budgetMonth.value, amount }),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to save override');
    }
    setBudgetMessage('Override saved.', 'success');
    await loadBudgetSummaries();
  } catch (error) {
    console.error('Failed to save budget override', error);
    setBudgetMessage(error.message || 'Failed to save override.', 'error');
  }
}

async function clearBudgetOverride(budget) {
  const summary = budgetSummaryMap.value[budget.id];
  if (!summary?.overrideId) {
    setBudgetMessage('No override to clear.', 'error');
    return;
  }
  try {
    const response = await apiFetch(`/budget/override?id=${encodeURIComponent(summary.overrideId)}`, {
      method: 'DELETE',
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to clear override');
    }
    getBudgetForm(budget.id).override = '';
    setBudgetMessage('Override removed.', 'success');
    await loadBudgetSummaries();
  } catch (error) {
    console.error('Failed to clear budget override', error);
    setBudgetMessage(error.message || 'Failed to clear override.', 'error');
  }
}

async function saveBudgetAdjustment(budget) {
  const form = getBudgetForm(budget.id);
  const raw = String(form.adjustment ?? '').trim();
  if (!raw) {
    setBudgetMessage('Enter an adjustment amount before saving.', 'error');
    return;
  }
  const amount = cleanAmountInput(raw);
  if (!Number.isFinite(amount) || amount === 0) {
    setBudgetMessage('Adjustment must be a non-zero number.', 'error');
    return;
  }
  try {
    const response = await apiFetch('/budget/adjustment', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ budgetId: budget.id, month: budgetMonth.value, amount }),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to save adjustment');
    }
    setBudgetMessage('Adjustment saved.', 'success');
    await loadBudgetSummaries();
  } catch (error) {
    console.error('Failed to save budget adjustment', error);
    setBudgetMessage(error.message || 'Failed to save adjustment.', 'error');
  }
}

async function clearBudgetAdjustment(budget) {
  const summary = budgetSummaryMap.value[budget.id];
  if (!summary?.adjustmentId) {
    setBudgetMessage('No adjustment to clear.', 'error');
    return;
  }
  try {
    const response = await apiFetch(`/budget/adjustment?id=${encodeURIComponent(summary.adjustmentId)}`, {
      method: 'DELETE',
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to clear adjustment');
    }
    getBudgetForm(budget.id).adjustment = '';
    setBudgetMessage('Adjustment removed.', 'success');
    await loadBudgetSummaries();
  } catch (error) {
    console.error('Failed to clear budget adjustment', error);
    setBudgetMessage(error.message || 'Failed to clear adjustment.', 'error');
  }
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
  const showingAll = categoryDisplayCount.value >= categories.value.length;
  categories.value.push(candidate);
  newCategory.value = '';
  if (showingAll) {
    categoryDisplayCount.value = categories.value.length;
  }
  ensureCategoryDisplayBounds();
}

function removeCategory(index) {
  categories.value.splice(index, 1);
  ensureCategoryDisplayBounds();
}

function moveCategory(index, delta) {
  const newIndex = index + delta;
  if (newIndex < 0 || newIndex >= categories.value.length) return;
  const updated = [...categories.value];
  const [item] = updated.splice(index, 1);
  updated.splice(newIndex, 0, item);
  categories.value = updated;
  ensureCategoryDisplayBounds();
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

function setExportMessage(text, type) {
  exportMessage.value = { text, type };
  if (text) {
    dismissAfter(() => (exportMessage.value = { text: '', type: '' }));
  }
}

function loadMoreCategories() {
  categoryDisplayCount.value = Math.min(categoryDisplayCount.value + 5, categories.value.length);
}

function ensureCategoryDisplayBounds() {
  if (categories.value.length === 0) {
    categoryDisplayCount.value = 5;
    return;
  }
  if (categoryDisplayCount.value > categories.value.length) {
    categoryDisplayCount.value = categories.value.length;
  }
  if (categoryDisplayCount.value < Math.min(5, categories.value.length)) {
    categoryDisplayCount.value = Math.min(5, categories.value.length);
  }
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

async function exportCsv() {
  if (exportingCsv.value) return;
  if (!exportAll.value && exportDateFilter.value === 'range' && exportRangeValidationMessage.value) {
    setExportMessage(exportRangeValidationMessage.value, 'error');
    return;
  }
  const params = new URLSearchParams();
  if (exportAll.value) {
    params.set('period', 'all');
  } else {
    params.set('period', exportDateFilter.value);
    if (exportDateFilter.value === 'range') {
      params.set('start', exportRangeStart.value);
      params.set('end', exportRangeEnd.value);
    }
  }
  const query = params.toString();
  const target = query ? `/export/csv?${query}` : '/export/csv';
  exportingCsv.value = true;
  setExportMessage('', '');
  try {
    const response = await apiFetch(target, {
      method: 'GET',
      headers: { Accept: 'text/csv' },
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to export CSV');
    }
    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'expenses.csv';
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
    setExportMessage('Download should begin shortly.', 'success');
  } catch (error) {
    console.error('Failed to export CSV', error);
    setExportMessage(error.message || 'Failed to export CSV', 'error');
  } finally {
    exportingCsv.value = false;
  }
}

function setStartDateMessage(text, type) {
  startDateMessage.value = { text, type };
  dismissAfter(() => (startDateMessage.value = { text: '', type: '' }));
}

async function saveStartDate() {
  if (endOfMonth.value) {
    setStartDateMessage('Disable end-of-month alignment to edit the start date.', 'error');
    return;
  }
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
    await loadBudgetSummaries();
  } catch (error) {
    console.error('Failed to save start date', error);
    setStartDateMessage(error.message || 'Failed to save start date', 'error');
  }
}

function setEndOfMonthMessage(text, type) {
  endOfMonthMessage.value = { text, type };
  dismissAfter(() => (endOfMonthMessage.value = { text: '', type: '' }));
}

async function saveEndOfMonth() {
  try {
    const response = await apiFetch('/end-of-month/edit', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(endOfMonth.value),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || 'Failed to update preference');
    }
    state.endOfMonth = endOfMonth.value;
    setEndOfMonthMessage('Monthly alignment updated.', 'success');
    await loadBudgetSummaries();
  } catch (error) {
    console.error('Failed to update end-of-month preference', error);
    endOfMonth.value = state.endOfMonth;
    setEndOfMonthMessage(error.message || 'Failed to update preference', 'error');
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

function handleRecurringAmountInput(event) {
  rawRecurringAmount.value = event.target.value.replace(/[^0-9.-]/g, '');
}

function normalizeRecurringAmount(event) {
  const numeric = Number(rawRecurringAmount.value.replace(/[^0-9.-]/g, '')) || 0;
  rawRecurringAmount.value = numeric === 0 ? '' : String(numeric);
  event.target.value = formattedRecurringAmount.value;
}

function setRecurringMessage(text, type) {
  recurringMessage.value = { text, type };
  dismissAfter(() => (recurringMessage.value = { text: '', type: '' }));
}

function resetRecurringForm() {
  recurringForm.value = createRecurringForm();
  editingRecurringId.value = null;
  rawRecurringAmount.value = '';
}

async function submitRecurring() {
  if (!recurringForm.value.category) {
    setRecurringMessage('Please select a category.', 'error');
    return;
  }
  let amount = Number(rawRecurringAmount.value.replace(/[^0-9.-]/g, ''));
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
    startDate: getISODateWithLocalTime(recurringForm.value.startDate),
    occurrences: recurringForm.value.occurrences,
  };
  const blob = await encryptPayload(payload);
  if (blob) {
    payload.blob = blob;
  }
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
    startDate: expense.startDate?.slice(0, 10) || createRecurringForm().startDate,
    occurrences: expense.occurrences ?? 0,
    reportGain: expense.amount > 0,
    submitLabel: 'Update Recurring Transaction',
  };
  rawRecurringAmount.value = String(Math.abs(expense.amount));
  nextTick(() => {
    recurringCardRef.value?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  });
}

function openDeleteRecurring(expense) {
  expenseToDelete.value = expense;
  showDeleteRecurring.value = true;
}

function closeDeleteRecurring() {
  showDeleteRecurring.value = false;
  expenseToDelete.value = null;
}

async function confirmDeleteRecurring() {
  if (!expenseToDelete.value) return;
  try {
    const params = new URLSearchParams({ removeAll: 'true' });
    const response = await apiFetch(`/recurring-expense/delete?id=${expenseToDelete.value.id}&${params.toString()}`, {
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
  } finally {
    closeDeleteRecurring();
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

<style scoped>
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.animate-in {
  animation: fadeIn 0.5s ease-out fill-mode-both;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}

::-webkit-scrollbar {
  width: 6px;
}
::-webkit-scrollbar-track {
  background: transparent;
}
::-webkit-scrollbar-thumb {
  background: var(--border);
  border-radius: 10px;
}
</style>
