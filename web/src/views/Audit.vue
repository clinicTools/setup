<script setup lang="ts">
import { RotateCw, ShieldCheck } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { formatDateTime } from "@/lib/utils";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import Button from "@/components/ui/Button.vue";
import Badge from "@/components/ui/Badge.vue";

const { data, loading, error, reload } = useAsyncData(() => api.audit(300));
</script>

<template>
  <div>
    <PageHeader
      title="Audit-Protokoll"
      description="Nachvollziehbare Aufzeichnung administrativer Aktionen"
      :breadcrumb="['Sicherheit', 'Audit-Protokoll']"
    >
      <template #actions>
        <Button variant="outline" size="sm" @click="reload">
          <RotateCw class="h-4 w-4" /> Aktualisieren
        </Button>
      </template>
    </PageHeader>

    <DataState :loading="loading" :error="error" :empty="(data?.length ?? 0) === 0" empty-text="Noch keine Aktionen protokolliert.">
      <div class="overflow-hidden rounded-xl border border-border bg-card shadow-sm">
        <table class="w-full text-sm">
          <thead class="border-b border-border bg-secondary/40 text-left text-xs text-muted-foreground">
            <tr>
              <th class="px-4 py-2 font-medium">Zeit</th>
              <th class="px-4 py-2 font-medium">Benutzer</th>
              <th class="px-4 py-2 font-medium">Aktion</th>
              <th class="px-4 py-2 font-medium">Ergebnis</th>
              <th class="px-4 py-2 font-medium">Quelle</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(e, i) in data || []" :key="i" class="border-b border-border/50 last:border-0 hover:bg-accent/40">
              <td class="whitespace-nowrap px-4 py-2 text-xs text-muted-foreground">{{ formatDateTime(e.time) }}</td>
              <td class="px-4 py-2 font-medium">{{ e.user }}</td>
              <td class="px-4 py-2 font-mono text-xs">{{ e.action }}</td>
              <td class="px-4 py-2">
                <Badge :variant="e.success ? 'success' : 'destructive'">
                  <ShieldCheck class="h-3 w-3" /> {{ e.detail || (e.success ? 'OK' : 'Fehler') }}
                </Badge>
              </td>
              <td class="px-4 py-2 font-mono text-xs text-muted-foreground">{{ e.ip }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </DataState>
  </div>
</template>
