import { ref, onMounted, onUnmounted, watch, type Ref } from "vue";
import { ws } from "@/lib/ws";

interface ChannelOptions {
  /** Reaktive Parameter; bei Änderung wird neu subscribed. */
  params?: Ref<unknown>;
  /** Sammelt eingehende Nachrichten in einem Ring-Puffer dieser Größe. */
  buffer?: number;
}

interface ChannelState<T> {
  data: Ref<T | null>;
  history: Ref<T[]>;
  error: Ref<string | null>;
}

/**
 * Abonniert einen WebSocket-Channel für die Lebensdauer der Komponente:
 * subscribe beim Mounten, unsubscribe beim Unmounten. Damit läuft der
 * serverseitige Collector — wie bei Cockpit — nur, solange die Seite offen ist.
 */
export function useChannel<T = unknown>(
  channel: string,
  options: ChannelOptions = {},
): ChannelState<T> {
  const data = ref<T | null>(null) as Ref<T | null>;
  const history = ref<T[]>([]) as Ref<T[]>;
  const error = ref<string | null>(null);

  let unsubscribe: (() => void) | null = null;

  function start(): void {
    stop();
    error.value = null;
    unsubscribe = ws.subscribe(
      channel,
      options.params?.value,
      (payload) => {
        data.value = payload as T;
        if (options.buffer) {
          history.value.push(payload as T);
          if (history.value.length > options.buffer) {
            history.value.splice(0, history.value.length - options.buffer);
          }
        }
      },
      (e) => (error.value = e),
    );
  }

  function stop(): void {
    unsubscribe?.();
    unsubscribe = null;
  }

  onMounted(start);
  onUnmounted(stop);

  if (options.params) {
    watch(options.params, start, { deep: true });
  }

  return { data, history, error };
}
