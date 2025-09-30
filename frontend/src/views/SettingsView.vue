<template>
  <section class="space-y-6">
    <div :class="cardClass">
      <h2 align="center" class="text-xl font-semibold text-[var(--text-primary)]">Category Settings</h2>
      <div class="mt-4 space-y-3">
        <div
          v-for="item in visibleCategories"
          :key="item.category"
          class="flex items-center justify-between gap-3 rounded-full border border-[var(--border)] bg-[var(--bg-primary)]/60 px-4 py-2 text-sm text-[var(--text-primary)]"
        >
          <div class="flex items-center gap-3">
            <span class="text-[var(--text-secondary)]"><i class="fa-solid fa-grip-lines"></i></span>
            <span>{{ item.category }}</span>
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              :class="iconButtonTiny"
              @click="moveCategory(item.index, -1)"
              :disabled="item.index === 0"
            >
              <i class="fa-solid fa-arrow-up"></i>
            </button>
            <button
              type="button"
              :class="iconButtonTiny"
              @click="moveCategory(item.index, 1)"
              :disabled="item.index === categories.length - 1"
            >
              <i class="fa-solid fa-arrow-down"></i>
            </button>
            <button type="button" :class="iconDangerButtonTiny" @click="removeCategory(item.index)">
              <i class="fa-solid fa-times"></i>
            </button>
          </div>
        </div>
        <div
          v-if="canLoadMoreCategories"
          class="flex justify-center"
        >
          <button type="button" :class="primaryButtonClass" @click="loadMoreCategories">Load More</button>
        </div>
        <div class="flex flex-col gap-3 sm:flex-row">
          <input v-model="newCategory" type="text" placeholder="Add new category" :class="inputClass" />
          <button type="button" :class="[primaryButtonClass, 'sm:w-auto']" @click="addCategory">
            Add
          </button>
        </div>
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <button type="button" :class="primaryButtonClass" @click="saveCategories">Save Categories</button>
          <div
            v-if="categoryMessage.text"
            :class="[
              'rounded-full px-4 py-2 text-center text-sm font-medium',
              categoryMessage.type === 'success'
                ? 'bg-emerald-500/20 text-emerald-200'
                : 'bg-rose-500/20 text-rose-200'
            ]"
          >
            {{ categoryMessage.text }}
          </div>
        </div>
    </div>
  </div>

    <div :class="cardClass">
      <h2 align="center" class="text-xl font-semibold text-[var(--text-primary)]">Budget Settings</h2>
      <form class="mt-4 grid gap-4 md:grid-cols-2" @submit.prevent="submitBudget">
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)]" for="budgetCategory">Category</label>
          <select id="budgetCategory" v-model="budgetForm.category" :class="inputClass" required>
            <option value="" disabled>Select category</option>
            <option v-for="category in budgetCategoryOptions" :key="category" :value="category">
              {{ category }}
            </option>
          </select>
        </div>
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)]" for="budgetAmount">Monthly Limit</label>
          <input
            id="budgetAmount"
            :value="formattedBudgetAmount"
            inputmode="decimal"
            :class="inputClass"
            required
            @input="handleBudgetAmountInput"
            @blur="normalizeBudgetAmount"
          />
        </div>
        <div class="flex flex-col gap-2 md:col-span-2 md:flex-row md:items-center md:justify-between">
          <div class="flex gap-2">
            <button type="submit" :class="[primaryButtonClass, 'w-full md:w-auto']">{{ budgetForm.submitLabel }}</button>
            <button
              v-if="budgetForm.id"
              type="button"
              :class="[primaryButtonClass, 'w-full md:w-auto']"
              @click="resetBudgetForm"
            >
              Cancel
            </button>
          </div>
          <div
            v-if="budgetMessage.text"
            :class="[
              'rounded-full px-4 py-2 text-center text-sm font-medium',
              budgetMessage.type === 'success'
                ? 'bg-emerald-500/20 text-emerald-200'
                : 'bg-rose-500/20 text-rose-200'
            ]"
          >
            {{ budgetMessage.text }}
          </div>
        </div>
      </form>

      <div class="mt-6 space-y-4">
        <div v-if="budgetList.length" class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <label class="text-sm font-medium text-[var(--text-secondary)]" for="budgetMonth">Adjust month</label>
          <div class="flex items-center gap-3 sm:gap-4">
            <input id="budgetMonth" type="month" v-model="budgetMonth" :class="inputClass" @change="handleBudgetMonthChange" />
            <span class="text-xs text-[var(--text-secondary)]">Showing overrides for {{ budgetMonthLabel }}</span>
          </div>
        </div>
        <template v-if="budgetList.length === 0">
          <div
            class="rounded-3xl border border-dashed border-[var(--border)] bg-[var(--bg-secondary)]/60 py-6 text-center text-sm italic text-[var(--text-secondary)]"
          >
            No budgets configured yet. Add one above to get started.
          </div>
        </template>
        <template v-else>
          <div
            v-for="budget in budgetList"
            :key="budget.id"
            class="space-y-4 rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)]/60 px-5 py-5 shadow-card backdrop-blur"
          >
            <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <div class="space-y-1 text-sm text-[var(--text-secondary)]">
                <div class="text-base font-semibold text-[var(--text-primary)]">{{ budget.category }}</div>
                <div>Base limit: <span class="font-mono text-[var(--text-primary)]">{{ formatCurrency(budget.amount) }}</span></div>
                <div v-if="budgetSummaryMap[budget.id]" class="text-xs text-[var(--text-secondary)]">
                  Effective for {{ budgetMonthLabel }}:
                  <span class="font-mono text-[var(--text-primary)]">{{ formatCurrency(budgetSummaryMap[budget.id].effectiveAmount || budget.amount) }}</span>
                </div>
                <div v-if="budgetSummaryMap[budget.id]?.overrideAmount != null" class="text-xs text-[var(--text-secondary)]">
                  Override: {{ formatCurrency(budgetSummaryMap[budget.id].overrideAmount || 0) }}
                </div>
                <div v-if="budgetSummaryMap[budget.id]?.adjustmentId" class="text-xs text-[var(--text-secondary)]">
                  Adjustment: {{ formatCurrency(budgetSummaryMap[budget.id].adjustmentAmount || 0) }}
                </div>
              </div>
              <div class="flex items-center gap-2">
                <button type="button" :class="iconDangerButtonTiny" @click="openDeleteBudget(budget)">
                  <i class="fa-solid fa-trash-can"></i>
                </button>
              </div>
            </div>
            <div class="grid gap-4 md:grid-cols-2" v-if="budgetForms[budget.id]">
              <div class="space-y-2">
                <label class="text-xs font-semibold uppercase tracking-wide text-[var(--text-secondary)]" :for="`override-${budget.id}`">
                  Per-month override
                </label>
                <div class="flex flex-col gap-2 sm:flex-row sm:items-center">
                  <input
                    :id="`override-${budget.id}`"
                    v-model="budgetForms[budget.id].override"
                    :class="inputClass"
                    placeholder="Override amount"
                  />
                  <div class="flex gap-2">
                    <button type="button" :class="[primaryButtonClass, 'text-xs px-4 py-1.5']" @click="saveBudgetOverride(budget)">
                      Save
                    </button>
                    <button
                      v-if="budgetSummaryMap[budget.id]?.overrideId"
                      type="button"
                      :class="[primaryButtonClass, 'text-xs px-4 py-1.5']"
                      @click="clearBudgetOverride(budget)"
                    >
                      Clear
                    </button>
                  </div>
                </div>
              </div>
              <div class="space-y-2">
                <label class="text-xs font-semibold uppercase tracking-wide text-[var(--text-secondary)]" :for="`adjustment-${budget.id}`">
                  One-time adjustment
                </label>
                <div class="flex flex-col gap-2 sm:flex-row sm:items-center">
                  <input
                    :id="`adjustment-${budget.id}`"
                    v-model="budgetForms[budget.id].adjustment"
                    :class="inputClass"
                    placeholder="Adjustment (e.g., 50 or -25)"
                  />
                  <div class="flex gap-2">
                    <button type="button" :class="[primaryButtonClass, 'text-xs px-4 py-1.5']" @click="saveBudgetAdjustment(budget)">
                      Save
                    </button>
                    <button
                      v-if="budgetSummaryMap[budget.id]?.adjustmentId"
                      type="button"
                      :class="[primaryButtonClass, 'text-xs px-4 py-1.5']"
                      @click="clearBudgetAdjustment(budget)"
                    >
                      Clear
                    </button>
                  </div>
                </div>
                <p class="text-[11px] text-[var(--text-secondary)]">Positive numbers add to the limit; negative numbers reduce it for this month only.</p>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>

    <div class="grid gap-6 md:grid-cols-2">
      <div :class="cardClass">
        <h2 align="center" class="text-xl font-semibold text-[var(--text-primary)]">Currency Settings</h2>
        <div class="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center">
          <select v-model="currencyCode" :class="inputClass">
            <option v-for="code in currencyOptions" :key="code" :value="code">
              {{ code.toUpperCase() }} ({{ currencyBehaviors[code].symbol }})
            </option>
          </select>
          <button type="button" :class="primaryButtonClass" @click="saveCurrency">Save</button>
        </div>
        <div
          v-if="currencyMessage.text"
          :class="[
            'mt-4 rounded-full px-4 py-2 text-center text-sm font-medium',
            currencyMessage.type === 'success'
              ? 'bg-emerald-500/20 text-emerald-200'
              : 'bg-rose-500/20 text-rose-200'
          ]"
        >
          {{ currencyMessage.text }}
        </div>
      </div>
      <div :class="cardClass">
        <h2 align="center" class="text-xl font-semibold text-[var(--text-primary)]">Start Date Settings</h2>
        <div class="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center">
          <input v-model.number="startDate" type="number" min="1" max="31" :class="inputClass" :disabled="endOfMonth" />
          <button type="button" :class="primaryButtonClass" @click="saveStartDate" :disabled="endOfMonth">
            Save
          </button>
        </div>
        <div
          v-if="startDateMessage.text"
          :class="[
            'mt-4 rounded-full px-4 py-2 text-center text-sm font-medium',
            startDateMessage.type === 'success'
              ? 'bg-emerald-500/20 text-emerald-200'
              : 'bg-rose-500/20 text-rose-200'
          ]"
        >
          {{ startDateMessage.text }}
        </div>
        <div class="mt-6 space-y-4 rounded-2xl border border-[var(--border)] bg-[var(--bg-primary)]/60 p-4">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div class="sm:max-w-[70%]">
              <div class="text-sm font-medium text-[var(--text-primary)]">Align cycle to month end</div>
              <p class="text-xs text-[var(--text-secondary)]">When enabled, each budgeting period runs from the last day of the previous month through the last day of the current month.</p>
            </div>
            <label class="relative inline-flex h-6 w-12 cursor-pointer items-center">
              <input
                type="checkbox"
                class="peer sr-only"
                v-model="endOfMonth"
              />
              <span class="absolute inset-0 rounded-full bg-[var(--border)] transition-colors duration-200 peer-checked:bg-[var(--accent)]"></span>
              <span class="absolute left-1 h-4 w-4 rounded-full bg-white transition-transform duration-200 peer-checked:translate-x-6"></span>
            </label>
          </div>
          <div class="flex justify-end">
            <button
              type="button"
              :class="primaryButtonClass"
              @click="saveEndOfMonth"
              :disabled="endOfMonth === state.endOfMonth"
            >
              Update
            </button>
          </div>
          <div
            v-if="endOfMonthMessage.text"
            :class="[
              'rounded-full px-4 py-2 text-center text-sm font-medium',
              endOfMonthMessage.type === 'success'
                ? 'bg-emerald-500/20 text-emerald-200'
                : 'bg-rose-500/20 text-rose-200'
            ]"
          >
            {{ endOfMonthMessage.text }}
          </div>
        </div>
      </div>
    </div>

    <div class="grid gap-6 md:grid-cols-2">
      <div :class="cardClass">
        <h2 align="center" class="text-xl font-semibold text-[var(--text-primary)]">Theme Settings</h2>
        <div class="mt-4">
          <select v-model="theme" @change="applyTheme" :class="inputClass">
            <option value="system">System Default</option>
            <option value="light">Light</option>
            <option value="dark">Dark</option>
          </select>
        </div>
        <div
          v-if="themeMessage.text"
          :class="[
            'mt-4 rounded-full px-4 py-2 text-center text-sm font-medium',
            themeMessage.type === 'success'
              ? 'bg-emerald-500/20 text-emerald-200'
              : 'bg-rose-500/20 text-rose-200'
          ]"
        >
          {{ themeMessage.text }}
        </div>
      </div>
      <div :class="cardClass">
        <h2 align="center" class="text-xl font-semibold text-[var(--text-primary)]">Import/Export Data</h2>
        <div class="mt-4 flex flex-col items-center justify-center gap-3 text-xs md:flex-row md:text-sm">
          <button
            type="button"
            :class="[primaryButtonClass, 'text-xs md:text-sm w-full md:w-auto whitespace-nowrap']"
            @click="exportCsv"
            :disabled="exportingCsv"
          >
            {{ exportingCsv ? 'Exporting…' : 'Export to CSV' }}
          </button>
          <label :class="[primaryButtonClass, 'text-xs md:text-sm w-full md:w-auto whitespace-nowrap']" for="csv-import-file">Import from CSV</label>
          <input id="csv-import-file" ref="csvImportRef" type="file" accept=".csv" hidden @change="(event) => handleImport(event, '/import/csv')" />
        </div>
        <div
          v-if="exportMessage.text"
          :class="[
            'mt-3 rounded-full px-4 py-2 text-center text-sm font-medium',
            exportMessage.type === 'success'
              ? 'bg-emerald-500/20 text-emerald-200'
              : 'bg-rose-500/20 text-rose-200'
          ]"
        >
          {{ exportMessage.text }}
        </div>
        <div
          v-if="importMessage.text"
          :class="[
            'mt-4 rounded-full px-4 py-2 text-center text-sm font-medium',
            importMessage.type === 'success'
              ? 'bg-emerald-500/20 text-emerald-200'
              : 'bg-rose-500/20 text-rose-200'
          ]"
        >
          {{ importMessage.text }}
        </div>
        <div
          v-if="importSummary"
          class="mt-4 space-y-1 rounded-2xl border border-[var(--border)] bg-[var(--bg-primary)]/60 px-4 py-4 text-sm text-[var(--text-secondary)]"
        >
          <h3 class="text-base font-semibold text-[var(--text-primary)]">Import Summary</h3>
          <p>Total Processed: <span class="font-semibold text-[var(--text-primary)]">{{ importSummary.totalProcessed }}</span></p>
          <p>Imported: <span class="font-semibold text-emerald-300">{{ importSummary.imported }}</span></p>
          <p>Skipped: <span class="font-semibold text-rose-300">{{ importSummary.skipped }}</span></p>
          <p>New Categories: <span class="font-semibold text-[var(--text-primary)]">{{ importSummary.newCategories }}</span></p>
        </div>
      </div>
    </div>

    <div :class="cardClass" ref="recurringCardRef">
      <h2 align="center" class="text-xl font-semibold text-[var(--text-primary)]">Recurring Transactions</h2>
      <form class="mt-4 grid gap-4 md:grid-cols-2" @submit.prevent="submitRecurring">
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)]" for="recurringName">Name</label>
          <input id="recurringName" v-model="recurringForm.name" type="text" :class="inputClass" required />
        </div>
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)]" for="recurringAmount">Amount</label>
          <input
            id="recurringAmount"
            :value="formattedRecurringAmount"
            inputmode="decimal"
            :class="inputClass"
            required
            @input="handleRecurringAmountInput"
            @blur="normalizeRecurringAmount"
          />
        </div>
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)]" for="recurringCategory">Category</label>
          <select id="recurringCategory" v-model="recurringForm.category" :class="inputClass" required>
            <option value="" disabled>Select category</option>
            <option v-for="category in state.categories" :key="category" :value="category">{{ category }}</option>
          </select>
        </div>
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)]">Tags</label>
          <TagInput v-model="recurringForm.tags" :suggestions="allTags" />
        </div>
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)]" for="recurringInterval">Interval</label>
          <select id="recurringInterval" v-model="recurringForm.interval" :class="inputClass" required>
            <option value="daily">Daily</option>
            <option value="weekly">Weekly</option>
            <option value="monthly">Monthly</option>
            <option value="yearly">Yearly</option>
          </select>
        </div>
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)]" for="recurringStartDate">Start Date</label>
          <input id="recurringStartDate" v-model="recurringForm.startDate" type="date" :class="inputClass" required />
        </div>
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)]" for="recurringOccurrences">Occurrences (0 for indefinite)</label>
          <input id="recurringOccurrences" v-model.number="recurringForm.occurrences" type="number" min="0" :class="inputClass" required />
        </div>
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium text-[var(--text-secondary)] mb-2" for="recurringReportGain">Report Gain</label>
          <label class="relative inline-flex h-6 w-12 cursor-pointer items-center">
            <input
              id="recurringReportGain"
              v-model="recurringForm.reportGain"
              type="checkbox"
              class="peer sr-only"
            />
            <span class="absolute inset-0 rounded-full bg-[var(--border)] transition-colors duration-200 peer-checked:bg-[var(--accent)]"></span>
            <span class="absolute left-1 h-4 w-4 rounded-full bg-white transition-transform duration-200 peer-checked:translate-x-6"></span>
          </label>
        </div>
        <div class="md:col-span-2">
          <button type="submit" :class="[primaryButtonClass, 'w-full']">{{ recurringForm.submitLabel }}</button>
        </div>
      </form>
      <div
        v-if="recurringMessage.text"
        :class="[
          'mt-4 rounded-full px-4 py-2 text-center text-sm font-medium',
          recurringMessage.type === 'success'
            ? 'bg-emerald-500/20 text-emerald-200'
            : 'bg-rose-500/20 text-rose-200'
        ]"
      >
        {{ recurringMessage.text }}
      </div>

      <h3 class="mt-10 text-center text-lg font-semibold text-[var(--text-primary)]">Existing Recurring Transactions</h3>
      <div
        v-if="state.recurringExpenses.length === 0"
        class="mt-4 rounded-3xl border border-dashed border-[var(--border)] bg-[var(--bg-secondary)]/60 py-10 text-center text-base italic text-[var(--text-secondary)]"
      >
        No recurring transactions configured.
      </div>
      <div v-else class="mt-6 space-y-3">
        <div
          v-for="expense in state.recurringExpenses"
          :key="expense.id"
          class="flex flex-col gap-3 rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)]/60 px-5 py-4 shadow-card backdrop-blur sm:flex-row sm:items-center sm:justify-between"
        >
          <div class="space-y-2 text-sm text-[var(--text-secondary)]">
            <strong class="block text-base text-[var(--text-primary)]">{{ expense.name }}</strong>
            <div class="flex flex-wrap items-center gap-2">
              <span>{{ formatCurrency(expense.amount) }}</span>
              <span>• {{ expense.interval }}</span>
              <span>• Starts {{ formatDate(expense.startDate) }}</span>
              <span v-if="expense.occurrences && expense.occurrences > 0">• {{ expense.occurrences }} occurrences</span>
            </div>
            <div v-if="expense.tags && expense.tags.length" class="text-xs uppercase tracking-wide text-[var(--text-secondary)]">
              Tags: <span class="font-medium text-[var(--text-primary)]">{{ expense.tags.join(', ') }}</span>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <button type="button" :class="iconButtonTiny" @click="editRecurring(expense)">
              <i class="fa-solid fa-pen-to-square"></i>
            </button>
            <button type="button" :class="iconDangerButtonTiny" @click="openDeleteRecurring(expense)">
              <i class="fa-solid fa-trash-can"></i>
            </button>
          </div>
        </div>
      </div>
    </div>
  </section>

  <transition name="fade">
    <div
      v-if="showDeleteBudget"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 px-4"
      @click.self="closeDeleteBudget"
    >
      <div class="w-full max-w-md rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)]/95 p-6 shadow-card backdrop-blur">
        <h3 class="text-lg font-semibold text-[var(--text-primary)]">Delete Budget</h3>
        <p class="mt-2 text-sm text-[var(--text-secondary)]">
          Are you sure you want to remove the budget for
          <span class="font-semibold text-[var(--text-primary)]">{{ budgetToDelete?.category }}</span>? Any overrides or adjustments for this month will also be cleared.
        </p>
        <div class="mt-6 flex justify-end gap-3">
          <button :class="primaryButtonClass" @click="closeDeleteBudget">Cancel</button>
          <button :class="[primaryButtonClass, 'bg-rose-500 text-white hover:bg-rose-500/90']" @click="confirmDeleteBudget">
            Delete
          </button>
        </div>
      </div>
    </div>
  </transition>

  <transition name="fade">
    <div
      v-if="showDeleteRecurring"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 px-4"
      @click.self="closeDeleteRecurring"
    >
      <div class="w-full max-w-md rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)]/95 p-6 shadow-card backdrop-blur">
        <h3 class="text-lg font-semibold text-[var(--text-primary)]">Delete Recurring Expense</h3>
        <p class="mt-2 text-sm text-[var(--text-secondary)]">
          Are you sure you want to remove this recurring transaction? Future occurrences will be deleted.
        </p>
        <div class="mt-6 flex justify-end gap-3">
          <button :class="primaryButtonClass" @click="closeDeleteRecurring">Cancel</button>
          <button :class="[primaryButtonClass, 'bg-rose-500 text-white hover:bg-rose-500/90']" @click="confirmDeleteRecurring">
            Delete
          </button>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted, nextTick } from 'vue';
