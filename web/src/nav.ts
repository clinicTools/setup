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
  Layers,
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

// Navigationsstruktur — zu wenigen, klaren Kategorien zusammengefasst.
export const navItems: NavItem[] = [
  { to: "/", labelKey: "dashboard", icon: LayoutDashboard, groupKey: "system" },
  { to: "/processes", labelKey: "processes", icon: Activity, groupKey: "system" },
  { to: "/services", labelKey: "services", icon: Cog, groupKey: "system" },
  { to: "/datetime", labelKey: "datetime", icon: Clock, groupKey: "system" },
  { to: "/network", labelKey: "network", icon: Network, groupKey: "networkDevices" },
  { to: "/storage", labelKey: "storage", icon: HardDrive, groupKey: "networkDevices" },
  { to: "/packages", labelKey: "packages", icon: Package, groupKey: "appsContainers" },
  { to: "/containers", labelKey: "containers", icon: Boxes, groupKey: "appsContainers" },
  { to: "/stacks", labelKey: "stacks", icon: Layers, groupKey: "appsContainers" },
  { to: "/users", labelKey: "users", icon: Users, groupKey: "accountsSecurity" },
  { to: "/firewall", labelKey: "firewall", icon: ShieldCheck, groupKey: "accountsSecurity" },
  { to: "/audit", labelKey: "audit", icon: ClipboardList, groupKey: "accountsSecurity" },
  { to: "/scheduled", labelKey: "scheduled", icon: CalendarClock, groupKey: "tasksDiagnostics" },
  { to: "/logs", labelKey: "logs", icon: ScrollText, groupKey: "tasksDiagnostics" },
];
