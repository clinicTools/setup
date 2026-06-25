<script setup lang="ts">
import type { Component } from "vue";

/** Kachel für die Dashboard-Übersicht (Wert + Label + Icon). */
defineProps<{
  label: string;
  value: string;
  sub?: string;
  icon: Component;
  accent?: boolean;
}>();
</script>

<template>
  <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
    <div class="flex items-center justify-between">
      <span class="text-xs font-medium text-muted-foreground">{{ label }}</span>
      <component
        :is="icon"
        :class="['h-4 w-4', accent ? 'text-primary' : 'text-muted-foreground']"
      />
    </div>
    <div class="mt-2 text-2xl font-semibold tracking-tight text-foreground">{{ value }}</div>
    <div v-if="sub" class="mt-0.5 text-xs text-muted-foreground">{{ sub }}</div>
  </div>
</template>
