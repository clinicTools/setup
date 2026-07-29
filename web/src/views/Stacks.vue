<script setup lang="ts">
import { computed, ref } from "vue";
import {
  Layers,
  Plus,
  Play,
  Square,
  RotateCw,
  Download,
  Trash2,
  Save,
  ArrowLeft,
  FileCode2,
} from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { useToast } from "@/composables/useToast";
import { useAuthStore } from "@/stores/auth";
import type { Stack } from "@/lib/types";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import Button from "@/components/ui/Button.vue";
import Badge from "@/components/ui/Badge.vue";
import Input from "@/components/ui/Input.vue";
import CodeEditor from "@/components/CodeEditor.vue";
import JobConsole from "@/components/JobConsole.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";

const toast = useToast();
const auth = useAuthStore();
const { data: stacks, loading, error, reload } = useAsyncData(() => api.stacks());

/** Vorlage für einen neuen Stack. */
const TEMPLATE = `# Compose-Datei für diesen Stack
services:
  web:
    image: docker.io/library/nginx:alpine
    ports:
      - "8080:80"
    restart: unless-stopped
    volumes:
      - web-data:/usr/share/nginx/html

volumes:
  web-data:
`;

// Ansichtsmodus: Liste ⇄ Editor (Detailseite wie in den Windows-11-Einstellungen)
const editing = ref<null | { name: string; compose: string; isNew: boolean }>(null);
const jobId = ref<string | null>(null);
const busy = ref(false);
const deleteTarget = ref<Stack | null>(null);

const dirty = ref(false);
const originalCompose = ref("");
const canSave = computed(
  () => editing.value !== null && editing.value.name.trim().length > 0 && editing.value.compose.trim().length > 0,
);

/** Zielpfad der compose.yaml (Anzeige im Editor). */
const composePath = computed(
  () => `/etc/debian-admin/stacks/${editing.value?.name.trim() || "(name)"}/compose.yaml`,
);

function statusVariant(s: Stack["status"]): "success" | "warning" | "neutral" {
  if (s === "running") return "success";
  if (s === "partial") return "warning";
  return "neutral";
}
const statusLabel: Record<Stack["status"], string> = {
  running: "läuft",
  partial: "teilweise",
  stopped: "gestoppt",
  unknown: "unbekannt",
};

function newStack(): void {
  editing.value = { name: "", compose: TEMPLATE, isNew: true };
  originalCompose.value = TEMPLATE;
  dirty.value = false;
  jobId.value = null;
}

async function openStack(s: Stack): Promise<void> {
  try {
    const { compose } = await api.stack(s.name);
    editing.value = { name: s.name, compose, isNew: false };
    originalCompose.value = compose;
    dirty.value = false;
    jobId.value = null;
  } catch (e) {
    toast.error((e as Error).message);
  }
}

function closeEditor(): void {
  editing.value = null;
  jobId.value = null;
  reload();
}

async function save(): Promise<void> {
  if (!editing.value) return;
  busy.value = true;
  try {
    await api.saveStack(editing.value.name.trim(), editing.value.compose);
    originalCompose.value = editing.value.compose;
    editing.value.isNew = false;
    dirty.value = false;
    toast.success("compose.yaml gespeichert");
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  } finally {
    busy.value = false;
  }
}

/** Speichert bei Bedarf und startet anschließend die Compose-Aktion als Job. */
async function runAction(action: string, name?: string): Promise<void> {
  const stackName = name ?? editing.value?.name.trim();
  if (!stackName) return;
  try {
    if (editing.value && dirty.value) await save();
    jobId.value = (await api.stackAction(stackName, action)).jobId;
  } catch (e) {
    toast.error((e as Error).message);
  }
}

async function onJobFinished(status: string): Promise<void> {
  if (status === "succeeded") toast.success("Vorgang abgeschlossen");
  else if (status === "failed") toast.error("Vorgang fehlgeschlagen — siehe Protokoll");
  await reload();
}

async function confirmDelete(): Promise<void> {
  const s = deleteTarget.value;
  deleteTarget.value = null;
  if (!s) return;
  try {
    await api.deleteStack(s.name);
    toast.success(`Stack „${s.name}" entfernt`);
    if (editing.value?.name === s.name) editing.value = null;
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  }
}

function onComposeInput(value: string): void {
  if (!editing.value) return;
  editing.value.compose = value;
  dirty.value = value !== originalCompose.value;
}
</script>

