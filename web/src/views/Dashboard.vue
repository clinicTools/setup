<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { Package, Cog, HardDrive, Gauge } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { useChannel } from "@/composables/useChannel";
import { formatBytes } from "@/lib/utils";
import type { MetricsSample } from "@/lib/types";
import PageHeader from "@/components/PageHeader.vue";
import StatCard from "@/components/StatCard.vue";
import DataState from "@/components/DataState.vue";
import Sparkline from "@/components/Sparkline.vue";
import DeviceHero from "@/components/DeviceHero.vue";

const { t } = useI18n();
const { data: info, loading, error } = useAsyncData(() => api.info());
const { data: pkg } = useAsyncData(() => api.packages());
const { data: storage } = useAsyncData(() => api.storage());
const { data: services } = useAsyncData(() => api.services());

// Live-Kennzahlen über den WebSocket-Channel „metrics" (läuft nur, solange
// das Dashboard offen ist).
const { data: metrics, history } = useChannel<MetricsSample>("metrics", { buffer: 60 });
const cpuSeries = computed(() => history.value.map((m) => m.cpuPercent));
const memSeries = computed(() =>
  history.value.map((m) => (m.memory.total ? (m.memory.used / m.memory.total) * 100 : 0)),
);
const netTotalRate = computed(() =>
  (metrics.value?.interfaces ?? []).reduce((sum, i) => sum + i.rxRate + i.txRate, 0),
);
const liveMemPercent = computed(() => {
  const m = metrics.value?.memory;
  return m && m.total ? (m.used / m.total) * 100 : 0;
});

const rootFs = computed(() => storage.value?.filesystems?.find((f) => f.mountpoint === "/"));
const activeServices = computed(() => (services.value ?? []).filter((s) => s.activeState === "active").length);
const totalServices = computed(() => (services.value ?? []).length);
</script>

<template>
  <div>
    <PageHeader
      :title="t('dashboard.title')"
      :description="t('dashboard.description')"
      :breadcrumb="[t('nav.groups.system'), t('dashboard.title')]"
    />

    <DataState :loading="loading" :error="error">
      <div v-if="info" class="space-y-5">
        <!-- Geräte-Hero (Windows-11-System-Startseite) -->
        <DeviceHero :info="info" />

        <!-- Live-Auslastung (WebSocket) -->
        <div class="grid gap-4 md:grid-cols-3">
          <div class="rounded-xl border border-border/70 bg-card p-5 shadow-sm">
            <div class="flex items-baseline justify-between">
              <h2 class="text-sm font-semibold">CPU</h2>
              <span class="text-2xl font-semibold tabular-nums">
                {{ metrics ? `${metrics.cpuPercent.toFixed(0)} %` : "—" }}
              </span>
            </div>
            <Sparkline class="mt-3 w-full" :values="cpuSeries" :max="100" :width="240" :height="44" />
            <p class="mt-1 text-xs text-muted-foreground">{{ info.cpuCores }} Kerne · Last {{ info.loadAvg[0].toFixed(2) }}</p>
          </div>

          <div class="rounded-xl border border-border/70 bg-card p-5 shadow-sm">
            <div class="flex items-baseline justify-between">
              <h2 class="text-sm font-semibold">Arbeitsspeicher</h2>
              <span class="text-2xl font-semibold tabular-nums">
                {{ metrics ? `${liveMemPercent.toFixed(0)} %` : "—" }}
              </span>
            </div>
            <Sparkline class="mt-3 w-full" :values="memSeries" :max="100" :width="240" :height="44" />
            <p class="mt-1 text-xs text-muted-foreground">
              {{ formatBytes(info.memory.used) }} / {{ formatBytes(info.memory.total) }}
            </p>
          </div>

          <div class="rounded-xl border border-border/70 bg-card p-5 shadow-sm">
            <div class="flex items-baseline justify-between">
              <h2 class="text-sm font-semibold">Netzwerk</h2>
              <span class="text-lg font-semibold tabular-nums">
                {{ metrics ? `${formatBytes(netTotalRate)}/s` : "—" }}
              </span>
            </div>
            <div class="mt-3 space-y-1">
              <div
                v-for="iface in metrics?.interfaces || []"
                :key="iface.name"
                class="flex justify-between text-xs text-muted-foreground"
              >
                <span class="font-mono">{{ iface.name }}</span>
                <span>↓ {{ formatBytes(iface.rxRate) }}/s · ↑ {{ formatBytes(iface.txRate) }}/s</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Schnellzugriff (navigierbare Hero-Controls) -->
        <div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
          <StatCard
            label="Aktualisierungen"
            :value="String(pkg?.upgradable ?? 0)"
            :sub="`${pkg?.installed ?? 0} Pakete installiert`"
            :icon="Package"
            accent
            to="/packages"
          />
          <StatCard
            label="Dienste aktiv"
            :value="`${activeServices} / ${totalServices}`"
            sub="systemd-Units"
            :icon="Cog"
            to="/services"
          />
          <StatCard
            v-if="rootFs"
            label="Systemdatenträger"
            :value="`${rootFs.usePercent.toFixed(0)} %`"
            :sub="`${formatBytes(rootFs.free)} frei · ${rootFs.type}`"
            :icon="HardDrive"
            to="/storage"
          />
          <StatCard
            label="Lastdurchschnitt"
            :value="info.loadAvg[0].toFixed(2)"
            :sub="info.loadAvg.map((l) => l.toFixed(2)).join(' · ')"
            :icon="Gauge"
          />
        </div>
      </div>
    </DataState>
  </div>
</template>
