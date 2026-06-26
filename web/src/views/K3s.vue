<script setup lang="ts">
import { ref, computed } from "vue";
import {
  Boxes,
  CheckCircle2,
  Download,
  Trash2,
  RotateCw,
  KeyRound,
  FileKey,
  Server,
  Copy,
} from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { useToast } from "@/composables/useToast";
import { useAuthStore } from "@/stores/auth";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import SettingsCard from "@/components/SettingsCard.vue";
import Expander from "@/components/Expander.vue";
import Button from "@/components/ui/Button.vue";
import Badge from "@/components/ui/Badge.vue";
import Switch from "@/components/ui/Switch.vue";
import JobConsole from "@/components/JobConsole.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import { Settings2 } from "lucide-vue-next";

const toast = useToast();
const auth = useAuthStore();

const { data: status, loading, error, reload } = useAsyncData(() => api.k3sStatus());
const nodes = ref<Awaited<ReturnType<typeof api.k3sNodes>>>([]);
const pods = ref<Awaited<ReturnType<typeof api.k3sPods>>>([]);

const jobId = ref<string | null>(null);
const disableTraefik = ref(false);
const showUninstall = ref(false);
const kubeconfig = ref("");
const token = ref("");

const isInstalled = computed(() => status.value?.installed === true);

async function refreshCluster(): Promise<void> {
  await reload();
  if (status.value?.active) {
    try {
      nodes.value = await api.k3sNodes();
      pods.value = await api.k3sPods();
    } catch {
      /* Cluster evtl. noch nicht bereit */
    }
  }
}
refreshCluster();

async function install(): Promise<void> {
  try {
    const { jobId: id } = await api.k3sInstall({
      disableTraefik: disableTraefik.value,
      writeKubeconfigMode: "644",
    });
    jobId.value = id;
    toast.info("k3s-Installation gestartet …");
  } catch (e) {
    toast.error((e as Error).message);
  }
}

async function onInstallFinished(s: string): Promise<void> {
  if (s === "succeeded") {
    toast.success("k3s installiert");
    await refreshCluster();
  } else if (s === "failed") {
    toast.error("k3s-Installation fehlgeschlagen — siehe Protokoll");
  }
}

async function uninstall(): Promise<void> {
  showUninstall.value = false;
  try {
    const { jobId: id } = await api.k3sUninstall();
    jobId.value = id;
  } catch (e) {
    toast.error((e as Error).message);
  }
}

async function loadKubeconfig(): Promise<void> {
  try {
    kubeconfig.value = (await api.k3sKubeconfig()).kubeconfig;
  } catch (e) {
    toast.error((e as Error).message);
  }
}
async function loadToken(): Promise<void> {
  try {
    token.value = (await api.k3sToken()).token;
  } catch (e) {
    toast.error((e as Error).message);
  }
}

async function copy(text: string): Promise<void> {
  await navigator.clipboard.writeText(text);
  toast.success("In die Zwischenablage kopiert");
}
</script>

