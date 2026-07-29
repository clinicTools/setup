<script setup lang="ts">
import { CheckCircle2, XCircle, Info } from "lucide-vue-next";
import { useToast } from "@/composables/useToast";

const { toasts, dismiss } = useToast();

const icons = { success: CheckCircle2, error: XCircle, info: Info };
const colors = {
  success: "text-success",
  error: "text-destructive",
  info: "text-primary",
};
</script>

<template>
  <div class="fixed bottom-4 right-4 z-[60] flex w-full max-w-sm flex-col gap-2">
    <TransitionGroup name="toast">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="flex items-start gap-3 rounded-lg border border-border bg-popover px-4 py-3 shadow-lg"
        @click="dismiss(t.id)"
      >
        <component :is="icons[t.variant]" :class="['mt-0.5 h-5 w-5 shrink-0', colors[t.variant]]" />
        <span class="text-sm text-foreground">{{ t.message }}</span>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.25s ease;
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(1rem);
}
</style>
