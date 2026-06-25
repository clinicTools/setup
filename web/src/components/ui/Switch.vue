<script setup lang="ts">
import { cn } from "@/lib/utils";

const props = defineProps<{ modelValue: boolean; disabled?: boolean }>();
const emit = defineEmits<{ (e: "update:modelValue", value: boolean): void }>();

function toggle(): void {
  if (!props.disabled) emit("update:modelValue", !props.modelValue);
}
</script>

<template>
  <button
    type="button"
    role="switch"
    :aria-checked="modelValue"
    :disabled="disabled"
    :class="
      cn(
        'relative inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full transition-colors',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background',
        'disabled:cursor-not-allowed disabled:opacity-50',
        modelValue ? 'bg-primary' : 'bg-input',
      )
    "
    @click="toggle"
  >
    <span
      :class="
        cn(
          'inline-block h-5 w-5 transform rounded-full bg-white shadow transition-transform',
          modelValue ? 'translate-x-5' : 'translate-x-0.5',
        )
      "
    />
  </button>
</template>
