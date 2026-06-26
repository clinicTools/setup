<script setup lang="ts">
import { ref } from "vue";
import type { Component } from "vue";
import { ChevronDown } from "lucide-vue-next";

/**
 * Expander — angelehnt an das Windows-11-SettingsExpander-Control: eine
 * SettingsCard, die sich aufklappt und weitere Inhalte (Detailoptionen)
 * darunter einblendet.
 */
const props = withDefaults(
  defineProps<{
    icon?: Component;
    title?: string;
    description?: string;
    defaultOpen?: boolean;
  }>(),
  { defaultOpen: false },
);

const open = ref(props.defaultOpen);
</script>

<template>
  <div class="overflow-hidden rounded-xl border border-border/70 bg-card shadow-sm">
    <button
      type="button"
      class="flex w-full items-center gap-3.5 px-4 py-3 text-left transition-colors hover:bg-accent/40"
      :aria-expanded="open"
      @click="open = !open"
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
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <slot name="action" />
        <ChevronDown
          class="h-4 w-4 text-muted-foreground transition-transform"
          :class="open ? 'rotate-180' : ''"
        />
      </div>
    </button>
    <Transition
      enter-active-class="transition-all duration-200 ease-out"
      enter-from-class="opacity-0 -translate-y-1"
      leave-active-class="transition-all duration-150 ease-in"
      leave-to-class="opacity-0 -translate-y-1"
    >
      <div v-if="open" class="border-t border-border/60 px-4 py-3.5">
        <slot />
      </div>
    </Transition>
  </div>
</template>
