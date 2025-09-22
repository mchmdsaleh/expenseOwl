import { getCipher, clearCipher } from './cipher';
import { resetEncryptionCache } from './encryption';

const TOKEN_KEY = 'expenseowl_token';

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
  const response = await fetch(url, opts);
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
