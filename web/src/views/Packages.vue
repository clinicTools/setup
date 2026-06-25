<script setup lang="ts">
import { ref, computed } from "vue";
import { Package, RefreshCw, ArrowUpCircle, Download, Trash2, Search } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { useToast } from "@/composables/useToast";
import { useAuthStore } from "@/stores/auth";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import StatCard from "@/components/StatCard.vue";
import Button from "@/components/ui/Button.vue";
import Input from "@/components/ui/Input.vue";
import Badge from "@/components/ui/Badge.vue";
import JobConsole from "@/components/JobConsole.vue";

const { data, loading, error, reload } = useAsyncData(() => api.packages());
const { data: installed, reload: reloadInstalled } = useAsyncData(() => api.installedPackages());
const toast = useToast();
const auth = useAuthStore();

const jobId = ref<string | null>(null);
const updating = ref(false);
const installInput = ref("");
const search = ref("");

const filteredInstalled = computed(() => {
  const q = search.value.trim().toLowerCase();
  const list = installed.value ?? [];
  const shown = q ? list.filter((p) => p.name.toLowerCase().includes(q)) : list;
  return shown.slice(0, 300);
});

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

async function upgradeAll(): Promise<void> {
  try {
    jobId.value = (await api.packageUpgrade()).jobId;
    toast.info("System-Upgrade gestartet …");
  } catch (e) {
    toast.error((e as Error).message);
  }
}

async function install(): Promise<void> {
  const pkgs = installInput.value.split(/[\s,]+/).filter(Boolean);
  if (pkgs.length === 0) return;
  try {
    jobId.value = (await api.packageInstall(pkgs)).jobId;
    installInput.value = "";
  } catch (e) {
    toast.error((e as Error).message);
  }
}

async function remove(name: string): Promise<void> {
  try {
    jobId.value = (await api.packageRemove([name])).jobId;
  } catch (e) {
    toast.error((e as Error).message);
  }
}

function onJobFinished(): void {
  reload();
  reloadInstalled();
}
</script>

<template>
  <div>
    <PageHeader
      title="Pakete & Updates"
      description="Installierte Pakete und Aktualisierungen verwalten (APT)"
      :breadcrumb="['Apps', 'Pakete & Updates']"
    >
      <template #actions>
        <Button variant="outline" size="sm" :disabled="updating" @click="runUpdate">
          <RefreshCw :class="['h-4 w-4', updating && 'animate-spin']" /> Paketlisten
        </Button>
        <Button
          v-if="auth.user?.admin && (data?.upgradable ?? 0) > 0"
          variant="primary"
          size="sm"
          @click="upgradeAll"
        >
          <ArrowUpCircle class="h-4 w-4" /> Alle aktualisieren
        </Button>
      </template>
    </PageHeader>

    <DataState :loading="loading" :error="error">
      <div v-if="data" class="space-y-6">
        <div class="grid grid-cols-2 gap-3 sm:grid-cols-3">
          <StatCard label="Installiert" :value="String(data.installed)" sub="Pakete" :icon="Package" />
          <StatCard label="Aktualisierbar" :value="String(data.upgradable)" sub="Updates verfügbar" :icon="ArrowUpCircle" accent />
        </div>

        <!-- Aktive Job-Ausgabe -->
        <div v-if="jobId" class="space-y-2">
          <h2 class="text-sm font-semibold">Vorgang</h2>
          <JobConsole :job-id="jobId" @finished="onJobFinished" />
        </div>

        <!-- Paket installieren -->
        <div v-if="auth.user?.admin" class="rounded-xl border border-border bg-card p-4 shadow-sm">
          <label class="text-sm font-medium">Pakete installieren</label>
          <div class="mt-2 flex gap-2">
            <Input v-model="installInput" placeholder="z. B. htop curl git" @keyup.enter="install" />
            <Button variant="primary" :disabled="!installInput.trim()" @click="install">
              <Download class="h-4 w-4" /> Installieren
            </Button>
          </div>
        </div>

        <!-- Verfügbare Upgrades -->
        <section v-if="data.upgrades?.length" class="space-y-3">
          <h2 class="px-1 text-sm font-semibold">Verfügbare Aktualisierungen ({{ data.upgrades.length }})</h2>
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

        <!-- Installierte Pakete -->
        <section class="space-y-3">
          <div class="flex items-center justify-between">
            <h2 class="px-1 text-sm font-semibold">Installierte Pakete</h2>
            <div class="relative w-64">
              <Search class="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <input
                v-model="search"
                type="search"
                placeholder="Paket suchen …"
                class="h-9 w-full rounded-md border border-input bg-card pl-8 pr-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              />
            </div>
          </div>
          <div class="scrollbar-thin max-h-96 overflow-y-auto rounded-xl border border-border bg-card shadow-sm">
            <div
              v-for="p in filteredInstalled"
              :key="p.name"
              class="flex items-center gap-3 border-b border-border/60 px-4 py-2 text-sm last:border-0"
            >
              <span class="min-w-0 flex-1 truncate font-medium">{{ p.name }}</span>
              <span class="font-mono text-xs text-muted-foreground">{{ p.version }}</span>
              <Badge v-if="p.architecture" variant="neutral">{{ p.architecture }}</Badge>
              <Button v-if="auth.user?.admin" variant="ghost" size="icon" title="Entfernen" @click="remove(p.name)">
                <Trash2 class="h-4 w-4 text-destructive" />
              </Button>
            </div>
          </div>
        </section>
      </div>
    </DataState>
  </div>
</template>