import TagInput from '../components/TagInput.vue';
import state, { loadInitialData, refreshExpenses, refreshRecurringExpenses, refreshBudgets, refreshBudgetSummaries } from '../stores/appState';
import { apiFetch } from '../lib/api';
import { encryptPayload } from '../lib/encryption';
import { currencyBehaviors, formatCurrency as formatCurrencyRaw, getISODateWithLocalTime, formatMonth as formatMonthLabel } from '../lib/utils';

const primaryButtonClass =
  'inline-flex items-center justify-center gap-2 rounded-full border border-[var(--border)] bg-[var(--bg-secondary)] px-5 py-2 text-sm font-medium text-[var(--text-primary)] transition duration-150 ease-out hover:bg-[var(--accent)] hover:text-white hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40 focus:ring-offset-2 focus:ring-offset-[var(--bg-primary)] disabled:cursor-not-allowed disabled:opacity-50';

const inputClass =
  'w-full rounded-xl border border-[var(--border)] bg-[var(--bg-primary)] px-4 py-2 text-[var(--text-primary)] placeholder:text-[var(--text-secondary)] focus:border-[var(--accent)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/40';

const checkboxClass =
  'h-4 w-4 rounded border-[var(--border)] bg-[var(--bg-primary)] text-[var(--accent)] focus:ring-[var(--accent)]/60 focus:ring-offset-0';

