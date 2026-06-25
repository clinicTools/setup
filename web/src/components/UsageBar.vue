<script setup lang="ts">
import { computed } from "vue";

/** Fortschritts-/Auslastungsbalken mit farblicher Schwellenwertlogik. */
const props = withDefaults(defineProps<{ percent: number; label?: string }>(), { percent: 0 });

const clamped = computed(() => Math.min(100, Math.max(0, props.percent)));
const color = computed(() => {
  if (clamped.value >= 90) return "bg-destructive";
  if (clamped.value >= 75) return "bg-warning";
  return "bg-primary";
});
</script>

<template>
  <div class="space-y-1">
    <div v-if="label" class="flex justify-between text-xs text-muted-foreground">
      <span>{{ label }}</span>
      <span>{{ clamped.toFixed(0) }} %</span>
    </div>
    <div class="h-2 w-full overflow-hidden rounded-full bg-secondary">
      <div :class="['h-full rounded-full transition-all', color]" :style="{ width: clamped + '%' }" />
    </div>
  </div>
</template>
