<script setup lang="ts">
import { computed } from "vue";
import { Wifi, WifiOff, Loader2 } from "lucide-vue-next";
import { useI18n } from "vue-i18n";
import { ws } from "@/lib/ws";

// Spiegelt den Live-Verbindungsstatus (analog zur Cockpit-Statusanzeige).
const { t } = useI18n();
const status = ws.status;
// Ohne aktive Live-Abonnements ist „getrennt" der Normalzustand — dann wird
// die Anzeige ausgeblendet, statt einen Fehler zu suggerieren.
const visible = computed(() => ws.active.value > 0 || status.value !== "closed");
const label = computed(
  () =>
    ({
      open: t("connection.live"),
      connecting: t("connection.connecting"),
      closed: t("connection.disconnected"),
    })[status.value],
);
</script>

<template>
  <div
    v-if="visible"
    class="flex items-center gap-1.5 rounded-full px-2 py-1 text-xs font-medium"
    :class="
      status === 'open'
        ? 'text-success'
        : status === 'connecting'
          ? 'text-warning'
          : 'text-muted-foreground'
    "
    :title="`WebSocket: ${label}`"
  >
    <Wifi v-if="status === 'open'" class="h-3.5 w-3.5" />
    <Loader2 v-else-if="status === 'connecting'" class="h-3.5 w-3.5 animate-spin" />
    <WifiOff v-else class="h-3.5 w-3.5" />
    <span class="hidden sm:inline">{{ label }}</span>
  </div>
</template>
