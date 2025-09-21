<template>
  <section class="space-y-8">
    <header class="space-y-1">
      <h1 class="text-2xl font-semibold text-[var(--text-primary)]">Integrations</h1>
      <p class="text-sm text-[var(--text-secondary)]">
        Manage Telegram webhooks for automated expense capture. Generate a pairing code, link it from Telegram using <code>/link &lt;code&gt;</code>, then share the ingest token with your n8n workflow.
      </p>
    </header>

    <div v-if="errorMessage" class="rounded-md border border-red-400 bg-red-100 px-3 py-2 text-sm text-red-700">
      {{ errorMessage }}
    </div>
    <div v-if="successMessage" class="rounded-md border border-emerald-500/40 bg-emerald-500/10 px-3 py-2 text-sm text-[var(--text-primary)]">
      {{ successMessage }}
    </div>

    <form class="space-y-4 rounded-xl border border-[var(--border)] bg-[var(--bg-secondary)]/60 p-6 shadow-sm" @submit.prevent="handleCreate">
      <div class="flex flex-wrap items-end gap-3 md:gap-4">
        <div class="grow space-y-2">
          <label class="block text-sm font-medium text-[var(--text-secondary)]" for="integration-label">Label</label>
          <input
            id="integration-label"
            v-model="newLabel"
            type="text"
            maxlength="100"
            placeholder="e.g. Personal Telegram"
            class="w-full rounded-lg border border-[var(--border)] bg-[var(--bg-primary)] px-3 py-2 text-[var(--text-primary)] placeholder:text-[var(--text-secondary)] focus:border-[var(--accent)] focus:outline-none"
          />
        </div>
        <button
          type="submit"
          class="w-full rounded-lg border border-[var(--accent)] bg-[var(--accent)] px-4 py-2 text-sm font-semibold text-white transition hover:bg-[var(--accent)]/90 disabled:opacity-50 md:w-auto md:px-6 md:py-2.5"
          :disabled="creating"
        >
          {{ creating ? 'Generating…' : 'Generate Link' }}
        </button>
        <button
          type="button"
          class="w-full rounded-lg border border-[var(--border)] bg-[var(--bg-primary)] px-4 py-2 text-sm text-[var(--text-secondary)] transition hover:bg-[var(--bg-secondary)]/80 md:w-auto md:px-6 md:py-2.5"
          :disabled="loading"
          @click="fetchLinks"
        >
          {{ loading ? 'Refreshing…' : 'Refresh' }}
        </button>
      </div>
      <p class="mt-3 text-sm text-[var(--text-secondary)]">
        Each link issues a one-time pairing code and a long-lived ingest token. Pair the chat first, then copy the ingest token into your n8n credentials.
      </p>
    </form>

    <div v-if="secretLink" class="space-y-3 rounded-xl border border-[var(--border)] bg-[var(--bg-secondary)]/70 p-4 text-sm text-[var(--text-primary)] shadow-lg">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <strong class="text-[var(--text-primary)]">New link created &mdash; copy these secrets now.</strong>
        <button
          type="button"
          class="text-xs font-semibold uppercase tracking-wide text-[var(--text-secondary)] transition hover:text-[var(--text-primary)]"
          @click="secretLink = null"
        >
          Dismiss
        </button>
      </div>
      <div class="grid gap-3 md:grid-cols-2">
        <div class="space-y-1">
          <label class="block text-xs font-semibold uppercase text-[var(--text-secondary)]">Link Code</label>
          <div class="rounded-lg border border-[var(--border)] bg-[var(--bg-primary)] px-3 py-2 font-mono text-sm tracking-wide">
            {{ secretLink.linkCode }}
          </div>
          <p class="text-xs text-[var(--text-secondary)]">Share this code with the Telegram user and have them send <code>/link {{ secretLink.linkCode }}</code>.</p>
        </div>
        <div class="space-y-1">
          <label class="block text-xs font-semibold uppercase text-[var(--text-secondary)]">Ingest Token</label>
          <div class="overflow-x-auto rounded-lg border border-[var(--border)] bg-[var(--bg-primary)] px-3 py-2 font-mono text-xs tracking-wide">
            {{ secretLink.ingestToken }}
          </div>
          <p class="text-xs text-[var(--text-secondary)]">Store this securely in n8n (HTTP Bearer credential). It will not be shown again.</p>
        </div>
      </div>
    </div>

    <div class="overflow-x-auto rounded-lg border border-[var(--border)]">
      <table class="min-w-full divide-y divide-[var(--border)] text-sm">
        <thead class="bg-[var(--bg-secondary)] text-[var(--text-secondary)]">
          <tr>
            <th class="px-4 py-3 text-left font-medium">Label</th>
            <th class="px-4 py-3 text-left font-medium">Status</th>
            <th class="px-4 py-3 text-left font-medium">Chat</th>
            <th class="px-4 py-3 text-left font-medium">Created</th>
            <th class="px-4 py-3 text-left font-medium">Last Seen</th>
            <th class="px-4 py-3 text-left font-medium">Linked At</th>
            <th class="px-4 py-3 text-right font-medium">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[var(--border)] bg-[var(--bg-primary)] text-[var(--text-primary)]">
          <tr v-if="!loading && !hasLinks">
            <td class="px-4 py-6 text-center text-sm text-[var(--text-secondary)]" colspan="7">
              No integrations yet. Generate a link to get started.
            </td>
          </tr>
          <tr v-for="link in links" :key="link.id">
            <td class="px-4 py-3 font-medium">{{ link.label }}</td>
            <td class="px-4 py-3">
              <span
                :class="[
                  'inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold uppercase tracking-wide',
                  statusPillClass(link)
                ]"
              >
                {{ statusLabel(link) }}
              </span>
            </td>
            <td class="px-4 py-3">
              <div v-if="link.chatId" class="space-y-0.5">
                <div>ID: {{ link.chatId }}</div>
                <div v-if="link.telegramHandle" class="text-xs text-[var(--text-secondary)]">@{{ link.telegramHandle }}</div>
              </div>
              <span v-else class="text-sm text-[var(--text-secondary)]">Waiting for /link</span>
            </td>
            <td class="px-4 py-3">{{ formatDate(link.createdAt) }}</td>
            <td class="px-4 py-3">{{ formatDate(link.lastSeenAt) }}</td>
            <td class="px-4 py-3">{{ formatDate(link.linkedAt) }}</td>
            <td class="px-4 py-3 text-right">
              <button
                v-if="!link.revokedAt"
                type="button"
                class="rounded-lg border border-red-300 px-3 py-1 text-xs font-semibold text-red-600 hover:bg-red-50 disabled:opacity-50"
                :disabled="revokingId === link.id"
                @click="handleRevoke(link)"
              >
                {{ revokingId === link.id ? 'Revoking…' : 'Revoke' }}
              </button>
              <span v-else class="text-xs uppercase text-[var(--text-secondary)]">Revoked</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import { createTelegramLink, listTelegramLinks, revokeTelegramLink } from '../lib/api';

