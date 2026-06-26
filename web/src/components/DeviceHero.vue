<script setup lang="ts">
import { MonitorCog } from "lucide-vue-next";
import type { HostInfo } from "@/lib/types";
import { formatUptime } from "@/lib/utils";

/**
 * Geräte-Hero — angelehnt an die System-Startseite der Windows-11-Einstellungen:
 * eine hervorgehobene Karte mit Gerätesymbol, Gerätename und Eckdaten.
 */
defineProps<{ info: HostInfo }>();
</script>

<template>
  <div
    class="relative overflow-hidden rounded-2xl border border-border/60 bg-card p-6 shadow-sm"
  >
    <!-- dezenter Farbverlauf (Mica-Akzent) -->
    <div
      class="pointer-events-none absolute inset-0 opacity-90"
      style="
        background-image:
          radial-gradient(600px 240px at 12% -20%, color-mix(in oklab, var(--bloom-1) 16%, transparent), transparent 70%),
          radial-gradient(520px 240px at 100% 120%, color-mix(in oklab, var(--bloom-2) 14%, transparent), transparent 70%);
      "
    />
    <div class="relative flex flex-col gap-5 sm:flex-row sm:items-center">
      <div
        class="flex h-20 w-20 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-primary/15 to-primary/5 text-primary ring-1 ring-primary/15"
      >
        <MonitorCog class="h-10 w-10" />
      </div>
      <div class="min-w-0 flex-1">
        <h2 class="truncate text-2xl font-semibold tracking-tight">{{ info.hostname || "—" }}</h2>
        <p class="text-sm text-muted-foreground">{{ info.prettyName }}</p>
        <div class="mt-3 flex flex-wrap gap-x-6 gap-y-1.5 text-xs">
          <span class="text-muted-foreground">
            Kernel <span class="font-mono text-foreground/80">{{ info.kernel }}</span>
          </span>
          <span class="text-muted-foreground">
            Architektur <span class="font-medium text-foreground/80">{{ info.architecture }}</span>
          </span>
          <span class="text-muted-foreground">
            Laufzeit <span class="font-medium text-foreground/80">{{ formatUptime(info.uptimeSeconds) }}</span>
          </span>
          <span v-if="info.virtualization" class="text-muted-foreground">
            Virtualisierung <span class="font-medium text-foreground/80">{{ info.virtualization }}</span>
          </span>
          <span v-if="info.cpuModel" class="text-muted-foreground">
            Prozessor <span class="font-medium text-foreground/80">{{ info.cpuModel }}</span>
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
