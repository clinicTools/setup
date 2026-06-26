<script setup lang="ts">
import type { Component } from "vue";
import { ChevronRight } from "lucide-vue-next";

/** Kachel für die Dashboard-Übersicht (Wert + Label + Icon). Mit `to` wird sie
 * zu einer navigierbaren „Hero-Control" (Windows-11-Muster). */
defineProps<{
  label: string;
  value: string;
  sub?: string;
  icon: Component;
  accent?: boolean;
  to?: string;
}>();
</script>

<template>
  <component
    :is="to ? 'RouterLink' : 'div'"
    :to="to"
    :class="[
      'group block rounded-xl border border-border/70 bg-card p-4 shadow-sm transition-all',
      to ? 'hover:border-border hover:bg-accent/40 active:scale-[0.997]' : '',
    ]"
  >
    <div class="flex items-center justify-between">
      <span class="text-xs font-medium text-muted-foreground">{{ label }}</span>
      <component
        :is="icon"
        :class="['h-4 w-4', accent ? 'text-primary' : 'text-muted-foreground']"
      />
    </div>
    <div class="mt-2 flex items-end justify-between">
      <div class="text-2xl font-semibold tracking-tight text-foreground">{{ value }}</div>
      <ChevronRight
        v-if="to"
        class="h-4 w-4 text-muted-foreground/60 transition-transform group-hover:translate-x-0.5"
      />
    </div>
    <div v-if="sub" class="mt-0.5 text-xs text-muted-foreground">{{ sub }}</div>
  </component>
</template>
