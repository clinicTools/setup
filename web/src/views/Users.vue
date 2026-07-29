<script setup lang="ts">
import { computed, ref } from "vue";
import { UserPlus, Trash2, KeyRound, Users as UsersIcon, RotateCw, Pencil, FolderPlus } from "lucide-vue-next";
import { api } from "@/lib/api";
import { useAsyncData } from "@/composables/useAsyncData";
import { useToast } from "@/composables/useToast";
import { useAuthStore } from "@/stores/auth";
import type { SystemUser } from "@/lib/types";
import PageHeader from "@/components/PageHeader.vue";
import DataState from "@/components/DataState.vue";
import Button from "@/components/ui/Button.vue";
import Input from "@/components/ui/Input.vue";
import Badge from "@/components/ui/Badge.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";

const toast = useToast();
const auth = useAuthStore();
const tab = ref<"users" | "groups">("users");
const showSystem = ref(false);

const { data: users, loading, error, reload } = useAsyncData(() => api.users());
const { data: groups, reload: reloadGroups } = useAsyncData(() => api.groups());

const visibleUsers = computed(() =>
  (users.value ?? []).filter((u) => showSystem.value || !u.system),
);
const visibleGroups = computed(() =>
  (groups.value ?? []).filter((g) => showSystem.value || !g.system),
);

// Benutzer anlegen
const showCreate = ref(false);
const form = ref({ username: "", fullName: "", password: "", shell: "/bin/bash" });
const creating = ref(false);

async function createUser(): Promise<void> {
  creating.value = true;
  try {
    await api.createUser({ ...form.value });
    toast.success(`Benutzer „${form.value.username}" angelegt`);
    showCreate.value = false;
    form.value = { username: "", fullName: "", password: "", shell: "/bin/bash" };
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  } finally {
    creating.value = false;
  }
}

// Benutzer löschen
const deleteTarget = ref<SystemUser | null>(null);
async function confirmDelete(): Promise<void> {
  const u = deleteTarget.value;
  deleteTarget.value = null;
  if (!u) return;
  try {
    await api.deleteUser(u.username, true);
    toast.success(`Benutzer „${u.username}" gelöscht`);
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  }
}

// Passwort setzen
const pwTarget = ref<SystemUser | null>(null);
const newPassword = ref("");
async function setPassword(): Promise<void> {
  const u = pwTarget.value;
  if (!u || !newPassword.value) return;
  try {
    await api.setPassword(u.username, newPassword.value);
    toast.success(`Passwort für „${u.username}" gesetzt`);
  } catch (e) {
    toast.error((e as Error).message);
  } finally {
    pwTarget.value = null;
    newPassword.value = "";
  }
}

function refreshAll(): void {
  reload();
  reloadGroups();
}

// Benutzer bearbeiten (Gruppen/Shell)
const editTarget = ref<SystemUser | null>(null);
const editForm = ref({ groups: "", shell: "" });
function openEdit(u: SystemUser): void {
  editTarget.value = u;
  editForm.value = { groups: (u.groups || []).join(", "), shell: u.shell };
}
async function saveEdit(): Promise<void> {
  const u = editTarget.value;
  if (!u) return;
  const groups = editForm.value.groups.split(/[\s,]+/).filter(Boolean);
  try {
    await api.modifyUser(u.username, groups, editForm.value.shell);
    toast.success(`Benutzer „${u.username}" aktualisiert`);
    editTarget.value = null;
    await reload();
  } catch (e) {
    toast.error((e as Error).message);
  }
}

// Gruppe anlegen
const showGroup = ref(false);
const groupForm = ref({ name: "", system: false });
async function createGroup(): Promise<void> {
  try {
    await api.createGroup(groupForm.value.name, groupForm.value.system);
    toast.success(`Gruppe „${groupForm.value.name}" angelegt`);
    showGroup.value = false;
    groupForm.value = { name: "", system: false };
    await reloadGroups();
  } catch (e) {
    toast.error((e as Error).message);
  }
}
const deleteGroupTarget = ref<string | null>(null);
async function confirmDeleteGroup(): Promise<void> {
  const name = deleteGroupTarget.value;
  deleteGroupTarget.value = null;
  if (!name) return;
  try {
    await api.deleteGroup(name);
    toast.success(`Gruppe „${name}" gelöscht`);
    await reloadGroups();
  } catch (e) {
    toast.error((e as Error).message);
  }
}
</script>

