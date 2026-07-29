<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from "vue";
import { RotateCw, Search, Play, Pause } from "lucide-vue-next";
import { api } from "@/lib/api";
import { ws } from "@/lib/ws";
import { useAsyncData } from "@/composables/useAsyncData";
import { formatDateTime } from "@/lib/utils";
import type { LogEntry } from "@/lib/types";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import Button from "@/components/ui/Button.vue";

const unit = ref("");
const priority = ref("");
const lines = ref(200);

const { data, loading, error, reload } = useAsyncData(() =>
  api.logs({ unit: unit.value || undefined, priority: priority.value || undefined, lines: lines.value }),
);

// Live-Follow (journalctl -f) über den WebSocket-Channel „journal".
const live = ref(false);
const liveEntries = ref<LogEntry[]>([]);
let unsub: (() => void) | null = null;

function startLive(): void {
  stopLive();
  liveEntries.value = [];
  unsub = ws.subscribe(
    "journal",
    { unit: unit.value || undefined, priority: priority.value || undefined, lines: 50 },
    (payload) => {
      liveEntries.value.push(payload as LogEntry);
      if (liveEntries.value.length > 1000) {
        liveEntries.value.splice(0, liveEntries.value.length - 1000);
      }
    },
  );
}
function stopLive(): void {
  unsub?.();
  unsub = null;
}

watch(live, (v) => (v ? startLive() : stopLive()));
watch([unit, priority], () => {
  if (live.value) startLive();
});
onUnmounted(stopLive);

const displayed = computed<LogEntry[]>(() => (live.value ? liveEntries.value : data.value || []));

// Farbliche Hervorhebung nach syslog-Priorität (0=emerg … 7=debug).
function priorityClass(p: number): string {
  if (p <= 3) return "text-destructive";
  if (p === 4) return "text-warning";
  if (p === 5) return "text-foreground";
  return "text-muted-foreground";
}
const priorityLabels: Record<number, string> = {
  0: "emerg", 1: "alert", 2: "crit", 3: "err",
  4: "warning", 5: "notice", 6: "info", 7: "debug",
};
</script>

<template>
  <div>
    <PageHeader
      title="Systemprotokolle"
      description="Journal-Einträge (journald)"
      :breadcrumb="['Diagnose', 'Systemprotokolle']"
    >
      <template #actions>
        <Button
          :variant="live ? 'primary' : 'outline'"
          size="sm"
          @click="live = !live"
        >
          <Pause v-if="live" class="h-4 w-4" />
          <Play v-else class="h-4 w-4" />
          {{ live ? "Live aktiv" : "Live-Verfolgung" }}
        </Button>
        <Button variant="outline" size="sm" :disabled="live" @click="reload">
          <RotateCw class="h-4 w-4" /> Aktualisieren
        </Button>
      </template>
    </PageHeader>

    <div class="mb-4 flex flex-wrap items-center gap-3">
      <div class="relative min-w-48 flex-1">
        <Search class="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          v-model="unit"
          type="search"
          placeholder="Unit filtern (z. B. ssh.service)"
          class="h-9 w-full rounded-md border border-input bg-card pl-8 pr-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          @keyup.enter="reload"
        />
      </div>
      <select
        v-model="priority"
        class="h-9 rounded-md border border-input bg-card px-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        @change="reload"
      >
        <option value="">Alle Prioritäten</option>
        <option value="err">Fehler und höher</option>
        <option value="warning">Warnungen und höher</option>
        <option value="info">Info und höher</option>
      </select>
      <Button variant="secondary" size="sm" @click="reload">Anwenden</Button>
    </div>

    <DataState
      :loading="loading && !live"
      :error="error"
      :empty="!live && displayed.length === 0"
      empty-text="Keine Protokolleinträge. Aktiviere die Live-Verfolgung, um neue Einträge zu sehen."
    >
      <div class="scrollbar-thin max-h-[70vh] overflow-y-auto overflow-x-auto rounded-xl border border-border bg-card font-mono text-xs shadow-sm">
        <div
          v-if="live"
          class="sticky top-0 flex items-center gap-2 border-b border-border bg-card/95 px-4 py-1.5 text-[11px] font-medium text-success backdrop-blur"
        >
          <span class="relative flex h-2 w-2">
            <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-success opacity-75" />
            <span class="relative inline-flex h-2 w-2 rounded-full bg-success" />
          </span>
          Live — folgt dem Journal ({{ displayed.length }} Einträge)
        </div>
        <div
          v-for="(entry, i) in displayed"
          :key="i"
          class="flex items-start gap-3 border-b border-border/40 px-4 py-1.5 last:border-0 hover:bg-accent/40"
        >
          <span class="shrink-0 text-muted-foreground">{{ formatDateTime(entry.timestamp) }}</span>
          <span :class="['w-16 shrink-0 font-semibold', priorityClass(entry.priority)]">
            {{ priorityLabels[entry.priority] || entry.priority }}
          </span>
          <span class="shrink-0 text-primary/80">{{ entry.unit || "—" }}</span>
          <span class="min-w-0 flex-1 whitespace-pre-wrap break-words text-foreground/90">{{ entry.message }}</span>
        </div>
      </div>
    </DataState>
  </div>
</template>
