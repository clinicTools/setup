<script setup lang="ts">
import { computed } from "vue";
import { cn } from "@/lib/utils";

const props = withDefaults(
  defineProps<{ variant?: "neutral" | "success" | "warning" | "destructive" | "info" }>(),
  { variant: "neutral" },
);

const classes = computed(() =>
  cn(
    "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium",
    {
      neutral: "bg-secondary text-secondary-foreground",
      success: "bg-success/15 text-success",
      warning: "bg-warning/20 text-warning",
      destructive: "bg-destructive/15 text-destructive",
      info: "bg-primary/10 text-primary",
    }[props.variant],
  ),
);
</script>

<template>
  <span :class="classes"><slot /></span>
</template>
