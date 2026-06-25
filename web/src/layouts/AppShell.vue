<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
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

const route = useRoute();
const router = useRouter();
const { theme, toggle } = useTheme();
const toast = useToast();
const auth = useAuthStore();

const search = ref("");
const sidebarOpen = ref(false);

// Navigationssuche (Windows-11-Einstellungen: Suchfeld über der Navigation).
const filtered = computed(() => {
  const q = search.value.trim().toLowerCase();
  const items = q ? navItems.filter((i) => i.label.toLowerCase().includes(q)) : navItems;
  const groups = new Map<string, typeof navItems>();
  for (const item of items) {
    if (!groups.has(item.group)) groups.set(item.group, []);
    groups.get(item.group)!.push(item);
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
  <div class="flex h-full overflow-hidden bg-background">
    <!-- Navigationsleiste -->
    <aside
      :class="
        cn(
          'fixed inset-y-0 left-0 z-40 flex w-64 flex-col border-r border-border bg-sidebar transition-transform md:static md:translate-x-0',
          sidebarOpen ? 'translate-x-0' : '-translate-x-full',
        )
      "
    >
      <div class="flex h-14 items-center gap-2 px-4">
        <div class="flex h-8 w-8 items-center justify-center rounded-md bg-primary text-primary-foreground">
          <Server class="h-[18px] w-[18px]" />
        </div>
        <div class="leading-tight">
          <div class="text-sm font-semibold">Debian Admin</div>
          <div class="text-[11px] text-muted-foreground">Systemverwaltung</div>
        </div>
      </div>

      <div class="px-3 pb-2">
        <div class="relative">
          <Search class="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            v-model="search"
            type="search"
            placeholder="Einstellung suchen"
            class="h-9 w-full rounded-md border border-input bg-card pl-8 pr-3 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>
      </div>

      <nav class="scrollbar-thin flex-1 space-y-3 overflow-y-auto px-3 py-2">
        <div v-for="[group, items] in filtered" :key="group">
          <div class="px-2 pb-1 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
            {{ group }}
          </div>
          <RouterLink
            v-for="item in items"
            :key="item.to"
            :to="item.to"
            class="group relative flex items-center gap-3 rounded-md px-2.5 py-2 text-sm transition-colors"
            :class="
              route.path === item.to
                ? 'bg-sidebar-accent font-medium text-foreground'
                : 'text-foreground/70 hover:bg-accent/60 hover:text-foreground'
            "
            @click="sidebarOpen = false"
          >
            <!-- Akzent-Pille des aktiven Eintrags (Windows-11-Muster) -->
            <span
              v-if="route.path === item.to"
              class="absolute left-0 top-1/2 h-5 w-1 -translate-y-1/2 rounded-full bg-primary"
            />
            <component :is="item.icon" class="h-[18px] w-[18px] shrink-0" />
            <span class="truncate">{{ item.label }}</span>
          </RouterLink>
        </div>
      </nav>

      <!-- Benutzer + Power -->
      <div class="border-t border-border p-3">
        <div class="flex items-center gap-2">
          <div class="flex h-8 w-8 items-center justify-center rounded-full bg-primary/15 text-xs font-semibold text-primary">
            {{ initials }}
          </div>
          <div class="min-w-0 flex-1 leading-tight">
            <div class="truncate text-sm font-medium">{{ auth.user?.fullName || auth.user?.username }}</div>
            <div class="truncate text-[11px] text-muted-foreground">
              {{ auth.user?.admin ? "Administrator" : "Benutzer" }}
            </div>
          </div>
          <Button variant="ghost" size="icon" title="Abmelden" @click="doLogout">
            <LogOut class="h-4 w-4" />
          </Button>
        </div>
      </div>
    </aside>

    <!-- Backdrop für mobile Navigation -->
    <div
      v-if="sidebarOpen"
      class="fixed inset-0 z-30 bg-black/30 md:hidden"
      @click="sidebarOpen = false"
    />

    <!-- Inhalt -->
    <div class="flex min-w-0 flex-1 flex-col">
      <header class="flex h-14 shrink-0 items-center justify-between gap-2 border-b border-border bg-background/80 px-4 backdrop-blur">
        <Button variant="ghost" size="icon" class="md:hidden" @click="sidebarOpen = true">
          <Menu class="h-5 w-5" />
        </Button>
        <div class="flex-1" />
        <Button variant="ghost" size="icon" :title="theme === 'dark' ? 'Helles Design' : 'Dunkles Design'" @click="toggle">
          <Sun v-if="theme === 'dark'" class="h-4 w-4" />
          <Moon v-else class="h-4 w-4" />
        </Button>
        <Button
          v-if="auth.user?.admin"
          variant="ghost"
          size="icon"
          title="Energieoptionen"
          @click="powerDialog = 'reboot'"
        >
          <Power class="h-4 w-4" />
        </Button>
      </header>

      <main class="scrollbar-thin flex-1 overflow-y-auto">
        <div class="mx-auto max-w-5xl px-4 py-6 md:px-8 md:py-8">
          <RouterView />
        </div>
      </main>
    </div>

    <ConfirmDialog
      :open="powerDialog !== null"
      title="System neu starten?"
      message="Alle laufenden Dienste werden ordnungsgemäß beendet und das System startet neu."
      confirm-label="Neu starten"
      destructive
      @confirm="confirmPower"
      @cancel="powerDialog = null"
    />
  </div>
</template>
