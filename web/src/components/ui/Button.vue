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

// Windows-11-/Fluent-Anmutung: 4-px-Radius, dezente dunklere Unterkante statt
// Schlagschatten, kurze Anfassanimation beim Drücken.
const classes = computed(() =>
  cn(
    "inline-flex select-none items-center justify-center gap-2 rounded font-medium whitespace-nowrap",
    "transition-all duration-100 active:scale-[0.98]",
    "disabled:pointer-events-none disabled:opacity-40",
    {
      primary: "bg-primary text-primary-foreground hover:bg-primary/90 fluent-edge",
      secondary:
        "bg-card text-secondary-foreground hover:bg-accent/60 border border-border fluent-edge",
      outline: "border border-border bg-card hover:bg-accent/60 fluent-edge",
      ghost: "hover:bg-accent/60 hover:text-accent-foreground",
      destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90 fluent-edge",
    }[props.variant],
    {
      sm: "h-8 px-3 text-xs",
      md: "h-8 px-4 text-[13px]",
      icon: "h-8 w-8",
    }[props.size],
  ),
);
</script>

<template>
  <button :type="type" :class="classes" :disabled="disabled">
    <slot />
  </button>
</template>
