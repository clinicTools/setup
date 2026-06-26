<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import {
  Search,
  Sun,
  Moon,
  LogOut,
  Power,
  Server,
  Menu,
} from "lucide-vue-next";
import { navItems } from "@/nav";
import { useTheme } from "@/composables/useTheme";
import { useToast } from "@/composables/useToast";
import { useAuthStore } from "@/stores/auth";
import { api } from "@/lib/api";
import { cn } from "@/lib/utils";
import Button from "@/components/ui/Button.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import ConnectionStatus from "@/components/ConnectionStatus.vue";
import LanguageSwitcher from "@/components/LanguageSwitcher.vue";

const route = useRoute();
const router = useRouter();
const { theme, toggle } = useTheme();
const { t } = useI18n();
const toast = useToast();
const auth = useAuthStore();

const search = ref("");
const sidebarOpen = ref(false);

// Navigationssuche (Windows-11-Einstellungen: Suchfeld über der Navigation).
const filtered = computed(() => {
  const q = search.value.trim().toLowerCase();
  const items = q
    ? navItems.filter((i) => t(`nav.items.${i.labelKey}`).toLowerCase().includes(q))
    : navItems;
  const groups = new Map<string, typeof navItems>();
  for (const item of items) {
    if (!groups.has(item.groupKey)) groups.set(item.groupKey, []);
    groups.get(item.groupKey)!.push(item);
  }
  return groups;
});

const powerDialog = ref<null | "reboot" | "poweroff">(null);

async function confirmPower(): Promise<void> {
  const action = powerDialog.value;
  powerDialog.value = null;
  if (!action) return;
  try {
    await api.power(action);
    toast.success(action === "reboot" ? "Neustart eingeleitet" : "Herunterfahren eingeleitet");
  } catch (e) {
    toast.error((e as Error).message);
  }
}

async function doLogout(): Promise<void> {
  await auth.logout();
  router.push("/login");
}

const initials = computed(() => {
  const name = auth.user?.fullName || auth.user?.username || "?";
  return name.slice(0, 2).toUpperCase();
});
</script>

<template>
  <div class="flex h-full overflow-hidden">
    <!-- Navigationsleiste (Acrylic über dem Mica-Hintergrund) -->
    <aside
      :class="
        cn(
          'fixed inset-y-0 left-0 z-40 flex w-64 flex-col border-r border-border/60 bg-sidebar backdrop-blur-2xl transition-transform md:static md:translate-x-0',
          sidebarOpen ? 'translate-x-0' : '-translate-x-full',
        )
      "
    >
      <!-- Marke -->
      <div class="flex h-14 items-center gap-2.5 px-4">
        <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-primary to-primary/70 text-primary-foreground shadow-sm">
          <Server class="h-[18px] w-[18px]" />
        </div>
        <div class="leading-tight">
          <div class="text-sm font-semibold">Debian Admin</div>
          <div class="text-[11px] text-muted-foreground">{{ t("shell.subtitle") }}</div>
        </div>
      </div>

      <!-- Konto-Karte (Windows-11-Muster: Konto oben in der Navigation) -->
      <div class="px-3 pb-2">
        <div class="flex items-center gap-3 rounded-xl border border-border/70 bg-card/70 px-3 py-2.5 shadow-sm">
          <div class="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-primary to-primary/60 text-xs font-semibold text-primary-foreground">
            {{ initials }}
          </div>
          <div class="min-w-0 flex-1 leading-tight">
            <div class="truncate text-sm font-medium">{{ auth.user?.fullName || auth.user?.username }}</div>
            <div class="truncate text-[11px] text-muted-foreground">
              {{ auth.user?.admin ? t("shell.administrator") : t("shell.user") }}
            </div>
          </div>
          <Button variant="ghost" size="icon" class="h-8 w-8" :title="t('shell.logout')" @click="doLogout">
            <LogOut class="h-4 w-4" />
          </Button>
        </div>
      </div>

      <!-- Suche -->
      <div class="px-3 pb-2">
        <div class="relative">
          <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            v-model="search"
            type="search"
            :placeholder="t('shell.searchSetting')"
            class="h-9 w-full rounded-lg border border-input bg-card/80 pl-9 pr-3 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>
      </div>

      <nav class="scrollbar-thin flex-1 space-y-3 overflow-y-auto px-2 py-2">
        <div v-for="[group, items] in filtered" :key="group">
          <div class="px-3 pb-1 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground/80">
            {{ t(`nav.groups.${group}`) }}
          </div>
          <RouterLink
            v-for="item in items"
            :key="item.to"
            :to="item.to"
            class="group relative flex items-center gap-3 rounded-lg px-3 py-[7px] text-[13px] transition-colors"
            :class="
              route.path === item.to
                ? 'bg-sidebar-accent font-semibold text-foreground shadow-sm'
                : 'text-foreground/75 hover:bg-accent/50 hover:text-foreground'
            "
            @click="sidebarOpen = false"
          >
            <!-- Akzent-Balken des aktiven Eintrags (Windows-11-Muster) -->
            <span
              v-if="route.path === item.to"
              class="absolute left-0 top-1/2 h-4 w-[3px] -translate-y-1/2 rounded-r-full bg-primary"
            />
            <component
              :is="item.icon"
              class="h-[18px] w-[18px] shrink-0"
              :class="route.path === item.to ? 'text-primary' : 'text-foreground/60'"
            />
            <span class="truncate">{{ t(`nav.items.${item.labelKey}`) }}</span>
          </RouterLink>
        </div>
      </nav>
    </aside>

    <!-- Backdrop für mobile Navigation -->
    <div
      v-if="sidebarOpen"
      class="fixed inset-0 z-30 bg-black/30 md:hidden"
      @click="sidebarOpen = false"
    />

    <!-- Inhalt -->
    <div class="flex min-w-0 flex-1 flex-col">
      <header class="flex h-12 shrink-0 items-center justify-end gap-1 px-4">
        <Button variant="ghost" size="icon" class="mr-auto md:hidden" @click="sidebarOpen = true">
          <Menu class="h-5 w-5" />
        </Button>
        <ConnectionStatus />
        <LanguageSwitcher />
        <Button variant="ghost" size="icon" :title="theme === 'dark' ? t('shell.lightTheme') : t('shell.darkTheme')" @click="toggle">
          <Sun v-if="theme === 'dark'" class="h-4 w-4" />
          <Moon v-else class="h-4 w-4" />
        </Button>
        <Button
          v-if="auth.user?.admin"
          variant="ghost"
          size="icon"
          :title="t('shell.power')"
          @click="powerDialog = 'reboot'"
        >
          <Power class="h-4 w-4" />
        </Button>
      </header>

      <main class="scrollbar-thin flex-1 overflow-y-auto">
        <div class="mx-auto max-w-[1080px] px-5 pb-10 pt-2 md:px-10">
          <RouterView />
        </div>
      </main>
    </div>

    <ConfirmDialog
      :open="powerDialog !== null"
      :title="t('shell.rebootTitle')"
      :message="t('shell.rebootMessage')"
      :confirm-label="t('shell.rebootConfirm')"
      destructive
      @confirm="confirmPower"
      @cancel="powerDialog = null"
    />
  </div>
</template>
