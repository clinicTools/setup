<script setup lang="ts">
import { computed, ref, watch, onUnmounted } from "vue";
import {
  Boxes,
  Play,
  Square,
  RotateCw,
  Trash2,
  Download,
  ScrollText,
  Search,
  X,
} from "lucide-vue-next";
import { api } from "@/lib/api";
import { ws } from "@/lib/ws";
import { useAsyncData } from "@/composables/useAsyncData";
import { useChannel } from "@/composables/useChannel";
import { useToast } from "@/composables/useToast";
import { useAuthStore } from "@/stores/auth";
import { formatBytes } from "@/lib/utils";
import type { Container } from "@/lib/types";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import SettingsCard from "@/components/SettingsCard.vue";
import Button from "@/components/ui/Button.vue";
import Badge from "@/components/ui/Badge.vue";
import JobConsole from "@/components/JobConsole.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";

const toast = useToast();
const auth = useAuthStore();

const { data: status, loading, error, reload: reloadStatus } = useAsyncData(() => api.podmanStatus());
const { data: initial, reload: reloadContainers } = useAsyncData(() => api.containers());
const { data: images, reload: reloadImages } = useAsyncData(() => api.containerImages());

// Live-Aktualisierung (alle 3 s), solange die Seite offen ist.
const { data: live } = useChannel<Container[]>("containers");
const containers = computed<Container[]>(() => live.value ?? initial.value ?? []);

const query = ref("");
const busy = ref<string | null>(null);
const installJobId = ref<string | null>(null);
const removeTarget = ref<Container | null>(null);

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase();
  if (!q) return containers.value;
  return containers.value.filter(
    (c) =>
      c.name.toLowerCase().includes(q) ||
      c.image.toLowerCase().includes(q) ||
      (c.stack ?? "").toLowerCase().includes(q),
  );
});

const runningCount = computed(() => containers.value.filter((c) => c.state === "running").length);

function stateVariant(state: string): "success" | "destructive" | "warning" | "neutral" {
  if (state === "running") return "success";
  if (state === "exited" || state === "dead") return "neutral";
  if (state === "paused") return "warning";
  return "neutral";
}

async function act(c: Container, action: string): Promise<void> {
  busy.value = c.id;
  try {
    await api.containerAction(c.id, action);
    toast.success(`${c.name}: ${action}`);
    await reloadContainers();
  } catch (e) {
    toast.error((e as Error).message);
  } finally {
    busy.value = null;
  }
}

async function confirmRemove(): Promise<void> {
  const c = removeTarget.value;
  removeTarget.value = null;
  if (c) await act(c, "rm");
}

async function installPodman(): Promise<void> {
  try {
    installJobId.value = (await api.podmanInstall()).jobId;
  } catch (e) {
    toast.error((e as Error).message);
  }
}

async function onInstallFinished(s: string): Promise<void> {
  if (s === "succeeded") {
    toast.success("Podman installiert");
    await reloadStatus();
    await reloadContainers();
    await reloadImages();
  } else if (s === "failed") {
    toast.error("Installation fehlgeschlagen — siehe Protokoll");
  }
}

// --- Live-Logs eines Containers ---

const logTarget = ref<Container | null>(null);
const logLines = ref<string[]>([]);
let unsubLogs: (() => void) | null = null;

function stopLogs(): void {
  unsubLogs?.();
  unsubLogs = null;
}

watch(logTarget, (c) => {
  stopLogs();
  logLines.value = [];
  if (!c) return;
  unsubLogs = ws.subscribe("containerlogs", { id: c.id, tail: 200 }, (payload) => {
    const line = (payload as { line: string }).line;
    logLines.value.push(line);
    if (logLines.value.length > 2000) logLines.value.splice(0, logLines.value.length - 2000);
  });
});
onUnmounted(stopLogs);
</script>

