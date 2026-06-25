<script setup lang="ts">
import { ref, watch } from "vue";
import { Clock, Globe2, Wifi } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { useToast } from "@/composables/useToast";
import { useAuthStore } from "@/stores/auth";
import { formatDateTime } from "@/lib/utils";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import SettingsCard from "@/components/SettingsCard.vue";
import SettingsGroup from "@/components/SettingsGroup.vue";
import Switch from "@/components/ui/Switch.vue";
import Badge from "@/components/ui/Badge.vue";

const { data, loading, error, reload } = useAsyncData(() => api.time());
const { data: zones } = useAsyncData(() => api.timezones());
const toast = useToast();
const auth = useAuthStore();

const selectedZone = ref("");
watch(data, (d) => {
  if (d) selectedZone.value = d.timezone;
});

async function changeZone(): Promise<void> {
  try {
    await api.setTimezone(selectedZone.value);
    toast.success(`Zeitzone auf ${selectedZone.value} gesetzt`);
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  }
}

async function toggleNTP(value: boolean): Promise<void> {
  try {
    await api.setNTP(value);
    toast.success(value ? "Zeitsynchronisation aktiviert" : "Zeitsynchronisation deaktiviert");
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  }
}
</script>

<template>
  <div>
    <PageHeader
      title="Datum & Uhrzeit"
      description="Zeitzone und automatische Zeitsynchronisation"
      :breadcrumb="['Zeit & Sprache', 'Datum & Uhrzeit']"
    />

    <DataState :loading="loading" :error="error">
      <SettingsGroup v-if="data" title="Uhrzeit" class="mb-6">
        <SettingsCard :icon="Clock" title="Lokale Zeit" :description="formatDateTime(data.localTime)">
          <template #action>
            <Badge variant="info">{{ data.timezone }}</Badge>
          </template>
        </SettingsCard>
        <SettingsCard
          :icon="Globe2"
          title="UTC"
          :description="formatDateTime(data.universalTime)"
        />
      </SettingsGroup>

      <SettingsGroup v-if="data" title="Einstellungen">
        <SettingsCard
          :icon="Wifi"
          title="Zeit automatisch synchronisieren"
          :description="data.ntpSynced ? 'Synchronisiert über NTP.' : 'Aktiviert NTP-Zeitsynchronisation.'"
        >
          <template #action>
            <Badge v-if="data.ntpSynced" variant="success">synchron</Badge>
            <Switch
              :model-value="data.ntpEnabled"
              :disabled="!auth.user?.admin"
              @update:model-value="toggleNTP"
            />
          </template>
        </SettingsCard>

        <SettingsCard :icon="Globe2" title="Zeitzone" description="Systemweite Zeitzone festlegen">
          <template #action>
            <select
              v-model="selectedZone"
              :disabled="!auth.user?.admin"
              class="h-9 max-w-56 rounded-md border border-input bg-card px-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50"
              @change="changeZone"
            >
              <option v-for="z in zones || []" :key="z" :value="z">{{ z }}</option>
            </select>
          </template>
        </SettingsCard>
      </SettingsGroup>
    </DataState>
  </div>
</template>
