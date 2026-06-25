import { ref } from "vue";

export interface Toast {
  id: number;
  message: string;
  variant: "success" | "error" | "info";
}

const toasts = ref<Toast[]>([]);
let counter = 0;

/** Einfache, global geteilte Toast-Benachrichtigungen. */
export function useToast() {
  function push(message: string, variant: Toast["variant"] = "info"): void {
    const id = ++counter;
    toasts.value.push({ id, message, variant });
    setTimeout(() => dismiss(id), 4000);
  }

  function dismiss(id: number): void {
    toasts.value = toasts.value.filter((t) => t.id !== id);
  }

  return {
    toasts,
    dismiss,
    success: (m: string) => push(m, "success"),
    error: (m: string) => push(m, "error"),
    info: (m: string) => push(m, "info"),
  };
}
