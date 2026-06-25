<script setup lang="ts">
import { computed, ref } from "vue";
import { Search, RotateCw, X, Skull } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { useChannel } from "@/composables/useChannel";
import { useToast } from "@/composables/useToast";
import { useAuthStore } from "@/stores/auth";
import type { Process } from "@/lib/types";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import Button from "@/components/ui/Button.vue";

const { data, loading, error, reload } = useAsyncData(() => api.processes(100));
// Live-Aktualisierung (alle 2 s) solange die Seite offen ist.
const { data: live } = useChannel<Process[]>("processes");
const toast = useToast();
const auth = useAuthStore();
const query = ref("");

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase();
  const source = live.value ?? data.value ?? [];
  return source.filter(
    (p) => !q || p.command.toLowerCase().includes(q) || String(p.pid).includes(q),
  );
});

async function kill(p: Process, signal: string): Promise<void> {
  try {
    await api.killProcess(p.pid, signal);
    toast.success(`Signal ${signal} an PID ${p.pid} gesendet`);
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  }
}
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
              <th v-if="auth.user?.admin" class="px-4 py-2 text-right font-medium">Aktion</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in filtered" :key="p.pid" class="border-b border-border/50 last:border-0 hover:bg-accent/40">
              <td class="px-4 py-2 font-mono text-xs">{{ p.pid }}</td>
              <td class="px-4 py-2">{{ p.user || "—" }}</td>
              <td class="max-w-md truncate px-4 py-2 font-mono text-xs">{{ p.command }}</td>
              <td class="px-4 py-2 text-right tabular-nums">{{ p.rssMB.toFixed(1) }} MB</td>
              <td class="px-4 py-2 text-right tabular-nums text-muted-foreground">{{ p.threads }}</td>
              <td v-if="auth.user?.admin" class="px-4 py-2 text-right">
                <div class="flex justify-end gap-1">
                  <Button variant="ghost" size="icon" title="Beenden (SIGTERM)" @click="kill(p, 'TERM')">
                    <X class="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="icon" title="Erzwingen (SIGKILL)" @click="kill(p, 'KILL')">
                    <Skull class="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </DataState>
  </div>
</template>
