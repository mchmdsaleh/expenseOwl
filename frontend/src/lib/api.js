export async function apiFetch(url, options = {}) {
  const opts = {
    ...options,
    credentials: 'include'
  };
  opts.headers = {
    ...(options.headers || {})
  };
  if (!opts.headers['X-Requested-With']) {
    opts.headers['X-Requested-With'] = 'ExpenseOwl';
  }
  const response = await fetch(url, opts);
  if (response.status === 401) {
    window.location.href = '/login';
    throw new Error('Unauthorized');
  }
  return response;
}
