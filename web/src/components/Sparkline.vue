<script setup lang="ts">
import { computed } from "vue";

/** Leichtgewichtiger SVG-Sparkline-Graph (ohne Chart-Bibliothek). */
const props = withDefaults(
  defineProps<{ values: number[]; max?: number; height?: number; width?: number }>(),
  { height: 40, width: 240 },
);

const points = computed(() => {
  const vals = props.values;
  if (vals.length < 2) return "";
  const max = props.max ?? Math.max(1, ...vals);
  const stepX = props.width / (vals.length - 1);
  return vals
    .map((v, i) => {
      const x = i * stepX;
      const y = props.height - (Math.min(v, max) / max) * props.height;
      return `${x.toFixed(1)},${y.toFixed(1)}`;
    })
    .join(" ");
});

const areaPoints = computed(() =>
  points.value ? `0,${props.height} ${points.value} ${props.width},${props.height}` : "",
);
</script>

<template>
  <svg
    :viewBox="`0 0 ${width} ${height}`"
    :width="width"
    :height="height"
    preserveAspectRatio="none"
    class="overflow-visible"
  >
    <polygon :points="areaPoints" fill="var(--primary)" fill-opacity="0.12" />
    <polyline
      :points="points"
      fill="none"
      stroke="var(--primary)"
      stroke-width="1.5"
      stroke-linejoin="round"
      stroke-linecap="round"
      vector-effect="non-scaling-stroke"
    />
  </svg>
</template>
