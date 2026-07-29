<script setup lang="ts">
import { Loader2, AlertTriangle, Inbox } from "lucide-vue-next";

/** Einheitliche Darstellung von Lade-, Fehler- und Leer-Zuständen. */
defineProps<{
  loading: boolean;
  error?: string | null;
  empty?: boolean;
  emptyText?: string;
}>();
</script>

<template>
  <div v-if="loading" class="flex items-center justify-center gap-2 py-16 text-muted-foreground">
    <Loader2 class="h-5 w-5 animate-spin" />
    <span class="text-sm">Wird geladen …</span>
  </div>
  <div
    v-else-if="error"
    class="flex items-center gap-3 rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive"
  >
    <AlertTriangle class="h-5 w-5 shrink-0" />
    <span>{{ error }}</span>
  </div>
  <div
    v-else-if="empty"
    class="flex flex-col items-center justify-center gap-2 py-16 text-muted-foreground"
  >
    <Inbox class="h-8 w-8" />
    <span class="text-sm">{{ emptyText || "Keine Einträge vorhanden." }}</span>
  </div>
  <slot v-else />
</template>
