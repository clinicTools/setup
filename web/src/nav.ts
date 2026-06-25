import type { Component } from "vue";
import {
  LayoutDashboard,
  Users,
  Cog,
  Package,
  Network,
  HardDrive,
  ShieldCheck,
  ScrollText,
  Clock,
  CalendarClock,
  Activity,
} from "lucide-vue-next";

export interface NavItem {
  to: string;
  label: string;
  icon: Component;
  /** Gruppenüberschrift in der Navigationsleiste. */
  group: string;
}

// Navigationsstruktur — gruppiert wie in den Windows-11-Einstellungen.
export const navItems: NavItem[] = [
  { to: "/", label: "Übersicht", icon: LayoutDashboard, group: "System" },
  { to: "/processes", label: "Prozesse", icon: Activity, group: "System" },
  { to: "/services", label: "Dienste", icon: Cog, group: "System" },
  { to: "/users", label: "Benutzer & Gruppen", icon: Users, group: "Konten" },
  { to: "/network", label: "Netzwerk", icon: Network, group: "Verbindungen" },
  { to: "/storage", label: "Speicher", icon: HardDrive, group: "Geräte" },
  { to: "/packages", label: "Pakete & Updates", icon: Package, group: "Apps" },
  { to: "/firewall", label: "Firewall", icon: ShieldCheck, group: "Sicherheit" },
  { to: "/scheduled", label: "Geplante Aufgaben", icon: CalendarClock, group: "Aufgaben" },
  { to: "/datetime", label: "Datum & Uhrzeit", icon: Clock, group: "Zeit & Sprache" },
  { to: "/logs", label: "Systemprotokolle", icon: ScrollText, group: "Diagnose" },
];
