<script setup lang="ts">
import { computed, ref } from "vue";
import { Search, RotateCw } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import Button from "@/components/ui/Button.vue";

const { data, loading, error, reload } = useAsyncData(() => api.processes(100));
const query = ref("");

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase();
  return (data.value ?? []).filter(
    (p) => !q || p.command.toLowerCase().includes(q) || String(p.pid).includes(q),
  );
});
</script>

<template>
  <div>
    <PageHeader
      title="Prozesse"
      description="Laufende Prozesse, sortiert nach Speicherverbrauch"
      :breadcrumb="['System', 'Prozesse']"
    >
      <template #actions>
        <Button variant="outline" size="sm" @click="reload">
          <RotateCw class="h-4 w-4" /> Aktualisieren
        </Button>
      </template>
    </PageHeader>

    <div class="relative mb-4 max-w-sm">
      <Search class="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
      <input
        v-model="query"
        type="search"
        placeholder="Prozess oder PID suchen …"
        class="h-9 w-full rounded-md border border-input bg-card pl-8 pr-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
      />
    </div>

    <DataState :loading="loading" :error="error" :empty="filtered.length === 0">
      <div class="overflow-hidden rounded-xl border border-border bg-card shadow-sm">
        <table class="w-full text-sm">
          <thead class="border-b border-border bg-secondary/40 text-left text-xs text-muted-foreground">
            <tr>
              <th class="px-4 py-2 font-medium">PID</th>
              <th class="px-4 py-2 font-medium">Benutzer</th>
              <th class="px-4 py-2 font-medium">Kommando</th>
              <th class="px-4 py-2 text-right font-medium">Speicher</th>
              <th class="px-4 py-2 text-right font-medium">Threads</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in filtered" :key="p.pid" class="border-b border-border/50 last:border-0 hover:bg-accent/40">
              <td class="px-4 py-2 font-mono text-xs">{{ p.pid }}</td>
              <td class="px-4 py-2">{{ p.user || "—" }}</td>
              <td class="max-w-md truncate px-4 py-2 font-mono text-xs">{{ p.command }}</td>
              <td class="px-4 py-2 text-right tabular-nums">{{ p.rssMB.toFixed(1) }} MB</td>
              <td class="px-4 py-2 text-right tabular-nums text-muted-foreground">{{ p.threads }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </DataState>
  </div>
</template>