const cardClass =
  'rounded-3xl border border-[var(--border)] bg-[var(--bg-secondary)]/80 p-6 shadow-card backdrop-blur';

const iconButtonTiny =
  'inline-flex h-8 w-8 items-center justify-center rounded-full border border-[var(--border)] bg-[var(--bg-secondary)] text-xs text-[var(--text-primary)] transition duration-150 hover:bg-[var(--accent)] hover:text-white hover:shadow-md disabled:cursor-not-allowed disabled:opacity-40';

const iconDangerButtonTiny =
  'inline-flex h-8 w-8 items-center justify-center rounded-full border border-rose-500/40 bg-rose-500/10 text-xs text-rose-400 transition duration-150 hover:bg-rose-500/20 hover:text-white';

const infoLinkClass =
  'font-semibold text-[var(--text-primary)] underline decoration-dotted underline-offset-4 transition hover:text-[var(--accent)]';

const currencyOptions = Object.keys(currencyBehaviors);

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

const importMessage = ref({ text: '', type: '' });
const exportMessage = ref({ text: '', type: '' });
const exportingCsv = ref(false);
const importSummary = ref(null);
const csvImportRef = ref(null);
const csvImportOldRef = ref(null);

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
});

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
  exportingCsv.value = true;
  setExportMessage('', '');
  try {
    const response = await apiFetch('/export/csv', {
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
