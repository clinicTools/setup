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
import Input from "@/components/ui/Input.vue";

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
  <div class="relative flex min-h-full items-center justify-center bg-background p-4">
    <div class="absolute right-4 top-4 flex items-center gap-1">
      <LanguageSwitcher />
      <Button variant="ghost" size="icon" @click="toggle">
        <Sun v-if="theme === 'dark'" class="h-4 w-4" />
        <Moon v-else class="h-4 w-4" />
      </Button>
    </div>

    <div class="w-full max-w-sm">
      <div class="mb-6 flex flex-col items-center text-center">
        <div class="mb-3 flex h-14 w-14 items-center justify-center rounded-2xl bg-primary text-primary-foreground shadow-lg">
          <Server class="h-7 w-7" />
        </div>
        <h1 class="text-xl font-semibold tracking-tight">Debian Admin</h1>
        <p class="mt-1 text-sm text-muted-foreground">
          {{ t("login.subtitle") }}
        </p>
      </div>

      <form
        class="space-y-4 rounded-xl border border-border bg-card p-6 shadow-sm"
        @submit.prevent="submit"
      >
        <div class="space-y-1.5">
          <label for="username" class="text-sm font-medium">{{ t("login.username") }}</label>
          <Input id="username" v-model="username" placeholder="root" autocomplete="username" />
        </div>
        <div class="space-y-1.5">
          <label for="password" class="text-sm font-medium">{{ t("login.password") }}</label>
          <Input id="password" v-model="password" type="password" autocomplete="current-password" />
        </div>

        <p v-if="error" class="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {{ error }}
        </p>

        <Button
          type="submit"
          variant="primary"
          class="w-full"
          :disabled="loading || !username || !password"
        >
          <Loader2 v-if="loading" class="h-4 w-4 animate-spin" />
          <span>{{ loading ? t("login.submitting") : t("login.submit") }}</span>
        </Button>
      </form>

      <p class="mt-4 text-center text-xs text-muted-foreground">
        {{ t("login.pamNote") }}
      </p>
    </div>
  </div>
</template>
