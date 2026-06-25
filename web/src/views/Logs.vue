<script setup lang="ts">
import { ref } from "vue";
import { RotateCw, Search } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { formatDateTime } from "@/lib/utils";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import Button from "@/components/ui/Button.vue";

const unit = ref("");
const priority = ref("");
const lines = ref(200);

const { data, loading, error, reload } = useAsyncData(() =>
  api.logs({ unit: unit.value || undefined, priority: priority.value || undefined, lines: lines.value }),
);

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
        <Button variant="outline" size="sm" @click="reload">
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

    <DataState :loading="loading" :error="error" :empty="(data?.length ?? 0) === 0">
      <div class="scrollbar-thin overflow-x-auto rounded-xl border border-border bg-card font-mono text-xs shadow-sm">
        <div
          v-for="(entry, i) in data || []"
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
