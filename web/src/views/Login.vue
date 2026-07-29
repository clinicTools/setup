<script setup lang="ts">
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { Server, Loader2, Sun, Moon } from "lucide-vue-next";
import { useAuthStore } from "@/stores/auth";
import { useTheme } from "@/composables/useTheme";
import { ApiError } from "@/lib/api";
import LanguageSwitcher from "@/components/LanguageSwitcher.vue";
import Button from "@/components/ui/Button.vue";

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();
const { theme, toggle } = useTheme();
const { t } = useI18n();

const username = ref("");
const password = ref("");
const error = ref("");
const loading = ref(false);

async function submit(): Promise<void> {
  error.value = "";
  loading.value = true;
  try {
    await auth.login(username.value, password.value);
    // Open-Redirect verhindern: nur interne, relative Pfade zulassen.
    const target = (route.query.redirect as string) || "/";
    const safe = target.startsWith("/") && !target.startsWith("//") ? target : "/";
    router.push(safe);
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : t("login.failed");
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <!-- Windows-11-Anmeldebildschirm: immersiver „Bloom"-Hintergrund + Acrylic-Karte -->
  <div class="login-bg relative flex min-h-full items-center justify-center overflow-hidden p-4 text-white">
    <!-- dekorative, weich gezeichnete Farb-Blobs -->
    <div class="bloom bloom-a"></div>
    <div class="bloom bloom-b"></div>
    <div class="bloom bloom-c"></div>

    <div class="absolute right-4 top-4 z-10 flex items-center gap-1 text-white [&_button:hover]:bg-white/15">
      <LanguageSwitcher />
      <Button variant="ghost" size="icon" class="text-white hover:bg-white/15" @click="toggle">
        <Sun v-if="theme === 'dark'" class="h-4 w-4" />
        <Moon v-else class="h-4 w-4" />
      </Button>
    </div>

    <div class="relative z-10 w-full max-w-[400px]">
      <!-- Konto-Kopf (Win11-Sign-in: großer Avatar, Name, Untertitel) -->
      <div class="mb-7 flex flex-col items-center text-center">
        <div
          class="mb-4 flex h-24 w-24 items-center justify-center rounded-full bg-white/10 ring-1 ring-white/25 backdrop-blur-md"
        >
          <Server class="h-11 w-11 text-white/90" />
        </div>
        <h1 class="text-2xl font-semibold tracking-tight">Debian Admin</h1>
        <p class="mt-1 text-sm text-white/70">{{ t("login.subtitle") }}</p>
      </div>

      <!-- Acrylic-Anmeldekarte -->
      <form
        class="space-y-3.5 rounded-2xl border border-white/15 bg-white/10 p-6 shadow-2xl backdrop-blur-2xl"
        @submit.prevent="submit"
      >
        <div class="space-y-1.5">
          <label for="username" class="text-xs font-medium text-white/80">{{ t("login.username") }}</label>
          <input
            id="username"
            v-model="username"
            placeholder="root"
            autocomplete="username"
            class="h-11 w-full rounded-lg border border-white/20 bg-white/10 px-3.5 text-sm text-white placeholder:text-white/50 focus:border-white/40 focus:outline-none focus:ring-2 focus:ring-white/30"
          />
        </div>
        <div class="space-y-1.5">
          <label for="password" class="text-xs font-medium text-white/80">{{ t("login.password") }}</label>
          <input
            id="password"
            v-model="password"
            type="password"
            autocomplete="current-password"
            @keyup.enter="submit"
            class="h-11 w-full rounded-lg border border-white/20 bg-white/10 px-3.5 text-sm text-white placeholder:text-white/50 focus:border-white/40 focus:outline-none focus:ring-2 focus:ring-white/30"
          />
        </div>

        <p v-if="error" class="rounded-lg bg-red-500/25 px-3 py-2 text-sm text-red-50 ring-1 ring-red-300/30">
          {{ error }}
        </p>

        <button
          type="submit"
          :disabled="loading || !username || !password"
          class="flex h-11 w-full items-center justify-center gap-2 rounded-lg bg-white/90 text-sm font-semibold text-zinc-900 shadow-lg transition-colors hover:bg-white disabled:cursor-not-allowed disabled:opacity-50"
        >
          <Loader2 v-if="loading" class="h-4 w-4 animate-spin" />
          <span>{{ loading ? t("login.submitting") : t("login.submit") }}</span>
        </button>
      </form>

      <p class="mt-5 text-center text-xs text-white/55">
        {{ t("login.pamNote") }}
      </p>
    </div>
  </div>
</template>

<style scoped>
/* „Bloom"-artiger Hintergrund (an Windows 11 angelehnt). */
.login-bg {
  background:
    radial-gradient(120% 120% at 50% 0%, #1b2a6b 0%, #122050 35%, #0a1230 70%, #060a1c 100%);
}
.bloom {
  position: absolute;
  border-radius: 9999px;
  filter: blur(70px);
  opacity: 0.55;
  pointer-events: none;
}
.bloom-a {
  width: 38rem;
  height: 38rem;
  left: -8rem;
  top: -10rem;
  background: radial-gradient(circle, #4f7cff 0%, transparent 70%);
}
.bloom-b {
  width: 34rem;
  height: 34rem;
  right: -10rem;
  bottom: -8rem;
  background: radial-gradient(circle, #9b5cff 0%, transparent 70%);
}
.bloom-c {
  width: 26rem;
  height: 26rem;
  right: 14%;
  top: 6%;
  opacity: 0.4;
  background: radial-gradient(circle, #21d4c4 0%, transparent 70%);
}
</style>
