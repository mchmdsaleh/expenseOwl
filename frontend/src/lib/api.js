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
  const response = await fetch(url, opts);
  if (response.status === 401) {
    clearAuthToken();
    if (typeof window !== 'undefined') {
      const target = encodeURIComponent(window.location.pathname + window.location.search);
      window.location.href = `/login?redirect=${target}`;
    }
    throw new Error('Unauthorized');
  }
  return response;
}
