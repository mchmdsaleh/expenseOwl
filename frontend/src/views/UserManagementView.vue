<template>
  <section class="space-y-6">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-[var(--text-primary)]">User Management</h1>
        <p class="text-sm text-[var(--text-secondary)]">
          Promote or demote accounts and review who has access to ExpenseOwl.
        </p>
      </div>
      <button
        type="button"
        class="rounded-lg border border-[var(--border)] px-3 py-2 text-sm text-[var(--text-secondary)] hover:bg-[var(--bg-secondary)]"
        :disabled="loading"
        @click="fetchUsers"
      >
        Refresh
      </button>
    </header>

    <div v-if="errorMessage" class="rounded-md border border-red-400 bg-red-100 px-3 py-2 text-sm text-red-700">
      {{ errorMessage }}
    </div>

    <div class="overflow-x-auto rounded-lg border border-[var(--border)]">
      <table class="min-w-full divide-y divide-[var(--border)] text-sm">
        <thead class="bg-[var(--bg-secondary)] text-[var(--text-secondary)]">
          <tr>
            <th class="px-4 py-3 text-left font-medium">Email</th>
            <th class="px-4 py-3 text-left font-medium">Name</th>
            <th class="px-4 py-3 text-left font-medium">Role</th>
            <th class="px-4 py-3 text-left font-medium">Created</th>
            <th class="px-4 py-3 text-left font-medium">Updated</th>
            <th class="px-4 py-3">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[var(--border)] bg-[var(--bg-primary)] text-[var(--text-primary)]">
          <tr v-if="!loading && users.length === 0">
            <td class="px-4 py-6 text-center text-sm text-[var(--text-secondary)]" colspan="6">
              No users yet.
            </td>
          </tr>
          <tr v-for="user in users" :key="user.id">
            <td class="px-4 py-3">{{ user.email }}</td>
            <td class="px-4 py-3">{{ user.firstName }} {{ user.lastName }}</td>
            <td class="px-4 py-3">
              <select
                v-model="draftRoles[user.id]"
                class="rounded-lg border border-[var(--border)] bg-[var(--bg-primary)] px-2 py-1 text-sm"
              >
                <option value="admin">Admin</option>
                <option value="user">User</option>
              </select>
            </td>
            <td class="px-4 py-3">{{ formatDate(user.createdAt) }}</td>
            <td class="px-4 py-3">{{ formatDate(user.updatedAt) }}</td>
            <td class="px-4 py-3 text-right">
              <button
                type="button"
                class="rounded-lg bg-[var(--accent)] px-3 py-1 text-xs font-semibold text-white hover:bg-[var(--accent)]/90 disabled:opacity-50"
                :disabled="loading || draftRoles[user.id] === user.role"
                @click="updateRole(user)"
              >
                Save
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue';
import { apiFetch } from '../lib/api';
import state from '../stores/appState';

const users = ref([]);
const loading = ref(false);
const errorMessage = ref('');
const draftRoles = reactive({});

function formatDate(value) {
  if (!value) return '—';
  try {
    return new Intl.DateTimeFormat(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    }).format(new Date(value));
  } catch (error) {
    return value;
  }
}

async function fetchUsers() {
  loading.value = true;
  errorMessage.value = '';
  try {
    const response = await apiFetch('/api/v1/admin/users');
    if (!response.ok) throw new Error('Failed to load users');
    const data = await response.json();
    users.value = Array.isArray(data) ? data : [];
    users.value.forEach((user) => {
      draftRoles[user.id] = user.role;
    });
  } catch (error) {
    console.error('Failed to fetch users', error);
    errorMessage.value = error.message || 'Failed to fetch users';
  } finally {
    loading.value = false;
  }
}

async function updateRole(user) {
  loading.value = true;
  errorMessage.value = '';
  try {
    const response = await apiFetch('/api/v1/admin/users/role', {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ id: user.id, role: draftRoles[user.id] }),
    });
    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      throw new Error(data?.error || 'Failed to update role');
    }
    user.role = draftRoles[user.id];
    user.updatedAt = new Date().toISOString();
    if (state.user?.id === user.id) {
      state.user.role = user.role;
    }
  } catch (error) {
    console.error('Failed to update role', error);
    errorMessage.value = error.message || 'Failed to update role';
    draftRoles[user.id] = user.role;
  } finally {
    loading.value = false;
  }
}

onMounted(fetchUsers);
</script>
