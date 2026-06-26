<script setup lang="ts">
import { ChevronRight } from "lucide-vue-next";

/**
 * Seitenkopf mit Breadcrumb (Windows-11-BreadcrumbBar-Muster), Titel,
 * optionaler Beschreibung und einem Aktions-Slot rechts.
 */
defineProps<{
  title: string;
  description?: string;
  breadcrumb?: string[];
}>();
</script>

<template>
  <header class="mb-6 space-y-2.5">
    <nav v-if="breadcrumb?.length" class="flex items-center gap-1 text-xs text-muted-foreground">
      <template v-for="(crumb, i) in breadcrumb" :key="i">
        <span :class="i === breadcrumb.length - 1 ? 'font-medium text-foreground' : ''">{{ crumb }}</span>
        <ChevronRight v-if="i < breadcrumb.length - 1" class="h-3 w-3" />
      </template>
    </nav>
    <div class="flex items-end justify-between gap-4">
      <div>
        <h1 class="text-[28px] font-semibold leading-tight tracking-tight text-foreground">{{ title }}</h1>
        <p v-if="description" class="mt-1 text-sm text-muted-foreground">{{ description }}</p>
      </div>
      <div class="flex shrink-0 items-center gap-2 pb-1">
        <slot name="actions" />
      </div>
    </div>
  </header>
</template>
