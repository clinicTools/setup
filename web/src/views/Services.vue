<script setup lang="ts">
import { computed, ref } from "vue";
import { Play, Square, RotateCw, Search } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { useChannel } from "@/composables/useChannel";
import { useToast } from "@/composables/useToast";
import { useAuthStore } from "@/stores/auth";
import type { Service } from "@/lib/types";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import Button from "@/components/ui/Button.vue";
import Badge from "@/components/ui/Badge.vue";

const { data, loading, error, reload } = useAsyncData(() => api.services());
// Live-Aktualisierung des Dienst-Status (alle 3 s, solange die Seite offen ist).
const { data: liveServices } = useChannel<Service[]>("services");
const toast = useToast();
const auth = useAuthStore();

const query = ref("");
const onlyActive = ref(false);
const busy = ref<string | null>(null);

const filtered = computed<Service[]>(() => {
  const list = liveServices.value ?? data.value ?? [];
  const q = query.value.trim().toLowerCase();
  return list.filter((s) => {
    if (onlyActive.value && s.activeState !== "active") return false;
    if (q && !s.name.toLowerCase().includes(q) && !s.description.toLowerCase().includes(q))
      return false;
    return true;
  });
});

function stateVariant(s: Service): "success" | "destructive" | "neutral" {
  if (s.activeState === "active") return "success";
  if (s.activeState === "failed") return "destructive";
  return "neutral";
}

async function act(s: Service, action: string): Promise<void> {
  busy.value = s.name;
  try {
    await api.serviceAction(s.name, action);
    toast.success(`${s.name}: ${action} ausgeführt`);
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  } finally {
    busy.value = null;
  }
}
</script>

<template>
  <div>
    <PageHeader
      title="Dienste"
      description="systemd-Units verwalten — starten, stoppen, aktivieren"
      :breadcrumb="['System', 'Dienste']"
    >
      <template #actions>
        <Button variant="outline" size="sm" @click="reload">
          <RotateCw class="h-4 w-4" /> Aktualisieren
        </Button>
      </template>
    </PageHeader>

    <div class="mb-4 flex flex-wrap items-center gap-3">
      <div class="relative min-w-56 flex-1">
        <Search class="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          v-model="query"
          type="search"
          placeholder="Dienst suchen …"
          class="h-9 w-full rounded-md border border-input bg-card pl-8 pr-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        />
      </div>
      <label class="flex items-center gap-2 text-sm text-muted-foreground">
        <input v-model="onlyActive" type="checkbox" class="h-4 w-4 rounded border-input" />
        Nur aktive
      </label>
    </div>

    <DataState :loading="loading" :error="error" :empty="filtered.length === 0">
      <div class="overflow-hidden rounded-xl border border-border bg-card shadow-sm">
        <div
          v-for="s in filtered"
          :key="s.name"
          class="flex items-center gap-4 border-b border-border/60 px-4 py-3 last:border-0"
        >
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="truncate text-sm font-medium">{{ s.name }}</span>
              <Badge :variant="stateVariant(s)">{{ s.subState || s.activeState }}</Badge>
              <Badge v-if="s.unitFileState === 'enabled'" variant="info">aktiviert</Badge>
            </div>
            <p class="truncate text-xs text-muted-foreground">{{ s.description }}</p>
          </div>
          <div v-if="auth.user?.admin" class="flex shrink-0 items-center gap-1">
            <Button
              v-if="s.activeState !== 'active'"
              variant="ghost"
              size="icon"
              title="Starten"
              :disabled="busy === s.name"
              @click="act(s, 'start')"
            >
              <Play class="h-4 w-4 text-success" />
            </Button>
            <Button
              v-else
              variant="ghost"
              size="icon"
              title="Stoppen"
              :disabled="busy === s.name"
              @click="act(s, 'stop')"
            >
              <Square class="h-4 w-4 text-destructive" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              title="Neu starten"
              :disabled="busy === s.name"
              @click="act(s, 'restart')"
            >
              <RotateCw class="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>
    </DataState>
  </div>
</template>
