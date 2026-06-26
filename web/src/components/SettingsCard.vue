<script setup lang="ts">
import type { Component } from "vue";
import { ChevronRight } from "lucide-vue-next";

/**
 * SettingsCard — angelehnt an das Windows-11-SettingsCard-Control:
 * Icon-Kachel links, Titel + Beschreibung in der Mitte, Aktionssteuerung
 * (Slot) bzw. Chevron rechtsbündig. Abgerundete Karte mit dezentem Rahmen und
 * Hover-Zustand bei navigierbaren Karten.
 */
withDefaults(
  defineProps<{
    icon?: Component;
    title?: string;
    description?: string;
    clickable?: boolean;
  }>(),
  { clickable: false },
);
</script>

<template>
  <component
    :is="clickable ? 'button' : 'div'"
    :class="[
      'flex w-full items-center gap-3.5 rounded-xl border border-border/70 bg-card px-4 py-3 text-left shadow-sm transition-all',
      clickable ? 'cursor-pointer hover:border-border hover:bg-accent/40 active:scale-[0.997]' : '',
    ]"
  >
    <div
      v-if="icon"
      class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary"
    >
      <component :is="icon" class="h-5 w-5" />
    </div>
    <div class="min-w-0 flex-1">
      <div v-if="title" class="truncate text-sm font-medium text-foreground">{{ title }}</div>
      <div v-if="description" class="truncate text-xs text-muted-foreground">{{ description }}</div>
      <slot name="content" />
    </div>
    <div class="flex shrink-0 items-center gap-2">
      <slot name="action" />
      <ChevronRight v-if="clickable" class="h-4 w-4 text-muted-foreground" />
    </div>
  </component>
</template>
