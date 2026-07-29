<script setup lang="ts">
import { ref } from "vue";
import { Network as NetIcon, Globe, Router, RotateCw, Check, Power } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { useToast } from "@/composables/useToast";
import { useAuthStore } from "@/stores/auth";
import type { NetworkInterface } from "@/lib/types";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import Input from "@/components/ui/Input.vue";

const { data, loading, error, reload } = useAsyncData(() => api.network());
const toast = useToast();
const auth = useAuthStore();
const hostname = ref("");

async function saveHostname(): Promise<void> {
  if (!hostname.value) return;
  try {
    await api.setHostname(hostname.value);
    toast.success("Hostname gesetzt");
    hostname.value = "";
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  }
}

async function toggleInterface(iface: NetworkInterface): Promise<void> {
  try {
    await api.setInterfaceState(iface.name, !iface.up);
    toast.success(`${iface.name} ${iface.up ? "deaktiviert" : "aktiviert"}`);
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  }
}
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
            <div v-if="auth.user?.admin" class="mt-2 flex gap-2">
              <Input v-model="hostname" placeholder="Neuer Hostname" class="h-8 text-xs" />
              <Button variant="outline" size="icon" class="h-8 w-8" title="Setzen" :disabled="!hostname" @click="saveHostname">
                <Check class="h-4 w-4" />
              </Button>
            </div>
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
              <Button
                v-if="auth.user?.admin && !iface.loopback"
                variant="ghost"
                size="sm"
                class="ml-auto"
                :title="iface.up ? 'Deaktivieren' : 'Aktivieren'"
                @click="toggleInterface(iface)"
              >
                <Power class="h-3.5 w-3.5" /> {{ iface.up ? "Down" : "Up" }}
              </Button>
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
