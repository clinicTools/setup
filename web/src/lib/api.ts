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
  Job,
  AuditEntry,
  PodmanStatus,
  Container,
  ContainerImage,
  ContainerVolume,
  Stack,
} from "./types";

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
    this.name = "ApiError";
  }
}

/** Liest ein Cookie (für den CSRF-Double-Submit-Token). */
function readCookie(name: string): string {
  const match = document.cookie.match(new RegExp("(?:^|; )" + name + "=([^;]*)"));
  return match ? decodeURIComponent(match[1]) : "";
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {};
  if (body) headers["Content-Type"] = "application/json";
  // CSRF-Token bei zustandsändernden Methoden mitschicken.
  if (!["GET", "HEAD", "OPTIONS"].includes(method)) {
    headers["X-CSRF-Token"] = readCookie("da_csrf");
  }

  const res = await fetch(`/api${path}`, {
    method,
    credentials: "include",
    headers,
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
const del = <T>(p: string, b?: unknown) => request<T>("DELETE", p, b);

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

  modifyUser: (name: string, groups: string[] | null, shell: string) =>
    put<{ ok: boolean }>(`/system/users/${encodeURIComponent(name)}`, { groups, shell }),
  createGroup: (name: string, system: boolean) =>
    post<{ ok: boolean }>("/system/groups", { name, system }),
  deleteGroup: (name: string) =>
    del<{ ok: boolean }>(`/system/groups/${encodeURIComponent(name)}`),

  packages: () => get<PackageSummary>("/system/packages"),
  installedPackages: () => get<Package[]>("/system/packages/installed"),
  aptUpdate: () => post<{ output: string }>("/system/packages/update"),
  packageInstall: (packages: string[]) =>
    post<{ jobId: string }>("/system/packages/install", { packages }),
  packageRemove: (packages: string[]) =>
    post<{ jobId: string }>("/system/packages/remove", { packages }),
  packageUpgrade: () => post<{ jobId: string }>("/system/packages/upgrade"),

  network: () => get<NetworkInfo>("/system/network"),
  storage: () => get<StorageInfo>("/system/storage"),

  firewall: () => get<FirewallStatus>("/system/firewall"),
  setFirewall: (enabled: boolean) => put<{ ok: boolean }>("/system/firewall", { enabled }),
  firewallAddRule: (action: string, port: string, protocol: string) =>
    post<{ ok: boolean }>("/system/firewall/rules", { action, port, protocol }),
  firewallDeleteRule: (action: string, port: string, protocol: string) =>
    del<{ ok: boolean }>("/system/firewall/rules", { action, port, protocol }),

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
  killProcess: (pid: number, signal: string) =>
    post<{ ok: boolean }>(`/system/processes/${pid}/kill`, { signal }),

  createCron: (name: string, schedule: string, user: string, command: string) =>
    post<{ ok: boolean }>("/system/cron", { name, schedule, user, command }),
  deleteCron: (name: string) => del<{ ok: boolean }>(`/system/cron/${encodeURIComponent(name)}`),

  setHostname: (hostname: string) =>
    put<{ ok: boolean }>("/system/network/hostname", { hostname }),
  setInterfaceState: (iface: string, up: boolean) =>
    put<{ ok: boolean }>(`/system/network/interfaces/${encodeURIComponent(iface)}`, { up }),

  // Jobs
  jobs: () => get<Job[]>("/system/jobs"),
  job: (id: string) => get<Job>(`/system/jobs/${id}`),
  cancelJob: (id: string) => post<{ ok: boolean }>(`/system/jobs/${id}/cancel`),

  // Audit
  audit: (limit = 200) => get<AuditEntry[]>(`/system/audit?limit=${limit}`),

  // Podman
  podmanStatus: () => get<PodmanStatus>("/system/podman/status"),
  podmanInstall: () => post<{ jobId: string }>("/system/podman/install"),
  containers: () => get<Container[]>("/system/podman/containers"),
  containerLogs: (id: string, lines = 200) =>
    get<{ logs: string }>(`/system/podman/containers/${encodeURIComponent(id)}/logs?lines=${lines}`),
  containerAction: (id: string, action: string) =>
    post<{ ok: boolean }>(`/system/podman/containers/${encodeURIComponent(id)}/action`, { action }),
  containerImages: () => get<ContainerImage[]>("/system/podman/images"),
  containerVolumes: () => get<ContainerVolume[]>("/system/podman/volumes"),

  // Compose-Stacks
  stacks: () => get<Stack[]>("/system/podman/stacks"),
  stack: (name: string) =>
    get<{ name: string; compose: string }>(`/system/podman/stacks/${encodeURIComponent(name)}`),
  saveStack: (name: string, compose: string) =>
    put<{ ok: boolean }>(`/system/podman/stacks/${encodeURIComponent(name)}`, { compose }),
  deleteStack: (name: string) =>
    del<{ ok: boolean }>(`/system/podman/stacks/${encodeURIComponent(name)}`),
  validateStack: (name: string) =>
    post<{ valid: boolean }>(`/system/podman/stacks/${encodeURIComponent(name)}/validate`),
  stackAction: (name: string, action: string) =>
    post<{ jobId: string }>(
      `/system/podman/stacks/${encodeURIComponent(name)}/${encodeURIComponent(action)}`,
    ),

  power: (action: "reboot" | "poweroff") => post<{ ok: boolean }>("/system/power", { action }),
};
