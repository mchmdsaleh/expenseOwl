<template>
  <div class="flex min-h-screen items-center justify-center bg-[var(--bg-primary)] px-6 relative overflow-hidden">
    <!-- Animated background accents -->
    <div class="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-[var(--accent)] opacity-10 rounded-full blur-[120px]"></div>
    <div class="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-violet-600 opacity-10 rounded-full blur-[120px]"></div>

    <div class="w-full max-w-md glass-card rounded-[20px] md:rounded-[40px] p-6 md:p-10 shadow-2xl relative z-10 animate-in zoom-in duration-700">
      <div class="flex flex-col items-center mb-10">
         <div class="h-16 w-16 bg-gradient-to-br from-indigo-500 to-violet-600 text-white rounded-3xl flex items-center justify-center shadow-2xl shadow-indigo-500/20 mb-6 scale-110">
            <i class="fa-solid fa-owl text-3xl"></i>
         </div>
         <h1 class="text-3xl font-black tracking-tighter text-[var(--text-primary)]">
           {{ isLogin ? 'Welcome Back' : 'Get Started' }}
         </h1>
         <p class="text-xs font-bold text-[var(--text-secondary)] uppercase tracking-[0.2em] mt-2">{{ isLogin ? 'Sign in to continue' : 'Create your free account' }}</p>
      </div>

      <form class="space-y-6" @submit.prevent="handleSubmit">
        <div class="space-y-2">
          <label class="text-[10px] font-black uppercase tracking-widest text-[var(--text-secondary)] ml-1" for="email">Email Address</label>
          <input
            id="email"
            v-model="form.email"
            type="email"
            required
            class="input-modern w-full"
            placeholder="name@example.com"
          />
        </div>
        <div class="grid gap-6" v-if="!isLogin">
          <div class="space-y-2">
            <label class="text-[10px] font-black uppercase tracking-widest text-[var(--text-secondary)] ml-1" for="firstName">First Name</label>
            <input
              id="firstName"
              v-model="form.firstName"
              type="text"
              required
              class="input-modern w-full"
              placeholder="John"
            />
          </div>
          <div class="space-y-2">
            <label class="text-[10px] font-black uppercase tracking-widest text-[var(--text-secondary)] ml-1" for="lastName">Last Name</label>
            <input
              id="lastName"
              v-model="form.lastName"
              type="text"
              required
              class="input-modern w-full"
              placeholder="Doe"
            />
          </div>
        </div>
        <div class="space-y-2">
          <label class="text-[10px] font-black uppercase tracking-widest text-[var(--text-secondary)] ml-1" for="password">Security Password</label>
          <input
            id="password"
            v-model="form.password"
            type="password"
            required
            minlength="6"
            class="input-modern w-full"
            placeholder="••••••••"
          />
        </div>
        
        <transition name="fade">
          <div v-if="errorMessage" class="p-4 rounded-2xl bg-rose-500/10 text-rose-500 text-xs font-bold border border-rose-500/20 text-center">
            {{ errorMessage }}
          </div>
        </transition>

        <button
          type="submit"
          class="btn-primary w-full h-14 text-lg shadow-indigo-500/20 shadow-2xl mt-4"
          :disabled="submitting"
        >
          <span v-if="submitting" class="flex items-center gap-2">
             <i class="fa-solid fa-circle-notch animate-spin"></i>
             Authenticating...
          </span>
          <span v-else>{{ isLogin ? 'Sign In' : 'Create Account' }}</span>
        </button>
      </form>

      <div class="mt-10 pt-10 border-t border-[var(--border)] text-center">
        <p class="text-sm font-medium text-[var(--text-secondary)]">
          {{ isLogin ? "New to ExpenseOwl?" : 'Have an account already?' }}
          <button
            type="button"
            class="font-black text-[var(--accent)] hover:underline ml-1"
            @click="toggleMode"
          >
            {{ isLogin ? 'Join now' : 'Sign in here' }}
          </button>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import md5 from 'blueimp-md5';
import { apiFetch, setAuthToken } from '../lib/api';
import { loadInitialData } from '../stores/appState';
import { setCipher } from '../lib/cipher';
import { resetEncryptionCache } from '../lib/encryption';

const route = useRoute();
const router = useRouter();
const mode = ref('login');
const submitting = ref(false);
const errorMessage = ref('');

const form = ref({
  email: '',
  password: '',
  firstName: '',
  lastName: '',
});

const isLogin = computed(() => mode.value === 'login');

function toggleMode() {
  mode.value = isLogin.value ? 'signup' : 'login';
  errorMessage.value = '';
}

async function handleSubmit() {
  submitting.value = true;
  errorMessage.value = '';
  try {
    const endpoint = isLogin.value ? '/api/v1/user/login' : '/api/v1/user/signup';
    const payload = {
      email: form.value.email,
      password: form.value.password,
    };
    if (!isLogin.value) {
      payload.firstName = form.value.firstName;
      payload.lastName = form.value.lastName;
    }
    const response = await apiFetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    });
    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      errorMessage.value = data?.error || 'Authentication failed';
      return;
    }
    const data = await response.json();
    setAuthToken(data.token);
    const cipher = md5(form.value.password || '');
    setCipher(cipher);
    resetEncryptionCache();
    await loadInitialData();
    const target = typeof route.query.redirect === 'string' && route.query.redirect ? route.query.redirect : '/';
    router.push(target);
  } catch (error) {
    console.error('Authentication error', error);
    errorMessage.value = 'Something went wrong. Please try again.';
  } finally {
    submitting.value = false;
  }
}
</script>

<style scoped>
@keyframes zoomIn {
  from { opacity: 0; transform: scale(0.9); }
  to { opacity: 1; transform: scale(1); }
}

.animate-in {
  animation: zoomIn 0.6s cubic-bezier(0.34, 1.56, 0.64, 1) fill-mode-both;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>
