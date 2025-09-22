const STORAGE_KEY = 'expenseowl_cipher';
let cachedCipher = null;

function isBrowser() {
  return typeof window !== 'undefined' && typeof window.localStorage !== 'undefined';
}

function loadFromStorage() {
  if (cachedCipher !== null) return;
  if (!isBrowser()) return;
  const stored = window.localStorage.getItem(STORAGE_KEY);
  cachedCipher = stored && stored.trim() !== '' ? stored.trim() : null;
}

export function getCipher() {
  loadFromStorage();
  return cachedCipher;
}

export function setCipher(cipher) {
  const value = typeof cipher === 'string' ? cipher.trim() : '';
  cachedCipher = value || null;
  if (!isBrowser()) return;
  if (cachedCipher) {
    window.localStorage.setItem(STORAGE_KEY, cachedCipher);
  } else {
    window.localStorage.removeItem(STORAGE_KEY);
  }
}

export function clearCipher() {
  cachedCipher = null;
  if (isBrowser()) {
    window.localStorage.removeItem(STORAGE_KEY);
  }
}
