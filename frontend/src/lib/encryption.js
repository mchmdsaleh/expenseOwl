import { CompactEncrypt, compactDecrypt, importJWK } from 'jose';
import { getCipher } from './cipher';

let cachedCipher = null;
let cachedKeyPromise = null;

function base64UrlEncode(bytes) {
  let binary = '';
  const len = bytes.byteLength;
  for (let i = 0; i < len; i += 1) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/u, '');
}

async function deriveKey(cipher) {
  if (!cipher || typeof cipher !== 'string') return null;
  if (typeof crypto === 'undefined' || !crypto.subtle) {
    throw new Error('WebCrypto is not available in this environment');
  }
  const encoder = new TextEncoder();
  const digestBuffer = await crypto.subtle.digest('SHA-256', encoder.encode(cipher));
  const digestBytes = new Uint8Array(digestBuffer);
  const jwk = {
    kty: 'oct',
    k: base64UrlEncode(digestBytes),
    alg: 'A256KW',
    use: 'enc',
  };
  return importJWK(jwk, 'A256KW');
}

async function getCryptoKey() {
  const cipher = getCipher();
  if (!cipher) return null;
  if (cachedCipher === cipher && cachedKeyPromise) {
    return cachedKeyPromise;
  }
  cachedCipher = cipher;
  cachedKeyPromise = deriveKey(cipher);
  return cachedKeyPromise;
}

export function resetEncryptionCache() {
  cachedCipher = null;
  cachedKeyPromise = null;
}

function sanitizePayload(payload) {
  if (!payload || typeof payload !== 'object') return payload;
  const clone = JSON.parse(JSON.stringify(payload));
  if (clone && typeof clone === 'object' && 'blob' in clone) {
    delete clone.blob;
  }
  return clone;
}

export async function encryptPayload(payload) {
  const cryptoKey = await getCryptoKey();
  if (!cryptoKey) return null;
  const sanitized = sanitizePayload(payload);
  const plaintext = new TextEncoder().encode(JSON.stringify(sanitized ?? {}));
  const jwe = await new CompactEncrypt(plaintext)
    .setProtectedHeader({ alg: 'A256KW', enc: 'A128CBC-HS256', cty: 'json' })
    .encrypt(cryptoKey);
  return jwe;
}

export async function decryptPayload(blob) {
  const cryptoKey = await getCryptoKey();
  if (!cryptoKey || !blob) return null;
  const { plaintext } = await compactDecrypt(blob, cryptoKey);
  const json = new TextDecoder().decode(plaintext);
  return JSON.parse(json);
}
