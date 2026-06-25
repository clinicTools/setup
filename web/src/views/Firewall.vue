<script setup lang="ts">
import { ShieldCheck, ShieldOff, RotateCw } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { useToast } from "@/composables/useToast";
import { useAuthStore } from "@/stores/auth";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import SettingsCard from "@/components/SettingsCard.vue";
import Switch from "@/components/ui/Switch.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";

const { data, loading, error, reload } = useAsyncData(() => api.firewall());
const toast = useToast();
const auth = useAuthStore();

async function toggle(value: boolean): Promise<void> {
  try {
    await api.setFirewall(value);
    toast.success(value ? "Firewall aktiviert" : "Firewall deaktiviert");
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  }
}

function ruleVariant(action: string): "success" | "destructive" | "warning" | "neutral" {
  if (action.includes("ALLOW")) return "success";
  if (action.includes("DENY") || action.includes("REJECT")) return "destructive";
  if (action.includes("LIMIT")) return "warning";
  return "neutral";
}
</script>

<template>
  <div>
    <PageHeader
      title="Firewall"
      description="Paketfilter-Status und Regeln (UFW)"
      :breadcrumb="['Sicherheit', 'Firewall']"
    >
      <template #actions>
        <Button variant="outline" size="sm" @click="reload">
          <RotateCw class="h-4 w-4" /> Aktualisieren
        </Button>
      </template>
    </PageHeader>

    <DataState :loading="loading" :error="error">
      <div v-if="data" class="space-y-6">
        <div
          v-if="!data.available"
          class="rounded-xl border border-warning/30 bg-warning/5 px-4 py-6 text-center text-sm"
        >
          UFW ist auf diesem System nicht installiert.
        </div>

        <template v-else>
          <SettingsCard
            :icon="data.enabled ? ShieldCheck : ShieldOff"
            title="Firewall-Status"
            :description="data.enabled ? 'Die Firewall ist aktiv und filtert eingehenden Verkehr.' : 'Die Firewall ist derzeit deaktiviert.'"
          >
            <template #action>
              <Badge :variant="data.enabled ? 'success' : 'neutral'">
                {{ data.enabled ? "Aktiv" : "Inaktiv" }}
              </Badge>
              <Switch
                v-if="auth.user?.admin"
                :model-value="data.enabled"
                @update:model-value="toggle"
              />
            </template>
          </SettingsCard>

          <section v-if="data.rules?.length" class="space-y-3">
            <h2 class="px-1 text-sm font-semibold">Regeln</h2>
            <div class="overflow-hidden rounded-xl border border-border bg-card shadow-sm">
              <div
                v-for="(rule, i) in data.rules"
                :key="i"
                class="flex items-center gap-4 border-b border-border/60 px-4 py-2.5 last:border-0"
              >
                <span class="min-w-0 flex-1 truncate font-mono text-sm">{{ rule.to }}</span>
                <Badge :variant="ruleVariant(rule.action)">{{ rule.action }}</Badge>
                <span class="w-40 truncate text-right font-mono text-xs text-muted-foreground">
                  {{ rule.from }}
                </span>
              </div>
            </div>
          </section>
        </template>
      </div>
    </DataState>
  </div>
</template>
