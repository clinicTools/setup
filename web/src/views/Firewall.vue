<script setup lang="ts">
import { ref } from "vue";
import { ShieldCheck, ShieldOff, RotateCw, Plus, Trash2 } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { useToast } from "@/composables/useToast";
import { useAuthStore } from "@/stores/auth";
import type { FirewallRule } from "@/lib/types";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import SettingsCard from "@/components/SettingsCard.vue";
import Switch from "@/components/ui/Switch.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import Input from "@/components/ui/Input.vue";

const { data, loading, error, reload } = useAsyncData(() => api.firewall());
const toast = useToast();
const auth = useAuthStore();

const newRule = ref({ action: "allow", port: "", protocol: "tcp" });

async function toggle(value: boolean): Promise<void> {
  try {
    await api.setFirewall(value);
    toast.success(value ? "Firewall aktiviert" : "Firewall deaktiviert");
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  }
}

async function addRule(): Promise<void> {
  if (!newRule.value.port) return;
  try {
    await api.firewallAddRule(newRule.value.action, newRule.value.port, newRule.value.protocol);
    toast.success("Regel hinzugefügt");
    newRule.value.port = "";
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  }
}

async function deleteRule(rule: FirewallRule): Promise<void> {
  const [port, proto] = rule.to.split("/");
  const action = rule.action.toLowerCase().split(" ")[0];
  try {
    await api.firewallDeleteRule(action, port, proto || "");
    toast.success("Regel entfernt");
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

          <!-- Regel hinzufügen -->
          <div v-if="auth.user?.admin" class="rounded-xl border border-border bg-card p-4 shadow-sm">
            <label class="text-sm font-medium">Regel hinzufügen</label>
            <div class="mt-2 flex flex-wrap items-center gap-2">
              <select
                v-model="newRule.action"
                class="h-9 rounded-md border border-input bg-card px-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              >
                <option value="allow">allow</option>
                <option value="deny">deny</option>
                <option value="reject">reject</option>
                <option value="limit">limit</option>
              </select>
              <Input v-model="newRule.port" placeholder="Port (z. B. 22)" class="w-40" />
              <select
                v-model="newRule.protocol"
                class="h-9 rounded-md border border-input bg-card px-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              >
                <option value="tcp">tcp</option>
                <option value="udp">udp</option>
                <option value="">beide</option>
              </select>
              <Button variant="primary" :disabled="!newRule.port" @click="addRule">
                <Plus class="h-4 w-4" /> Hinzufügen
              </Button>
            </div>
          </div>

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
                <span class="w-32 truncate text-right font-mono text-xs text-muted-foreground">
                  {{ rule.from }}
                </span>
                <Button v-if="auth.user?.admin" variant="ghost" size="icon" title="Regel löschen" @click="deleteRule(rule)">
                  <Trash2 class="h-4 w-4 text-destructive" />
                </Button>
              </div>
            </div>
          </section>
        </template>
      </div>
    </DataState>
  </div>
</template>