<template>
  <div>
    <PageHeader
      title="k3s (Kubernetes)"
      description="Leichtgewichtigen Kubernetes-Cluster einrichten und verwalten"
      :breadcrumb="['Container', 'k3s']"
    >
      <template #actions>
        <Button variant="outline" size="sm" @click="refreshCluster">
          <RotateCw class="h-4 w-4" /> Aktualisieren
        </Button>
      </template>
    </PageHeader>

    <DataState :loading="loading" :error="error">
      <div class="space-y-6">
        <!-- Status -->
        <SettingsCard
          :icon="Boxes"
          title="Cluster-Status"
          :description="
            isInstalled
              ? status?.active
                ? `k3s läuft — ${status?.version || ''}`
                : 'k3s ist installiert, aber nicht aktiv.'
              : 'k3s ist nicht installiert.'
          "
        >
          <template #action>
            <Badge v-if="status?.active" variant="success">aktiv</Badge>
            <Badge v-else-if="isInstalled" variant="warning">inaktiv</Badge>
            <Badge v-else variant="neutral">nicht installiert</Badge>
          </template>
        </SettingsCard>

        <!-- Installation -->
        <div v-if="!isInstalled && auth.user?.admin" class="rounded-xl border border-border bg-card p-5 shadow-sm">
          <h2 class="text-sm font-semibold">k3s installieren</h2>
          <p class="mt-1 text-sm text-muted-foreground">
            Richtet einen Single-Node-Cluster über den offiziellen Installer ein
            (<code class="font-mono text-xs">get.k3s.io</code>).
          </p>
          <div class="mt-4">
            <Expander :icon="Settings2" title="Erweiterte Optionen" description="Standardkomponenten anpassen">
              <div class="flex items-center justify-between">
                <div>
                  <div class="text-sm font-medium">Traefik-Ingress deaktivieren</div>
                  <div class="text-xs text-muted-foreground">Empfohlen, wenn ein eigener Ingress genutzt wird.</div>
                </div>
                <Switch v-model="disableTraefik" />
              </div>
            </Expander>
          </div>
          <Button variant="primary" class="mt-4" :disabled="jobId !== null" @click="install">
            <Download class="h-4 w-4" /> Installation starten
          </Button>
        </div>

        <!-- Job-Konsole (Installation/Deinstallation) -->
        <div v-if="jobId" class="space-y-2">
          <h2 class="text-sm font-semibold">Installationsprotokoll</h2>
          <JobConsole :job-id="jobId" @finished="onInstallFinished" />
        </div>

        <!-- Cluster-Details -->
        <template v-if="isInstalled && status?.active">
          <section class="space-y-3">
            <h2 class="flex items-center gap-2 px-1 text-sm font-semibold">
              <Server class="h-4 w-4" /> Knoten ({{ nodes.length }})
            </h2>
            <div class="overflow-hidden rounded-xl border border-border bg-card shadow-sm">
              <div
                v-for="node in nodes"
                :key="node.name"
                class="flex items-center gap-3 border-b border-border/60 px-4 py-2.5 last:border-0"
              >
                <CheckCircle2 :class="['h-4 w-4', node.ready ? 'text-success' : 'text-muted-foreground']" />
                <span class="font-medium">{{ node.name }}</span>
                <Badge v-for="role in node.roles || []" :key="role" variant="info">{{ role }}</Badge>
                <span class="text-xs text-muted-foreground">{{ node.version }}</span>
                <span v-if="node.ip" class="ml-auto font-mono text-xs text-muted-foreground">{{ node.ip }}</span>
              </div>
            </div>
          </section>

          <section class="space-y-3">
            <h2 class="flex items-center gap-2 px-1 text-sm font-semibold">
              <Boxes class="h-4 w-4" /> Pods ({{ pods.length }})
            </h2>
            <div class="scrollbar-thin max-h-80 overflow-y-auto rounded-xl border border-border bg-card shadow-sm">
              <div
                v-for="pod in pods"
                :key="pod.namespace + '/' + pod.name"
                class="flex items-center gap-3 border-b border-border/60 px-4 py-2 text-sm last:border-0"
              >
                <Badge variant="neutral">{{ pod.namespace }}</Badge>
                <span class="min-w-0 flex-1 truncate font-mono text-xs">{{ pod.name }}</span>
                <span class="text-xs text-muted-foreground">{{ pod.ready }}</span>
                <Badge :variant="pod.phase === 'Running' ? 'success' : pod.phase === 'Failed' ? 'destructive' : 'warning'">
                  {{ pod.phase }}
                </Badge>
              </div>
            </div>
          </section>

          <!-- Zugang -->
          <section class="grid gap-4 md:grid-cols-2">
            <div class="rounded-xl border border-border bg-card p-5 shadow-sm">
              <div class="mb-2 flex items-center gap-2 text-sm font-semibold">
                <FileKey class="h-4 w-4" /> Kubeconfig
              </div>
              <Button v-if="!kubeconfig" variant="outline" size="sm" @click="loadKubeconfig">Anzeigen</Button>
              <div v-else class="space-y-2">
                <pre class="scrollbar-thin max-h-40 overflow-auto rounded-md bg-zinc-950 p-3 font-mono text-[11px] text-zinc-300">{{ kubeconfig }}</pre>
                <Button variant="ghost" size="sm" @click="copy(kubeconfig)"><Copy class="h-3.5 w-3.5" /> Kopieren</Button>
              </div>
            </div>
            <div class="rounded-xl border border-border bg-card p-5 shadow-sm">
              <div class="mb-2 flex items-center gap-2 text-sm font-semibold">
                <KeyRound class="h-4 w-4" /> Join-Token (Agents)
              </div>
              <Button v-if="!token" variant="outline" size="sm" @click="loadToken">Anzeigen</Button>
              <div v-else class="space-y-2">
                <pre class="scrollbar-thin overflow-auto rounded-md bg-zinc-950 p-3 font-mono text-[11px] text-zinc-300">{{ token }}</pre>
                <Button variant="ghost" size="sm" @click="copy(token)"><Copy class="h-3.5 w-3.5" /> Kopieren</Button>
              </div>
            </div>
          </section>
        </template>

        <!-- Deinstallation -->
        <div v-if="isInstalled && auth.user?.admin" class="rounded-xl border border-destructive/30 bg-destructive/5 p-5">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-sm font-semibold text-destructive">k3s deinstallieren</h2>
              <p class="text-xs text-muted-foreground">Entfernt k3s und alle Cluster-Daten unwiderruflich.</p>
            </div>
            <Button variant="destructive" size="sm" @click="showUninstall = true">
              <Trash2 class="h-4 w-4" /> Deinstallieren
            </Button>
          </div>
        </div>
      </div>
    </DataState>

    <ConfirmDialog
      :open="showUninstall"
      title="k3s deinstallieren?"
      message="k3s, alle Workloads und Cluster-Daten werden vollständig entfernt."
      confirm-label="Deinstallieren"
      destructive
      @confirm="uninstall"
      @cancel="showUninstall = false"
    />
  </div>
</template>
