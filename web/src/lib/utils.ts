import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

/** cn führt bedingte Klassen zusammen und löst Tailwind-Konflikte auf. */
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}

/** Formatiert Bytes menschenlesbar (KB/MB/GB/TB, Basis 1024). */
export function formatBytes(bytes: number, digits = 1): string {
  if (!bytes || bytes < 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(i === 0 ? 0 : digits)} ${units[i]}`;
}

/** Wandelt Sekunden in eine kompakte Uptime-Angabe (z. B. "3 T 4 Std"). */
export function formatUptime(seconds: number): string {
  if (!seconds) return "—";
  const d = Math.floor(seconds / 86400);
  const h = Math.floor((seconds % 86400) / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const parts: string[] = [];
  if (d) parts.push(`${d} T`);
  if (h) parts.push(`${h} Std`);
  if (!d && m) parts.push(`${m} Min`);
  return parts.join(" ") || "< 1 Min";
}

/** Formatiert ISO-Zeitstempel als deutsches Datum/Uhrzeit. */
export function formatDateTime(iso: string): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  return d.toLocaleString("de-DE", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}
