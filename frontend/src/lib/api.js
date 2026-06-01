import { getCipher, clearCipher } from './cipher';
import { resetEncryptionCache } from './encryption';

const TOKEN_KEY = 'expenseowl_token';

const explicitApiBase = import.meta.env?.VITE_API_BASE_URL || '';
let detectedApiBase = '';

if (typeof window !== 'undefined' && !explicitApiBase) {
  const { protocol, hostname, port } = window.location;
  if (port === '5173') {
    detectedApiBase = `${protocol}//${hostname}:9080`;
  }
}

const API_BASE = explicitApiBase || detectedApiBase;

export function getAuthToken() {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function setAuthToken(token) {
  if (typeof window === 'undefined') return;
  if (token) {
    localStorage.setItem(TOKEN_KEY, token);
  } else {
    localStorage.removeItem(TOKEN_KEY);
  }
}

export function clearAuthToken() {
  if (typeof window === 'undefined') return;
  localStorage.removeItem(TOKEN_KEY);
}

export async function apiFetch(url, options = {}) {
  const opts = {
    ...options,
  };
  opts.headers = {
    ...(options.headers || {}),
  };
  if (!opts.headers['X-Requested-With']) {
    opts.headers['X-Requested-With'] = 'ExpenseOwl';
  }
  const token = getAuthToken();
  if (token && !opts.headers.Authorization) {
    opts.headers.Authorization = `Bearer ${token}`;
  }
  const cipher = getCipher();
  if (cipher && !opts.headers['X-Encryption-Key']) {
    opts.headers['X-Encryption-Key'] = cipher;
  }
  let targetUrl = url;
  if (!/^https?:\/\//i.test(url) && API_BASE) {
    targetUrl = url.startsWith('/') ? `${API_BASE}${url}` : `${API_BASE}/${url}`;
  }
  const response = await fetch(targetUrl, opts);
  if (response.status === 401) {
    clearAuthToken();
    clearCipher();
    resetEncryptionCache();
    if (typeof window !== 'undefined') {
      const target = encodeURIComponent(window.location.pathname + window.location.search);
      window.location.href = `/login?redirect=${target}`;
    }
    throw new Error('Unauthorized');
  }
  return response;
}

export async function listTelegramLinks() {
  const response = await apiFetch('/api/v1/integrations/telegram/links');
  if (!response.ok) {
    throw new Error('Failed to load Telegram links');
  }
  return response.json();
}

export async function createTelegramLink({ label }) {
  const response = await apiFetch('/api/v1/integrations/telegram/links', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ label }),
  });
  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data?.error || 'Failed to create Telegram link');
  }
  return response.json();
}

export async function revokeTelegramLink(id) {
  const response = await apiFetch(`/api/v1/integrations/telegram/links?id=${encodeURIComponent(id)}`, {
    method: 'DELETE',
  });
  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data?.error || 'Failed to revoke Telegram link');
  }
}

// AI Chatbot & Configuration
export async function getAIContext() {
  const response = await apiFetch('/api/v1/ai/context');
  if (!response.ok) throw new Error('Failed to load AI context');
  const data = await response.json();
  return data.context;
}

export async function updateAIContext(context) {
  const response = await apiFetch('/api/v1/ai/context/edit', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ context }),
  });
  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data?.error || 'Failed to update AI context');
  }
}

export async function getAIConfig() {
  const response = await apiFetch('/api/v1/ai/config');
  if (!response.ok) throw new Error('Failed to load AI config');
  return response.json();
}

export async function updateAIConfig(config) {
  const response = await apiFetch('/api/v1/ai/config/edit', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config),
  });
  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data?.error || 'Failed to update AI config');
  }
}

export async function sendChatMessage(messages, files = []) {
  const formData = new FormData();
  formData.append('messages', JSON.stringify(messages));
  files.forEach((file) => {
    formData.append('files', file);
  });

  const response = await apiFetch('/api/v1/chat', {
    method: 'POST',
    body: formData,
  });

  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data?.error || 'AI failed to respond');
  }
  return response.json();
}
