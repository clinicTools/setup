<script setup lang="ts">
import { HardDrive, RotateCw } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { formatBytes } from "@/lib/utils";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import UsageBar from "@/components/UsageBar.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";

const { data, loading, error, reload } = useAsyncData(() => api.storage());
</script>

<template>
  <div>
    <PageHeader
      title="Speicher"
      description="Dateisysteme und Blockgeräte"
      :breadcrumb="['Geräte', 'Speicher']"
    >
      <template #actions>
        <Button variant="outline" size="sm" @click="reload">
          <RotateCw class="h-4 w-4" /> Aktualisieren
        </Button>
      </template>
    </PageHeader>

    <DataState :loading="loading" :error="error">
      <div v-if="data" class="space-y-6">
        <section class="space-y-3">
          <h2 class="px-1 text-sm font-semibold">Eingehängte Dateisysteme</h2>
          <div class="grid gap-3 md:grid-cols-2">
            <div
              v-for="fs in data.filesystems || []"
              :key="fs.mountpoint"
              class="rounded-xl border border-border bg-card p-4 shadow-sm"
            >
              <div class="mb-2 flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <HardDrive class="h-4 w-4 text-primary" />
                  <span class="font-medium">{{ fs.mountpoint }}</span>
                </div>
                <Badge variant="neutral">{{ fs.type }}</Badge>
              </div>
              <UsageBar :percent="fs.usePercent" />
              <p class="mt-2 text-xs text-muted-foreground">
                {{ formatBytes(fs.used) }} belegt · {{ formatBytes(fs.free) }} frei ·
                {{ formatBytes(fs.total) }} gesamt
              </p>
              <p class="truncate font-mono text-[11px] text-muted-foreground">{{ fs.device }}</p>
            </div>
          </div>
        </section>

        <section v-if="data.devices?.length" class="space-y-3">
          <h2 class="px-1 text-sm font-semibold">Blockgeräte</h2>
          <div class="overflow-hidden rounded-xl border border-border bg-card shadow-sm">
            <template v-for="dev in data.devices" :key="dev.name">
              <div class="flex items-center gap-3 border-b border-border/60 px-4 py-2.5 last:border-0">
                <span class="font-mono text-sm font-medium">{{ dev.name }}</span>
                <Badge variant="info">{{ dev.type }}</Badge>
                <span class="text-sm text-muted-foreground">{{ dev.size }}</span>
                <span v-if="dev.model" class="truncate text-xs text-muted-foreground">{{ dev.model }}</span>
              </div>
              <div
                v-for="child in dev.children || []"
                :key="child.name"
                class="flex items-center gap-3 border-b border-border/60 px-4 py-2 pl-10 last:border-0"
              >
                <span class="font-mono text-sm">{{ child.name }}</span>
                <Badge variant="neutral">{{ child.fstype || child.type }}</Badge>
                <span class="text-sm text-muted-foreground">{{ child.size }}</span>
                <span v-if="child.mountpoint" class="text-xs text-muted-foreground">→ {{ child.mountpoint }}</span>
              </div>
            </template>
          </div>
        </section>
      </div>
    </DataState>
  </div>
</template>
