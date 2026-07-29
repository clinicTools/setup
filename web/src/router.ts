import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import AppShell from "@/layouts/AppShell.vue";

const routes: RouteRecordRaw[] = [
  {
    path: "/login",
    name: "login",
    component: () => import("@/views/Login.vue"),
    meta: { public: true },
  },
  {
    path: "/",
    component: AppShell,
    children: [
      { path: "", name: "dashboard", component: () => import("@/views/Dashboard.vue") },
      { path: "processes", name: "processes", component: () => import("@/views/Processes.vue") },
      { path: "services", name: "services", component: () => import("@/views/Services.vue") },
      { path: "users", name: "users", component: () => import("@/views/Users.vue") },
      { path: "network", name: "network", component: () => import("@/views/Network.vue") },
      { path: "storage", name: "storage", component: () => import("@/views/Storage.vue") },
      { path: "packages", name: "packages", component: () => import("@/views/Packages.vue") },
      { path: "containers", name: "containers", component: () => import("@/views/Containers.vue") },
      { path: "stacks", name: "stacks", component: () => import("@/views/Stacks.vue") },
      { path: "firewall", name: "firewall", component: () => import("@/views/Firewall.vue") },
      { path: "audit", name: "audit", component: () => import("@/views/Audit.vue") },
      { path: "scheduled", name: "scheduled", component: () => import("@/views/Scheduled.vue") },
      { path: "datetime", name: "datetime", component: () => import("@/views/DateTime.vue") },
      { path: "logs", name: "logs", component: () => import("@/views/Logs.vue") },
    ],
  },
  { path: "/:pathMatch(.*)*", redirect: "/" },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

// Auth-Guard: ungeschützte Routen frei, sonst gültige Session erforderlich.
router.beforeEach(async (to) => {
  const auth = useAuthStore();
  if (!auth.ready) await auth.fetchMe();

  if (to.meta.public) {
    if (auth.user && to.name === "login") return { path: "/" };
    return true;
  }
  if (!auth.user) return { path: "/login", query: { redirect: to.fullPath } };
  return true;
});
