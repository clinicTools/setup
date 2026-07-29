<script setup lang="ts">
// Gruppiert mehrere SettingsCards unter einer Abschnittsüberschrift
// (Windows-11-Muster: gruppierte Einstellungen mit Section-Header).
defineProps<{ title?: string; description?: string }>();
</script>

<template>
  <section class="space-y-2">
    <div v-if="title" class="px-1">
      <h2 class="text-sm font-semibold text-foreground">{{ title }}</h2>
      <p v-if="description" class="text-xs text-muted-foreground">{{ description }}</p>
    </div>
    <div class="space-y-1.5">
      <slot />
    </div>
  </section>
</template>
