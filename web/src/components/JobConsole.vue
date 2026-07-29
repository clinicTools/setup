<script setup lang="ts">
import { ref, watch, nextTick, onUnmounted } from "vue";
import { Loader2, CheckCircle2, XCircle, Ban } from "lucide-vue-next";
import { ws } from "@/lib/ws";
import { api } from "@/lib/api";
import type { JobLine } from "@/lib/types";
import Button from "./ui/Button.vue";

// Verfolgt einen Job über den WebSocket-Channel „jobs": Verlauf + Live-Stream
// + Endstatus. Wird für apt-Aktionen, Podman/Compose und Installationen verwendet.
const props = defineProps<{ jobId: string | null }>();
const emit = defineEmits<{ (e: "finished", status: string): void }>();

const lines = ref<JobLine[]>([]);
const status = ref<"running" | "succeeded" | "failed" | "canceled">("running");
const container = ref<HTMLElement | null>(null);
let unsubscribe: (() => void) | null = null;

function start(id: string): void {
  stop();
  lines.value = [];
  status.value = "running";
  unsubscribe = ws.subscribe("jobs", { jobId: id }, (payload) => {
    const p = payload as JobLine & { finished?: boolean; status?: string };
    if (p.finished) {
      status.value = (p.status as typeof status.value) ?? "succeeded";
      emit("finished", status.value);
    } else {
      lines.value.push(p as JobLine);
      void nextTick(() => {
        if (container.value) container.value.scrollTop = container.value.scrollHeight;
      });
    }
  });
}

function stop(): void {
  unsubscribe?.();
  unsubscribe = null;
}

watch(
  () => props.jobId,
  (id) => (id ? start(id) : stop()),
  { immediate: true },
);
onUnmounted(stop);

async function cancel(): Promise<void> {
  if (props.jobId) await api.cancelJob(props.jobId);
}

/** Entfernt ANSI-Steuerzeichen (docker-compose faerbt seine Ausgabe ein). */
const ANSI = /\u001b\[[0-9;?]*[ -/]*[@-~]/g;
function clean(text: string): string {
  return text.replace(ANSI, "");
}

const streamColor = (s: string) => (s === "stderr" ? "text-red-400" : s === "system" ? "text-primary" : "text-zinc-300");
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-border">
    <div class="flex items-center justify-between border-b border-border bg-secondary/40 px-3 py-1.5">
      <div class="flex items-center gap-2 text-xs font-medium">
        <Loader2 v-if="status === 'running'" class="h-3.5 w-3.5 animate-spin text-warning" />
        <CheckCircle2 v-else-if="status === 'succeeded'" class="h-3.5 w-3.5 text-success" />
        <Ban v-else-if="status === 'canceled'" class="h-3.5 w-3.5 text-muted-foreground" />
        <XCircle v-else class="h-3.5 w-3.5 text-destructive" />
        <span>{{
          { running: "läuft …", succeeded: "erfolgreich", failed: "fehlgeschlagen", canceled: "abgebrochen" }[status]
        }}</span>
      </div>
      <Button v-if="status === 'running'" variant="ghost" size="sm" @click="cancel">Abbrechen</Button>
    </div>
    <div
      ref="container"
      class="scrollbar-thin h-64 overflow-y-auto bg-zinc-950 p-3 font-mono text-xs leading-relaxed"
    >
      <div v-for="line in lines" :key="line.seq" :class="streamColor(line.stream)">
        {{ clean(line.text) }}
      </div>
      <div v-if="lines.length === 0" class="text-zinc-500">Warte auf Ausgabe …</div>
    </div>
  </div>
</template>