<template>
  <div>
    <!-- ---------- Detailansicht: Editor ---------- -->
    <template v-if="editing">
      <PageHeader
        :title="editing.isNew ? 'Neuer Stack' : editing.name"
        description="compose.yaml bearbeiten und Stack steuern"
        :breadcrumb="['Apps & Container', 'Compose-Stacks', editing.isNew ? 'Neu' : editing.name]"
      >
        <template #actions>
          <Button variant="ghost" size="sm" @click="closeEditor">
            <ArrowLeft class="h-4 w-4" /> Zurück
          </Button>
          <Button variant="primary" size="sm" :disabled="!canSave || busy" @click="save">
            <Save class="h-4 w-4" /> {{ busy ? "Speichert …" : "Speichern" }}
          </Button>
        </template>
      </PageHeader>

      <div class="space-y-4">
        <!-- Name (nur bei neuem Stack änderbar) -->
        <div v-if="editing.isNew" class="max-w-sm space-y-1.5">
          <label class="text-sm font-medium">Stack-Name</label>
          <Input v-model="editing.name" placeholder="z. B. nextcloud" />
          <p class="text-xs text-muted-foreground">Erlaubt: Kleinbuchstaben, Ziffern, Bindestrich und Unterstrich.</p>
        </div>

        <!-- Compose-Editor -->
        <div class="space-y-2">
          <div class="flex items-center justify-between px-1">
            <h2 class="flex items-center gap-2 text-sm font-semibold">
              <FileCode2 class="h-4 w-4" /> compose.yaml
              <Badge v-if="dirty" variant="warning">ungespeichert</Badge>
            </h2>
            <div v-if="auth.user?.admin && !editing.isNew" class="flex items-center gap-1">
              <Button variant="outline" size="sm" @click="runAction('up')">
                <Play class="h-4 w-4 text-success" /> Starten
              </Button>
              <Button variant="outline" size="sm" @click="runAction('down')">
                <Square class="h-4 w-4 text-destructive" /> Stoppen
              </Button>
              <Button variant="outline" size="sm" @click="runAction('restart')">
                <RotateCw class="h-4 w-4" /> Neu starten
              </Button>
              <Button variant="outline" size="sm" @click="runAction('pull')">
                <Download class="h-4 w-4" /> Images holen
              </Button>
            </div>
          </div>
          <CodeEditor
            :model-value="editing.compose"
            min-height="24rem"
            @update:model-value="onComposeInput"
          />
          <p class="px-1 text-xs text-muted-foreground">
            Gespeichert unter
            <code class="app-selectable font-mono">{{ composePath }}</code>
            (nur für root lesbar).
          </p>
        </div>

        <!-- Ausgabe der Compose-Aktion -->
        <div v-if="jobId" class="space-y-2">
          <h2 class="px-1 text-sm font-semibold">Ausgabe</h2>
          <JobConsole :job-id="jobId" @finished="onJobFinished" />
        </div>
      </div>
    </template>

    <!-- ---------- Listenansicht ---------- -->
    <template v-else>
      <PageHeader
        title="Compose-Stacks"
        description="Container-Verbunde per compose.yaml verwalten"
        :breadcrumb="['Apps & Container', 'Compose-Stacks']"
      >
        <template #actions>
          <Button variant="outline" size="sm" @click="reload">
            <RotateCw class="h-4 w-4" /> Aktualisieren
          </Button>
          <Button v-if="auth.user?.admin" variant="primary" size="sm" @click="newStack">
            <Plus class="h-4 w-4" /> Neuer Stack
          </Button>
        </template>
      </PageHeader>

      <DataState
        :loading="loading"
        :error="error"
        :empty="(stacks?.length ?? 0) === 0"
        empty-text="Noch keine Stacks angelegt — oben rechts einen neuen Stack anlegen."
      >
        <div class="space-y-1.5">
          <div
            v-for="s in stacks || []"
            :key="s.name"
            class="flex w-full items-center gap-3.5 rounded-xl border border-border/70 bg-card px-4 py-3 shadow-sm transition-all hover:border-border hover:bg-accent/40"
          >
            <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
              <Layers class="h-5 w-5" />
            </div>
            <button class="min-w-0 flex-1 text-left" @click="openStack(s)">
              <div class="flex items-center gap-2">
                <span class="truncate text-sm font-medium">{{ s.name }}</span>
                <Badge :variant="statusVariant(s.status)">{{ statusLabel[s.status] }}</Badge>
              </div>
              <p class="truncate text-xs text-muted-foreground">
                {{ s.running }} von {{ s.containers }} Containern aktiv
              </p>
            </button>
            <div v-if="auth.user?.admin" class="flex shrink-0 items-center gap-1">
              <Button
                v-if="s.status !== 'running'"
                variant="ghost"
                size="icon"
                title="Starten"
                @click="runAction('up', s.name)"
              >
                <Play class="h-4 w-4 text-success" />
              </Button>
              <Button v-else variant="ghost" size="icon" title="Stoppen" @click="runAction('down', s.name)">
                <Square class="h-4 w-4 text-destructive" />
              </Button>
              <Button variant="ghost" size="icon" title="Bearbeiten" @click="openStack(s)">
                <FileCode2 class="h-4 w-4" />
              </Button>
              <Button variant="ghost" size="icon" title="Löschen" @click="deleteTarget = s">
                <Trash2 class="h-4 w-4 text-destructive" />
              </Button>
            </div>
          </div>
        </div>
      </DataState>

      <!-- Ausgabe einer Aktion aus der Liste -->
      <div v-if="jobId" class="mt-5 space-y-2">
        <h2 class="px-1 text-sm font-semibold">Ausgabe</h2>
        <JobConsole :job-id="jobId" @finished="onJobFinished" />
      </div>
    </template>

    <ConfirmDialog
      :open="deleteTarget !== null"
      title="Stack löschen?"
      :message="`Die compose.yaml von „${deleteTarget?.name}“ wird gelöscht. Laufende Container werden dadurch nicht automatisch entfernt — zuvor „Stoppen“ ausführen.`"
      confirm-label="Löschen"
      destructive
      @confirm="confirmDelete"
      @cancel="deleteTarget = null"
    />
  </div>
</template>
