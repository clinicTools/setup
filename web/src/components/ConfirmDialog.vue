<script setup lang="ts">
import Button from "./ui/Button.vue";

/** Modaler Bestätigungsdialog für destruktive Aktionen. */
defineProps<{
  open: boolean;
  title: string;
  message: string;
  confirmLabel?: string;
  destructive?: boolean;
}>();
const emit = defineEmits<{ (e: "confirm"): void; (e: "cancel"): void }>();
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      @click.self="emit('cancel')"
    >
      <div class="w-full max-w-md rounded-xl border border-border bg-popover p-6 shadow-xl">
        <h3 class="text-lg font-semibold text-foreground">{{ title }}</h3>
        <p class="mt-2 text-sm text-muted-foreground">{{ message }}</p>
        <div class="mt-6 flex justify-end gap-2">
          <Button variant="outline" @click="emit('cancel')">Abbrechen</Button>
          <Button :variant="destructive ? 'destructive' : 'primary'" @click="emit('confirm')">
            {{ confirmLabel || "Bestätigen" }}
          </Button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
