// Schlanker API-Client um fetch. Cookies (Session) werden mitgesendet.
import type {
  User,
  HostInfo,
  SystemUser,
  SystemGroup,
  Service,
  PackageSummary,
  Package,
  NetworkInfo,
  StorageInfo,
  FirewallStatus,
  LogEntry,
  ScheduledTasks,
  TimeInfo,
  Process,
} from "./types";

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
    this.name = "ApiError";
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(`/api${path}`, {
    method,
    credentials: "include",
    headers: body ? { "Content-Type": "application/json" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!res.ok) {
    let message = `HTTP ${res.status}`;
    try {
      const data = await res.json();
      if (data?.error) message = data.error;
    } catch {
      /* ignorieren */
    }
    throw new ApiError(res.status, message);
  }

  if (res.status === 204) return undefined as T;
  const text = await res.text();
  return text ? (JSON.parse(text) as T) : (undefined as T);
}

const get = <T>(p: string) => request<T>("GET", p);
const post = <T>(p: string, b?: unknown) => request<T>("POST", p, b);
const put = <T>(p: string, b?: unknown) => request<T>("PUT", p, b);
const del = <T>(p: string) => request<T>("DELETE", p);

export const api = {
  // Authentifizierung
  login: (username: string, password: string) =>
    post<User>("/auth/login", { username, password }),
  logout: () => post<{ ok: boolean }>("/auth/logout"),
  me: () => get<User>("/auth/me"),

  // System
  info: () => get<HostInfo>("/system/info"),
  users: () => get<SystemUser[]>("/system/users"),
  groups: () => get<SystemGroup[]>("/system/groups"),
  createUser: (body: Record<string, unknown>) => post<{ ok: boolean }>("/system/users", body),
  deleteUser: (name: string, removeHome: boolean) =>
    del<{ ok: boolean }>(`/system/users/${encodeURIComponent(name)}?removeHome=${removeHome}`),
  setPassword: (name: string, password: string) =>
    put<{ ok: boolean }>(`/system/users/${encodeURIComponent(name)}/password`, { password }),

  services: () => get<Service[]>("/system/services"),
  serviceStatus: (name: string) =>
    get<{ status: string }>(`/system/services/${encodeURIComponent(name)}/status`),
  serviceAction: (name: string, action: string) =>
    post<{ ok: boolean }>(`/system/services/${encodeURIComponent(name)}/action`, { action }),

  packages: () => get<PackageSummary>("/system/packages"),
  installedPackages: () => get<Package[]>("/system/packages/installed"),
  aptUpdate: () => post<{ output: string }>("/system/packages/update"),

  network: () => get<NetworkInfo>("/system/network"),
  storage: () => get<StorageInfo>("/system/storage"),

  firewall: () => get<FirewallStatus>("/system/firewall"),
  setFirewall: (enabled: boolean) => put<{ ok: boolean }>("/system/firewall", { enabled }),

  logs: (params: { unit?: string; priority?: string; lines?: number }) => {
    const q = new URLSearchParams();
    if (params.unit) q.set("unit", params.unit);
    if (params.priority) q.set("priority", params.priority);
    if (params.lines) q.set("lines", String(params.lines));
    return get<LogEntry[]>(`/system/logs?${q.toString()}`);
  },

  scheduled: () => get<ScheduledTasks>("/system/scheduled"),

  time: () => get<TimeInfo>("/system/time"),
  timezones: () => get<string[]>("/system/timezones"),
  setTimezone: (timezone: string) =>
    put<{ ok: boolean }>("/system/time/timezone", { timezone }),
  setNTP: (enabled: boolean) => put<{ ok: boolean }>("/system/time/ntp", { enabled }),

  processes: (limit = 50) => get<Process[]>(`/system/processes?limit=${limit}`),

  power: (action: "reboot" | "poweroff") => post<{ ok: boolean }>("/system/power", { action }),
};
