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
  Boxes,
  ClipboardList,
} from "lucide-vue-next";

export interface NavItem {
  to: string;
  /** i18n-Schlüssel des Labels (nav.items.*). */
  labelKey: string;
  icon: Component;
  /** i18n-Schlüssel der Gruppenüberschrift (nav.groups.*). */
  groupKey: string;
}

// Navigationsstruktur — gruppiert wie in den Windows-11-Einstellungen.
export const navItems: NavItem[] = [
  { to: "/", labelKey: "dashboard", icon: LayoutDashboard, groupKey: "system" },
  { to: "/processes", labelKey: "processes", icon: Activity, groupKey: "system" },
  { to: "/services", labelKey: "services", icon: Cog, groupKey: "system" },
  { to: "/users", labelKey: "users", icon: Users, groupKey: "accounts" },
  { to: "/network", labelKey: "network", icon: Network, groupKey: "connections" },
  { to: "/storage", labelKey: "storage", icon: HardDrive, groupKey: "devices" },
  { to: "/packages", labelKey: "packages", icon: Package, groupKey: "apps" },
  { to: "/k3s", labelKey: "k3s", icon: Boxes, groupKey: "container" },
  { to: "/firewall", labelKey: "firewall", icon: ShieldCheck, groupKey: "security" },
  { to: "/audit", labelKey: "audit", icon: ClipboardList, groupKey: "security" },
  { to: "/scheduled", labelKey: "scheduled", icon: CalendarClock, groupKey: "tasks" },
  { to: "/datetime", labelKey: "datetime", icon: Clock, groupKey: "time" },
  { to: "/logs", labelKey: "logs", icon: ScrollText, groupKey: "diagnostics" },
];
