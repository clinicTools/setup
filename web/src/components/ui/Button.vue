<script setup lang="ts">
import { computed } from "vue";
import { cn } from "@/lib/utils";

const props = withDefaults(
  defineProps<{
    variant?: "primary" | "secondary" | "outline" | "ghost" | "destructive";
    size?: "sm" | "md" | "icon";
    disabled?: boolean;
    type?: "button" | "submit";
  }>(),
  { variant: "secondary", size: "md", disabled: false, type: "button" },
);

const classes = computed(() =>
  cn(
    "inline-flex items-center justify-center gap-2 rounded-md font-medium whitespace-nowrap transition-colors",
    "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1 focus-visible:ring-offset-background",
    "disabled:pointer-events-none disabled:opacity-50",
    {
      primary: "bg-primary text-primary-foreground hover:bg-primary/90 shadow-sm",
      secondary: "bg-secondary text-secondary-foreground hover:bg-secondary/70 border border-border",
      outline: "border border-border bg-card hover:bg-accent hover:text-accent-foreground",
      ghost: "hover:bg-accent hover:text-accent-foreground",
      destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90 shadow-sm",
    }[props.variant],
    {
      sm: "h-8 px-3 text-xs",
      md: "h-9 px-4 text-sm",
      icon: "h-9 w-9",
    }[props.size],
  ),
);
</script>

<template>
  <button :type="type" :class="classes" :disabled="disabled">
    <slot />
  </button>
</template>