<template>
  <div>
    <PageHeader
      title="Benutzer & Gruppen"
      description="Systemkonten und Gruppenzugehörigkeiten verwalten"
      :breadcrumb="['Konten', 'Benutzer & Gruppen']"
    >
      <template #actions>
        <Button variant="outline" size="sm" @click="refreshAll">
          <RotateCw class="h-4 w-4" /> Aktualisieren
        </Button>
        <Button v-if="auth.user?.admin && tab === 'groups'" variant="primary" size="sm" @click="showGroup = true">
          <FolderPlus class="h-4 w-4" /> Gruppe
        </Button>
        <Button v-if="auth.user?.admin && tab === 'users'" variant="primary" size="sm" @click="showCreate = true">
          <UserPlus class="h-4 w-4" /> Benutzer
        </Button>
      </template>
    </PageHeader>

    <div class="mb-4 flex items-center justify-between">
      <div class="inline-flex rounded-lg border border-border bg-secondary/50 p-0.5">
        <button
          v-for="t in (['users', 'groups'] as const)"
          :key="t"
          class="rounded-md px-4 py-1.5 text-sm font-medium transition-colors"
          :class="tab === t ? 'bg-card shadow-sm text-foreground' : 'text-muted-foreground'"
          @click="tab = t"
        >
          {{ t === "users" ? "Benutzer" : "Gruppen" }}
        </button>
      </div>
      <label class="flex items-center gap-2 text-sm text-muted-foreground">
        <input v-model="showSystem" type="checkbox" class="h-4 w-4 rounded border-input" />
        Systemkonten anzeigen
      </label>
    </div>

    <DataState :loading="loading" :error="error">
      <!-- Benutzer -->
      <div v-if="tab === 'users'" class="overflow-hidden rounded-xl border border-border bg-card shadow-sm">
        <div
          v-for="u in visibleUsers"
          :key="u.uid"
          class="flex items-center gap-4 border-b border-border/60 px-4 py-3 last:border-0"
        >
          <div class="flex h-9 w-9 items-center justify-center rounded-full bg-primary/10 text-xs font-semibold text-primary">
            {{ (u.fullName || u.username).slice(0, 2).toUpperCase() }}
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="truncate text-sm font-medium">{{ u.username }}</span>
              <Badge v-if="u.system" variant="neutral">System</Badge>
              <span class="text-xs text-muted-foreground">UID {{ u.uid }}</span>
            </div>
            <p class="truncate text-xs text-muted-foreground">
              {{ u.fullName || "—" }} · {{ u.shell }} · {{ (u.groups || []).join(", ") || "keine Gruppen" }}
            </p>
          </div>
          <div v-if="auth.user?.admin && !u.system" class="flex shrink-0 items-center gap-1">
            <Button variant="ghost" size="icon" title="Bearbeiten" @click="openEdit(u)">
              <Pencil class="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon" title="Passwort setzen" @click="pwTarget = u">
              <KeyRound class="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon" title="Löschen" @click="deleteTarget = u">
              <Trash2 class="h-4 w-4 text-destructive" />
            </Button>
          </div>
        </div>
      </div>

      <!-- Gruppen -->
      <div v-else class="overflow-hidden rounded-xl border border-border bg-card shadow-sm">
        <div
          v-for="g in visibleGroups"
          :key="g.gid"
          class="flex items-center gap-4 border-b border-border/60 px-4 py-3 last:border-0"
        >
          <div class="flex h-9 w-9 items-center justify-center rounded-md bg-secondary text-foreground/70">
            <UsersIcon class="h-4 w-4" />
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="truncate text-sm font-medium">{{ g.name }}</span>
              <Badge v-if="g.system" variant="neutral">System</Badge>
              <span class="text-xs text-muted-foreground">GID {{ g.gid }}</span>
            </div>
            <p class="truncate text-xs text-muted-foreground">
              {{ (g.members || []).join(", ") || "keine Mitglieder" }}
            </p>
          </div>
          <Button
            v-if="auth.user?.admin && !g.system"
            variant="ghost"
            size="icon"
            title="Gruppe löschen"
            @click="deleteGroupTarget = g.name"
          >
            <Trash2 class="h-4 w-4 text-destructive" />
          </Button>
        </div>
      </div>
    </DataState>

    <!-- Anlegen-Dialog -->
    <Teleport to="body">
      <div
        v-if="showCreate"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
        @click.self="showCreate = false"
      >
        <form
          class="w-full max-w-md space-y-4 rounded-xl border border-border bg-popover p-6 shadow-xl"
          @submit.prevent="createUser"
        >
          <h3 class="text-lg font-semibold">Neuen Benutzer anlegen</h3>
          <div class="space-y-1.5">
            <label class="text-sm font-medium">Benutzername</label>
            <Input v-model="form.username" placeholder="z. B. maria" />
          </div>
          <div class="space-y-1.5">
            <label class="text-sm font-medium">Vollständiger Name</label>
            <Input v-model="form.fullName" placeholder="Maria Muster" />
          </div>
          <div class="space-y-1.5">
            <label class="text-sm font-medium">Passwort</label>
            <Input v-model="form.password" type="password" />
          </div>
          <div class="space-y-1.5">
            <label class="text-sm font-medium">Login-Shell</label>
            <Input v-model="form.shell" />
          </div>
          <div class="flex justify-end gap-2 pt-2">
            <Button variant="outline" type="button" @click="showCreate = false">Abbrechen</Button>
            <Button variant="primary" type="submit" :disabled="creating || !form.username">
              {{ creating ? "Wird angelegt …" : "Anlegen" }}
            </Button>
          </div>
        </form>
      </div>
    </Teleport>

    <!-- Passwort-Dialog -->
    <Teleport to="body">
      <div
        v-if="pwTarget"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
        @click.self="pwTarget = null"
      >
        <form
          class="w-full max-w-sm space-y-4 rounded-xl border border-border bg-popover p-6 shadow-xl"
          @submit.prevent="setPassword"
        >
          <h3 class="text-lg font-semibold">Passwort für „{{ pwTarget.username }}“</h3>
          <Input v-model="newPassword" type="password" placeholder="Neues Passwort" />
          <div class="flex justify-end gap-2">
            <Button variant="outline" type="button" @click="pwTarget = null">Abbrechen</Button>
            <Button variant="primary" type="submit" :disabled="!newPassword">Setzen</Button>
          </div>
        </form>
      </div>
    </Teleport>

    <!-- Benutzer bearbeiten -->
    <Teleport to="body">
      <div
        v-if="editTarget"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
        @click.self="editTarget = null"
      >
        <form
          class="w-full max-w-md space-y-4 rounded-xl border border-border bg-popover p-6 shadow-xl"
          @submit.prevent="saveEdit"
        >
          <h3 class="text-lg font-semibold">„{{ editTarget.username }}“ bearbeiten</h3>
          <div class="space-y-1.5">
            <label class="text-sm font-medium">Gruppen (kommagetrennt)</label>
            <Input v-model="editForm.groups" placeholder="sudo, docker" />
          </div>
          <div class="space-y-1.5">
            <label class="text-sm font-medium">Login-Shell</label>
            <Input v-model="editForm.shell" />
          </div>
          <div class="flex justify-end gap-2 pt-2">
            <Button variant="outline" type="button" @click="editTarget = null">Abbrechen</Button>
            <Button variant="primary" type="submit">Speichern</Button>
          </div>
        </form>
      </div>
    </Teleport>

    <!-- Gruppe anlegen -->
    <Teleport to="body">
      <div
        v-if="showGroup"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
        @click.self="showGroup = false"
      >
        <form
          class="w-full max-w-sm space-y-4 rounded-xl border border-border bg-popover p-6 shadow-xl"
          @submit.prevent="createGroup"
        >
          <h3 class="text-lg font-semibold">Neue Gruppe</h3>
          <Input v-model="groupForm.name" placeholder="Gruppenname" />
          <label class="flex items-center gap-2 text-sm text-muted-foreground">
            <input v-model="groupForm.system" type="checkbox" class="h-4 w-4 rounded border-input" />
            Systemgruppe
          </label>
          <div class="flex justify-end gap-2">
            <Button variant="outline" type="button" @click="showGroup = false">Abbrechen</Button>
            <Button variant="primary" type="submit" :disabled="!groupForm.name">Anlegen</Button>
          </div>
        </form>
      </div>
    </Teleport>

    <ConfirmDialog
      :open="deleteTarget !== null"
      title="Benutzer löschen?"
      :message="`Der Benutzer „${deleteTarget?.username}“ und sein Home-Verzeichnis werden entfernt.`"
      confirm-label="Löschen"
      destructive
      @confirm="confirmDelete"
      @cancel="deleteTarget = null"
    />

    <ConfirmDialog
      :open="deleteGroupTarget !== null"
      title="Gruppe löschen?"
      :message="`Die Gruppe „${deleteGroupTarget}“ wird entfernt.`"
      confirm-label="Löschen"
      destructive
      @confirm="confirmDeleteGroup"
      @cancel="deleteGroupTarget = null"
    />
  </div>
</template>
