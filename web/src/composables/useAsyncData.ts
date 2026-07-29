import { ref, onMounted, type Ref } from "vue";
import { ApiError } from "@/lib/api";

interface AsyncState<T> {
  data: Ref<T | null>;
  loading: Ref<boolean>;
  error: Ref<string | null>;
  reload: () => Promise<void>;
}

/** Lädt asynchrone Daten mit Lade-/Fehlerzustand und Reload-Funktion. */
export function useAsyncData<T>(loader: () => Promise<T>, immediate = true): AsyncState<T> {
  const data = ref<T | null>(null) as Ref<T | null>;
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function reload(): Promise<void> {
    loading.value = true;
    error.value = null;
    try {
      data.value = await loader();
    } catch (e) {
      error.value = e instanceof ApiError ? e.message : (e as Error).message;
    } finally {
      loading.value = false;
    }
  }

  if (immediate) onMounted(reload);
  return { data, loading, error, reload };
}