<template>
  <div>
    <PageHeader
      title="Container"
      description="Podman-Container verwalten"
      :breadcrumb="['Apps & Container', 'Container']"
    >
      <template #actions>
        <Button variant="outline" size="sm" @click="reloadContainers">
          <RotateCw class="h-4 w-4" /> Aktualisieren
        </Button>
      </template>
    </PageHeader>

    <DataState :loading="loading" :error="error">
      <div class="space-y-5">
        <!-- Status -->
        <SettingsCard
          :icon="Boxes"
          title="Podman"
          :description="
            status?.installed
              ? `Version ${status.version ?? '?'} · ${runningCount} von ${containers.length} Containern laufen` +
                (status.composeAvailable ? ` · Compose: ${status.composeCommand}` : ' · kein Compose-Werkzeug')
              : 'Podman ist auf diesem System nicht installiert.'
          "
        >
          <template #action>
            <Badge v-if="status?.installed" variant="success">installiert</Badge>
            <Badge v-else variant="neutral">nicht installiert</Badge>
            <Button
              v-if="!status?.installed && auth.user?.admin"
              variant="primary"
              size="sm"
              :disabled="installJobId !== null"
              @click="installPodman"
            >
              <Download class="h-4 w-4" /> Installieren
            </Button>
          </template>
        </SettingsCard>

        <div v-if="installJobId" class="space-y-2">
          <h2 class="px-1 text-sm font-semibold">Installationsprotokoll</h2>
          <JobConsole :job-id="installJobId" @finished="onInstallFinished" />
        </div>

        <template v-if="status?.installed">
          <!-- Suche -->
          <div class="relative max-w-sm">
            <Search
              class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
            />
            <input
              v-model="query"
              type="search"
              placeholder="Container suchen …"
              class="h-8 w-full rounded border border-input bg-card pl-9 pr-3 text-[13px] focus:border-primary focus:outline-none"
            />
          </div>

          <!-- Container-Liste -->
          <section class="space-y-3">
            <h2 class="px-1 text-sm font-semibold">Container ({{ filtered.length }})</h2>
            <div
              v-if="filtered.length"
              class="overflow-hidden rounded-xl border border-border/70 bg-card shadow-sm"
            >
              <div
                v-for="c in filtered"
                :key="c.id"
                class="flex items-center gap-3 border-b border-border/50 px-4 py-3 last:border-0 hover:bg-accent/30"
              >
                <span
                  class="h-2 w-2 shrink-0 rounded-full"
                  :class="c.state === 'running' ? 'bg-success' : 'bg-muted-foreground/40'"
                />
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <span class="truncate text-sm font-medium">{{ c.name }}</span>
                    <Badge :variant="stateVariant(c.state)">{{ c.state }}</Badge>
                    <Badge v-if="c.stack" variant="info">{{ c.stack }}</Badge>
                  </div>
                  <p class="truncate text-xs text-muted-foreground">
                    {{ c.image }}
                    <template v-if="c.ports?.length"> · {{ c.ports.join(", ") }}</template>
                    <template v-if="c.status"> · {{ c.status }}</template>
                  </p>
                </div>
                <div class="flex shrink-0 items-center gap-1">
                  <Button variant="ghost" size="icon" title="Protokoll" @click="logTarget = c">
                    <ScrollText class="h-4 w-4" />
                  </Button>
                  <template v-if="auth.user?.admin">
                    <Button
                      v-if="c.state !== 'running'"
                      variant="ghost"
                      size="icon"
                      title="Starten"
                      :disabled="busy === c.id"
                      @click="act(c, 'start')"
                    >
                      <Play class="h-4 w-4 text-success" />
                    </Button>
                    <Button
                      v-else
                      variant="ghost"
                      size="icon"
                      title="Stoppen"
                      :disabled="busy === c.id"
                      @click="act(c, 'stop')"
                    >
                      <Square class="h-4 w-4 text-destructive" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      title="Neu starten"
                      :disabled="busy === c.id"
                      @click="act(c, 'restart')"
                    >
                      <RotateCw class="h-4 w-4" />
                    </Button>
                    <Button variant="ghost" size="icon" title="Entfernen" @click="removeTarget = c">
                      <Trash2 class="h-4 w-4 text-destructive" />
                    </Button>
                  </template>
                </div>
              </div>
            </div>
            <p v-else class="px-1 text-sm text-muted-foreground">Keine Container vorhanden.</p>
          </section>

          <!-- Images -->
          <section v-if="images?.length" class="space-y-3">
            <h2 class="px-1 text-sm font-semibold">Images ({{ images.length }})</h2>
            <div class="scrollbar-thin max-h-72 overflow-y-auto rounded-xl border border-border/70 bg-card shadow-sm">
              <div
                v-for="img in images"
                :key="img.id"
                class="flex items-center gap-3 border-b border-border/50 px-4 py-2 text-sm last:border-0"
              >
                <span class="min-w-0 flex-1 truncate">{{ (img.names || []).join(", ") || img.id }}</span>
                <span class="shrink-0 text-xs text-muted-foreground">{{ formatBytes(img.size) }}</span>
              </div>
            </div>
          </section>
        </template>
      </div>
    </DataState>

    <!-- Live-Protokoll eines Containers -->
    <Teleport to="body">
      <div
        v-if="logTarget"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
        @click.self="logTarget = null"
      >
        <div class="flex h-[70vh] w-full max-w-3xl flex-col overflow-hidden rounded-xl border border-border bg-popover shadow-2xl">
          <div class="flex items-center justify-between border-b border-border px-4 py-2.5">
            <div class="flex items-center gap-2 text-sm font-medium">
              <ScrollText class="h-4 w-4" />
              {{ logTarget.name }}
              <span class="relative flex h-2 w-2">
                <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-success opacity-75" />
                <span class="relative inline-flex h-2 w-2 rounded-full bg-success" />
              </span>
            </div>
            <Button variant="ghost" size="icon" @click="logTarget = null">
              <X class="h-4 w-4" />
            </Button>
          </div>
          <div class="scrollbar-thin app-selectable flex-1 overflow-auto bg-zinc-950 p-3 font-mono text-xs text-zinc-300">
            <div v-for="(l, i) in logLines" :key="i" class="whitespace-pre-wrap break-words">{{ l }}</div>
            <div v-if="!logLines.length" class="text-zinc-500">Warte auf Ausgabe …</div>
          </div>
        </div>
      </div>
    </Teleport>

    <ConfirmDialog
      :open="removeTarget !== null"
      title="Container entfernen?"
      :message="`„${removeTarget?.name}“ wird gestoppt und gelöscht.`"
      confirm-label="Entfernen"
      destructive
      @confirm="confirmRemove"
      @cancel="removeTarget = null"
    />
  </div>
</template>
