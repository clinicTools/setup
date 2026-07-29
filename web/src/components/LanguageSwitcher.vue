<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { Languages } from "lucide-vue-next";
import { setLocale, type Locale } from "@/i18n";
import Button from "./ui/Button.vue";

const { locale } = useI18n();

function toggle(): void {
  setLocale((locale.value === "de" ? "en" : "de") as Locale);
}
</script>

<template>
  <Button variant="ghost" size="sm" :title="'Sprache / Language'" @click="toggle">
    <Languages class="h-4 w-4" />
    <span class="uppercase">{{ locale }}</span>
  </Button>
</template>
