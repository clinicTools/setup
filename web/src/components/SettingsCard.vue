<script setup lang="ts">
import type { Component } from "vue";

/**
 * SettingsCard — angelehnt an das Windows-11-SettingsCard-Control:
 * Icon links, Titel + Beschreibung in der Mitte, Aktionssteuerung
 * (Slot) rechtsbündig. Abgerundete Karte mit dezentem Rahmen.
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
      'flex w-full items-center gap-4 rounded-lg border border-border bg-card px-4 py-3 text-left transition-colors',
      clickable ? 'hover:bg-accent/60 cursor-pointer' : '',
    ]"
  >
    <div
      v-if="icon"
      class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-secondary text-foreground/80"
    >
      <component :is="icon" class="h-[18px] w-[18px]" />
    </div>
    <div class="min-w-0 flex-1">
      <div v-if="title" class="truncate text-sm font-medium text-foreground">{{ title }}</div>
      <div v-if="description" class="truncate text-xs text-muted-foreground">{{ description }}</div>
      <slot name="content" />
    </div>
    <div class="flex shrink-0 items-center gap-2">
      <slot name="action" />
    </div>
  </component>
</template>