const links = ref([]);
const loading = ref(false);
const creating = ref(false);
const revokingId = ref('');
const errorMessage = ref('');
const successMessage = ref('');
const newLabel = ref('');

const secretLink = ref(null);

const hasLinks = computed(() => Array.isArray(links.value) && links.value.length > 0);

function statusLabel(link) {
  if (link.revokedAt) return 'Revoked';
  if (link.linkedAt) return 'Active';
  return 'Awaiting link';
}

function statusPillClass(link) {
  if (link.revokedAt) return 'bg-red-100 text-red-700';
  if (link.linkedAt) return 'bg-emerald-100 text-emerald-700';
  return 'bg-amber-100 text-amber-700';
}

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

async function fetchLinks() {
  loading.value = true;
  errorMessage.value = '';
  try {
    const data = await listTelegramLinks();
    links.value = Array.isArray(data) ? data : [];
  } catch (error) {
    console.error('Failed to load telegram links', error);
    errorMessage.value = error.message || 'Failed to load telegram links';
  } finally {
    loading.value = false;
  }
}

async function handleCreate() {
  if (creating.value) return;
  creating.value = true;
  errorMessage.value = '';
  successMessage.value = '';
  secretLink.value = null;
  try {
    const label = newLabel.value.trim() || 'Telegram';
    const data = await createTelegramLink({ label });
    secretLink.value = {
      id: data.id,
      label: data.label,
      linkCode: data.linkCode,
      ingestToken: data.ingestToken,
    };
    successMessage.value = 'Link generated successfully. Copy the details below before navigating away.';
    newLabel.value = '';
    await fetchLinks();
  } catch (error) {
    console.error('Failed to create telegram link', error);
    errorMessage.value = error.message || 'Failed to create telegram link';
  } finally {
    creating.value = false;
  }
}

async function handleRevoke(link) {
  if (!window.confirm('Revoke this Telegram link? Ingest tokens using it will stop working.')) {
    return;
  }
  revokingId.value = link.id;
  errorMessage.value = '';
  successMessage.value = '';
  try {
    await revokeTelegramLink(link.id);
    successMessage.value = 'Link revoked successfully.';
    if (secretLink.value?.id === link.id) {
      secretLink.value = null;
    }
    await fetchLinks();
  } catch (error) {
    console.error('Failed to revoke telegram link', error);
    errorMessage.value = error.message || 'Failed to revoke telegram link';
  } finally {
    revokingId.value = '';
  }
}

onMounted(fetchLinks);
</script>
