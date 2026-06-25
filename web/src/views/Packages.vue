<script setup lang="ts">
import { ref } from "vue";
import { Package, RefreshCw, ArrowUpCircle } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { useToast } from "@/composables/useToast";
import { useAuthStore } from "@/stores/auth";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import StatCard from "@/components/StatCard.vue";
import Button from "@/components/ui/Button.vue";

const { data, loading, error, reload } = useAsyncData(() => api.packages());
const toast = useToast();
const auth = useAuthStore();
const updating = ref(false);

async function runUpdate(): Promise<void> {
  updating.value = true;
  try {
    await api.aptUpdate();
    toast.success("Paketlisten aktualisiert");
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  } finally {
    updating.value = false;
  }
}
</script>

<template>
  <div>
    <PageHeader
      title="Pakete & Updates"
      description="Installierte Pakete und verfügbare Aktualisierungen (APT)"
      :breadcrumb="['Apps', 'Pakete & Updates']"
    >
      <template #actions>
        <Button
          v-if="auth.user?.admin"
          variant="primary"
          size="sm"
          :disabled="updating"
          @click="runUpdate"
        >
          <RefreshCw :class="['h-4 w-4', updating && 'animate-spin']" />
          {{ updating ? "Aktualisiere …" : "Paketlisten aktualisieren" }}
        </Button>
      </template>
    </PageHeader>

    <DataState :loading="loading" :error="error">
      <div v-if="data" class="space-y-6">
        <div class="grid grid-cols-2 gap-3 sm:grid-cols-3">
          <StatCard label="Installiert" :value="String(data.installed)" sub="Pakete" :icon="Package" />
          <StatCard
            label="Aktualisierbar"
            :value="String(data.upgradable)"
            sub="Updates verfügbar"
            :icon="ArrowUpCircle"
            accent
          />
        </div>

        <section v-if="data.upgrades?.length" class="space-y-3">
          <h2 class="px-1 text-sm font-semibold">Verfügbare Aktualisierungen</h2>
          <div class="overflow-hidden rounded-xl border border-border bg-card shadow-sm">
            <div
              v-for="p in data.upgrades"
              :key="p.name"
              class="flex items-center gap-4 border-b border-border/60 px-4 py-2.5 last:border-0"
            >
              <Package class="h-4 w-4 shrink-0 text-muted-foreground" />
              <span class="min-w-0 flex-1 truncate text-sm font-medium">{{ p.name }}</span>
              <span class="font-mono text-xs text-muted-foreground">{{ p.version || "?" }}</span>
              <ArrowUpCircle class="h-3.5 w-3.5 text-success" />
              <span class="font-mono text-xs text-success">{{ p.availableVersion }}</span>
            </div>
          </div>
        </section>

        <div
          v-else
          class="rounded-xl border border-success/30 bg-success/5 px-4 py-6 text-center text-sm text-success"
        >
          Das System ist auf dem aktuellen Stand.
        </div>
      </div>
    </DataState>
  </div>
</template>
