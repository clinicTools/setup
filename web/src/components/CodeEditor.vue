<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import { EditorView, basicSetup } from "codemirror";
import { EditorState, Compartment } from "@codemirror/state";
import { yaml } from "@codemirror/lang-yaml";
import { oneDark } from "@codemirror/theme-one-dark";
import { useTheme } from "@/composables/useTheme";

/**
 * CodeMirror-6-Editor für compose.yaml.
 *
 * Das Theme wird über ein Compartment reaktiv umgeschaltet (Hell/Dunkel), ohne
 * den Editor neu aufzubauen. Änderungen von außen werden nur eingespielt, wenn
 * sie vom aktuellen Dokument abweichen — so springt der Cursor beim Tippen nicht.
 */
const props = withDefaults(
  defineProps<{ modelValue: string; readonly?: boolean; minHeight?: string }>(),
  { readonly: false, minHeight: "22rem" },
);
const emit = defineEmits<{ (e: "update:modelValue", value: string): void }>();

const host = ref<HTMLElement | null>(null);
const { theme } = useTheme();

let view: EditorView | null = null;
const themeCompartment = new Compartment();
const readonlyCompartment = new Compartment();

/** Grundlayout des Editors (Schrift, Höhe, Rahmenlosigkeit). */
const baseTheme = EditorView.theme({
  "&": { fontSize: "12.5px", backgroundColor: "transparent" },
  "&.cm-focused": { outline: "none" },
  ".cm-scroller": {
    fontFamily:
      'ui-monospace, "Cascadia Code", "Segoe UI Mono", SFMono-Regular, Menlo, Consolas, monospace',
    lineHeight: "1.6",
    minHeight: props.minHeight,
  },
  ".cm-gutters": { backgroundColor: "transparent", border: "none", opacity: "0.6" },
  ".cm-activeLine": { backgroundColor: "color-mix(in oklab, currentColor 6%, transparent)" },
  ".cm-activeLineGutter": { backgroundColor: "transparent" },
});

function currentThemeExt() {
  return theme.value === "dark" ? oneDark : [];
}

onMounted(() => {
  if (!host.value) return;
  view = new EditorView({
    parent: host.value,
    state: EditorState.create({
      doc: props.modelValue,
      extensions: [
        basicSetup,
        yaml(),
        baseTheme,
        themeCompartment.of(currentThemeExt()),
        readonlyCompartment.of(EditorState.readOnly.of(props.readonly)),
        EditorView.updateListener.of((update) => {
          if (update.docChanged) emit("update:modelValue", update.state.doc.toString());
        }),
      ],
    }),
  });
});

onBeforeUnmount(() => {
  view?.destroy();
  view = null;
});

// Externe Änderungen einspielen (z. B. Laden einer anderen compose.yaml).
watch(
  () => props.modelValue,
  (value) => {
    if (!view || value === view.state.doc.toString()) return;
    view.dispatch({
      changes: { from: 0, to: view.state.doc.length, insert: value },
    });
  },
);

watch(theme, () => {
  view?.dispatch({ effects: themeCompartment.reconfigure(currentThemeExt()) });
});

watch(
  () => props.readonly,
  (ro) => {
    view?.dispatch({ effects: readonlyCompartment.reconfigure(EditorState.readOnly.of(ro)) });
  },
);
</script>

<template>
  <div
    ref="host"
    class="app-selectable overflow-hidden rounded-lg border border-border/70 bg-card/60"
  />
</template>
