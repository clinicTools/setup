<script setup lang="ts">
import { Network as NetIcon, Globe, Router, RotateCw } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";

const { data, loading, error, reload } = useAsyncData(() => api.network());
</script>

<template>
  <div>
    <PageHeader
      title="Netzwerk"
      description="Schnittstellen, Adressen und Namensauflösung"
      :breadcrumb="['Verbindungen', 'Netzwerk']"
    >
      <template #actions>
        <Button variant="outline" size="sm" @click="reload">
          <RotateCw class="h-4 w-4" /> Aktualisieren
        </Button>
      </template>
    </PageHeader>

    <DataState :loading="loading" :error="error">
      <div v-if="data" class="space-y-6">
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
            <div class="flex items-center gap-2 text-xs text-muted-foreground">
              <Globe class="h-4 w-4" /> Hostname
            </div>
            <div class="mt-1 font-medium">{{ data.hostname || "—" }}</div>
          </div>
          <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
            <div class="flex items-center gap-2 text-xs text-muted-foreground">
              <Router class="h-4 w-4" /> Gateway
            </div>
            <div class="mt-1 font-mono text-sm">{{ data.gateway || "—" }}</div>
          </div>
          <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
            <div class="flex items-center gap-2 text-xs text-muted-foreground">
              <NetIcon class="h-4 w-4" /> DNS-Server
            </div>
            <div class="mt-1 font-mono text-sm">{{ (data.dns || []).join(", ") || "—" }}</div>
          </div>
        </div>

        <div class="overflow-hidden rounded-xl border border-border bg-card shadow-sm">
          <div
            v-for="iface in data.interfaces"
            :key="iface.name"
            class="border-b border-border/60 px-4 py-3 last:border-0"
          >
            <div class="flex items-center gap-2">
              <span class="font-medium">{{ iface.name }}</span>
              <Badge :variant="iface.up ? 'success' : 'neutral'">{{ iface.up ? "up" : "down" }}</Badge>
              <Badge v-if="iface.loopback" variant="neutral">loopback</Badge>
              <span class="text-xs text-muted-foreground">MTU {{ iface.mtu }}</span>
            </div>
            <div class="mt-1 space-y-0.5">
              <p v-if="iface.mac" class="font-mono text-xs text-muted-foreground">MAC {{ iface.mac }}</p>
              <p
                v-for="addr in iface.addresses || []"
                :key="addr"
                class="font-mono text-xs text-foreground/80"
              >
                {{ addr }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </DataState>
  </div>
</template>
