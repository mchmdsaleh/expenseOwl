<template>
  <section class="space-y-8">
    <header class="space-y-1">
      <h1 class="text-2xl font-semibold text-[var(--text-primary)]">Profile</h1>
      <p class="text-sm text-[var(--text-secondary)]">
        Update your account details or change your password.
      </p>
    </header>

    <div class="space-y-8">
      <form class="space-y-4" @submit.prevent="saveProfile">
        <h2 class="text-lg font-semibold text-[var(--text-primary)]">Account Details</h2>
        <p class="text-sm text-[var(--text-secondary)]">
          Update the email address and name associated with your account.
        </p>

        <div class="space-y-2">
          <label class="block text-sm font-medium text-[var(--text-secondary)]" for="profile-email">Email</label>
          <input
            id="profile-email"
            v-model="profileForm.email"
            type="email"
            required
            class="w-full rounded-lg border border-[var(--border)] bg-[var(--bg-primary)] px-3 py-2 text-[var(--text-primary)] focus:border-[var(--accent)] focus:outline-none"
          />
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <div class="space-y-2">
            <label class="block text-sm font-medium text-[var(--text-secondary)]" for="profile-first-name">First Name</label>
            <input
              id="profile-first-name"
              v-model="profileForm.firstName"
              type="text"
              required
              class="w-full rounded-lg border border-[var(--border)] bg-[var(--bg-primary)] px-3 py-2 text-[var(--text-primary)] focus:border-[var(--accent)] focus:outline-none"
            />
          </div>
          <div class="space-y-2">
            <label class="block text-sm font-medium text-[var(--text-secondary)]" for="profile-last-name">Last Name</label>
            <input
              id="profile-last-name"
              v-model="profileForm.lastName"
              type="text"
              required
              class="w-full rounded-lg border border-[var(--border)] bg-[var(--bg-primary)] px-3 py-2 text-[var(--text-primary)] focus:border-[var(--accent)] focus:outline-none"
            />
          </div>
        </div>

        <div v-if="profileError" class="rounded-md border border-red-400 bg-red-100 px-3 py-2 text-sm text-red-700">
          {{ profileError }}
        </div>
        <div v-if="profileSuccess" class="rounded-md border border-emerald-400 bg-emerald-100 px-3 py-2 text-sm text-emerald-800">
          {{ profileSuccess }}
        </div>

        <div class="flex items-center gap-3">
          <button
            type="submit"
            class="rounded-lg bg-[var(--accent)] px-4 py-2 text-sm font-semibold text-white transition hover:bg-[var(--bg-secondary)] hover:border-[var(--border)] disabled:opacity-50"
            :disabled="profileSaving"
          >
            {{ profileSaving ? 'Saving…' : 'Save Changes' }}
          </button>
          <button
            type="button"
            class="rounded-lg border border-[var(--border)] px-4 py-2 text-sm font-semibold text-[var(--text-primary)] transition hover:bg-[var(--bg-secondary)]"
            @click="openPasswordModal"
          >
            Change Password
          </button>
        </div>
      </form>
    </div>

    <transition name="fade">
      <div v-if="showPasswordModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
        <div class="w-full max-w-md rounded-xl border border-[var(--border)] bg-[var(--bg-primary)] p-6 shadow-2xl">
          <header class="mb-4 flex items-center justify-between">
            <h3 class="text-lg font-semibold text-[var(--text-primary)]">Update Password</h3>
            <button
              type="button"
              class="text-[var(--text-secondary)] transition hover:text-[var(--text-primary)]"
              @click="closePasswordModal"
            >
              <i class="fa-solid fa-xmark"></i>
            </button>
          </header>

          <form class="space-y-4" @submit.prevent="changePassword">
            <div class="space-y-2">
              <label class="block text-sm font-medium text-[var(--text-secondary)]" for="current-password">Current Password</label>
              <input
                id="current-password"
                v-model="passwordForm.currentPassword"
                type="password"
                required
                class="w-full rounded-lg border border-[var(--border)] bg-[var(--bg-primary)] px-3 py-2 text-[var(--text-primary)] focus:border-[var(--accent)] focus:outline-none"
              />
            </div>

            <div class="space-y-2">
              <label class="block text-sm font-medium text-[var(--text-secondary)]" for="new-password">New Password</label>
              <input
                id="new-password"
                v-model="passwordForm.newPassword"
                type="password"
                minlength="6"
                required
                class="w-full rounded-lg border border-[var(--border)] bg-[var(--bg-primary)] px-3 py-2 text-[var(--text-primary)] focus:border-[var(--accent)] focus:outline-none"
              />
            </div>

            <div class="space-y-2">
              <label class="block text-sm font-medium text-[var(--text-secondary)]" for="confirm-password">Confirm New Password</label>
              <input
                id="confirm-password"
                v-model="passwordForm.confirmPassword"
                type="password"
                minlength="6"
                required
                class="w-full rounded-lg border border-[var(--border)] bg-[var(--bg-primary)] px-3 py-2 text-[var(--text-primary)] focus:border-[var(--accent)] focus:outline-none"
              />
            </div>

            <div v-if="passwordError" class="rounded-md border border-red-400 bg-red-100 px-3 py-2 text-sm text-red-700">
              {{ passwordError }}
            </div>
            <div v-if="passwordSuccess" class="rounded-md border border-emerald-400 bg-emerald-100 px-3 py-2 text-sm text-emerald-800">
              {{ passwordSuccess }}
            </div>

            <div class="flex items-center justify-end gap-3">
              <button
                type="button"
                class="rounded-lg border border-[var(--border)] px-4 py-2 text-sm font-semibold text-[var(--text-secondary)] transition hover:bg-[var(--bg-secondary)]"
                @click="closePasswordModal"
              >
                Cancel
              </button>
              <button
                type="submit"
                class="rounded-lg bg-[var(--accent)] px-4 py-2 text-sm font-semibold text-white transition hover:bg-[var(--accent)]/90 disabled:opacity-50"
                :disabled="passwordSaving"
              >
                {{ passwordSaving ? 'Updating…' : 'Update Password' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </transition>
  </section>

  <section class="mt-12 space-y-4">
    <button
      type="button"
      class="flex w-full items-center justify-between rounded-xl border border-[var(--border)] bg-[var(--bg-secondary)]/40 px-4 py-3 text-left text-[var(--text-primary)] transition hover:bg-[var(--bg-secondary)]/70"
      @click="toggleIntegrations"
    >
      <div>
        <h2 class="text-lg font-semibold">Telegram Integrations</h2>
        <p class="text-sm text-[var(--text-secondary)]">Generate pairing codes and manage ingest tokens for the Telegram bot.</p>
      </div>
      <i :class="['fa-solid', showIntegrations ? 'fa-chevron-up' : 'fa-chevron-down']"></i>
    </button>

    <transition name="fade">
      <div v-if="showIntegrations" class="rounded-xl border border-[var(--border)] bg-[var(--bg-primary)]/60 p-4">
        <TelegramIntegrationsPanel />
      </div>
    </transition>
  </section>
</template>

<script setup>
import { reactive, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import TelegramIntegrationsPanel from '../components/TelegramIntegrationsPanel.vue';
import { apiFetch, clearAuthToken } from '../lib/api';
import state, { resetState, setUser } from '../stores/appState';

const profileSaving = ref(false);
const profileError = ref('');
const profileSuccess = ref('');

const passwordSaving = ref(false);
const passwordError = ref('');
const passwordSuccess = ref('');
const showPasswordModal = ref(false);
const showIntegrations = ref(false);

const profileForm = reactive({
  email: state.user?.email || '',
  firstName: state.user?.firstName || '',
  lastName: state.user?.lastName || '',
});

const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: '',
});

const router = useRouter();

watch(
  () => state.user,
  (user) => {
    if (user) {
      profileForm.email = user.email || '';
      profileForm.firstName = user.firstName || '';
      profileForm.lastName = user.lastName || '';
    }
  },
  { immediate: true }
);

async function saveProfile() {
  profileSaving.value = true;
  profileError.value = '';
  profileSuccess.value = '';
  try {
    const originalEmail = state.user?.email || '';
    const response = await apiFetch('/api/v1/user/profile', {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(profileForm),
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data?.error || 'Failed to update profile');
    }
    setUser(data);
    profileSuccess.value = 'Profile updated successfully';

    if (originalEmail && data.email && originalEmail !== data.email) {
      clearAuthToken();
      resetState();
      router.push({ path: '/login' });
    }
  } catch (error) {
    console.error('Failed to update profile', error);
    profileError.value = error.message || 'Failed to update profile';
  } finally {
    profileSaving.value = false;
  }
}

async function changePassword() {
  passwordSaving.value = true;
  passwordError.value = '';
  passwordSuccess.value = '';
  try {
    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      throw new Error('New passwords do not match');
    }
    const response = await apiFetch('/api/v1/user/update_password', {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        currentPassword: passwordForm.currentPassword,
        newPassword: passwordForm.newPassword,
      }),
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data?.error || 'Failed to update password');
    }
    passwordSuccess.value = 'Password updated successfully';
    passwordForm.currentPassword = '';
   passwordForm.newPassword = '';
   passwordForm.confirmPassword = '';
   closePasswordModal();

    clearAuthToken();
    resetState();
    router.push({ path: '/login' });
  } catch (error) {
    console.error('Failed to update password', error);
    passwordError.value = error.message || 'Failed to update password';
  } finally {
    passwordSaving.value = false;
  }
}

function openPasswordModal() {
  passwordError.value = '';
  passwordSuccess.value = '';
  passwordForm.currentPassword = '';
  passwordForm.newPassword = '';
  passwordForm.confirmPassword = '';
  showPasswordModal.value = true;
}

function closePasswordModal() {
  showPasswordModal.value = false;
}

function toggleIntegrations() {
  showIntegrations.value = !showIntegrations.value;
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
