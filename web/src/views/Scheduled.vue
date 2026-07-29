<script setup lang="ts">
import { ref } from "vue";
import { CalendarClock, Timer as TimerIcon, RotateCw, Plus, Trash2 } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { useToast } from "@/composables/useToast";
import { useAuthStore } from "@/stores/auth";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import Button from "@/components/ui/Button.vue";
import Input from "@/components/ui/Input.vue";

const { data, loading, error, reload } = useAsyncData(() => api.scheduled());
const toast = useToast();
const auth = useAuthStore();

const showForm = ref(false);
const form = ref({ name: "", schedule: "0 * * * *", user: "root", command: "" });

async function createCron(): Promise<void> {
  try {
    await api.createCron(form.value.name, form.value.schedule, form.value.user, form.value.command);
    toast.success("Cron-Job angelegt");
    showForm.value = false;
    form.value = { name: "", schedule: "0 * * * *", user: "root", command: "" };
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  }
}

// Nur von dieser Anwendung verwaltete Jobs sind löschbar.
function managedName(source: string): string | null {
  const m = source.match(/debian-admin-(.+)$/);
  return m ? m[1] : null;
}
async function deleteCron(name: string): Promise<void> {
  try {
    await api.deleteCron(name);
    toast.success("Cron-Job gelöscht");
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  }
}
</script>

<template>
  <div>
    <PageHeader
      title="Geplante Aufgaben"
      description="systemd-Timer und Cron-Jobs"
      :breadcrumb="['Aufgaben', 'Geplante Aufgaben']"
    >
      <template #actions>
        <Button variant="outline" size="sm" @click="reload">
          <RotateCw class="h-4 w-4" /> Aktualisieren
        </Button>
        <Button v-if="auth.user?.admin" variant="primary" size="sm" @click="showForm = !showForm">
          <Plus class="h-4 w-4" /> Cron-Job
        </Button>
      </template>
    </PageHeader>

    <!-- Cron anlegen -->
    <form
      v-if="showForm"
      class="mb-5 grid gap-3 rounded-xl border border-border bg-card p-4 shadow-sm sm:grid-cols-2"
      @submit.prevent="createCron"
    >
      <div class="space-y-1.5">
        <label class="text-sm font-medium">Name</label>
        <Input v-model="form.name" placeholder="z. B. backup (nur a-z0-9-)" />
      </div>
      <div class="space-y-1.5">
        <label class="text-sm font-medium">Zeitplan (cron)</label>
        <Input v-model="form.schedule" placeholder="0 * * * *" />
      </div>
      <div class="space-y-1.5">
        <label class="text-sm font-medium">Benutzer</label>
        <Input v-model="form.user" placeholder="root" />
      </div>
      <div class="space-y-1.5">
        <label class="text-sm font-medium">Kommando</label>
        <Input v-model="form.command" placeholder="/usr/local/bin/backup.sh" />
      </div>
      <div class="flex justify-end gap-2 sm:col-span-2">
        <Button variant="outline" type="button" @click="showForm = false">Abbrechen</Button>
        <Button variant="primary" type="submit" :disabled="!form.name || !form.command">Anlegen</Button>
      </div>
    </form>

    <DataState :loading="loading" :error="error">
      <div v-if="data" class="space-y-6">
        <section class="space-y-3">
          <h2 class="flex items-center gap-2 px-1 text-sm font-semibold">
            <TimerIcon class="h-4 w-4" /> systemd-Timer
          </h2>
          <div
            v-if="data.timers?.length"
            class="overflow-hidden rounded-xl border border-border bg-card shadow-sm"
          >
            <div
              v-for="t in data.timers"
              :key="t.unit"
              class="border-b border-border/60 px-4 py-3 last:border-0"
            >
              <div class="flex items-center justify-between gap-4">
                <span class="truncate text-sm font-medium">{{ t.unit }}</span>
                <span class="shrink-0 text-xs text-muted-foreground">in {{ t.left || "—" }}</span>
              </div>
              <p class="truncate text-xs text-muted-foreground">
                Nächster Lauf: {{ t.next || "—" }} · aktiviert {{ t.activates }}
              </p>
            </div>
          </div>
          <p v-else class="px-1 text-sm text-muted-foreground">Keine Timer vorhanden.</p>
        </section>

        <section class="space-y-3">
          <h2 class="flex items-center gap-2 px-1 text-sm font-semibold">
            <CalendarClock class="h-4 w-4" /> Cron-Jobs
          </h2>
          <div
            v-if="data.cron?.length"
            class="overflow-hidden rounded-xl border border-border bg-card shadow-sm"
          >
            <div
              v-for="(c, i) in data.cron"
              :key="i"
              class="flex items-center gap-4 border-b border-border/60 px-4 py-3 last:border-0"
            >
              <code class="shrink-0 rounded bg-secondary px-2 py-0.5 font-mono text-xs">{{ c.schedule }}</code>
              <span v-if="c.user" class="shrink-0 text-xs text-muted-foreground">{{ c.user }}</span>
              <span class="min-w-0 flex-1 truncate font-mono text-xs">{{ c.command }}</span>
              <Button
                v-if="auth.user?.admin && managedName(c.source)"
                variant="ghost"
                size="icon"
                title="Löschen"
                @click="deleteCron(managedName(c.source)!)"
              >
                <Trash2 class="h-4 w-4 text-destructive" />
              </Button>
            </div>
          </div>
          <p v-else class="px-1 text-sm text-muted-foreground">Keine System-Cron-Jobs gefunden.</p>
        </section>
      </div>
    </DataState>
  </div>
</template>
