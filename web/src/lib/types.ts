// Typdefinitionen spiegeln die JSON-Strukturen der Go-API wider.

export interface User {
  username: string;
  uid: number;
  gid: number;
  fullName: string;
  homeDir: string;
  shell: string;
  groups: string[] | null;
  admin: boolean;
}

export interface MemoryInfo {
  total: number;
  used: number;
  free: number;
}

export interface HostInfo {
  hostname: string;
  os: string;
  prettyName: string;
  kernel: string;
  architecture: string;
  virtualization?: string;
  uptimeSeconds: number;
  bootTime: string;
  cpuModel: string;
  cpuCores: number;
  loadAvg: [number, number, number];
  memory: MemoryInfo;
  swap: MemoryInfo;
}

export interface SystemUser {
  username: string;
  uid: number;
  gid: number;
  fullName: string;
  homeDir: string;
  shell: string;
  system: boolean;
  groups: string[] | null;
}

export interface SystemGroup {
  name: string;
  gid: number;
  system: boolean;
  members: string[] | null;
}

export interface Service {
  name: string;
  description: string;
  loadState: string;
  activeState: string;
  subState: string;
  unitFileState?: string;
}

export interface Package {
  name: string;
  version: string;
  availableVersion?: string;
  architecture?: string;
}

export interface PackageSummary {
  installed: number;
  upgradable: number;
  upgrades: Package[] | null;
  lastUpdated?: string;
}

export interface NetworkInterface {
  name: string;
  mac: string;
  mtu: number;
  up: boolean;
  loopback: boolean;
  addresses: string[] | null;
}

export interface NetworkInfo {
  hostname: string;
  interfaces: NetworkInterface[];
  dns: string[] | null;
  gateway?: string;
}

export interface Filesystem {
  device: string;
  mountpoint: string;
  type: string;
  total: number;
  used: number;
  free: number;
  usePercent: number;
}

export interface BlockDevice {
  name: string;
  size: string;
  type: string;
  mountpoint?: string;
  fstype?: string;
  model?: string;
  children?: BlockDevice[];
}

export interface StorageInfo {
  filesystems: Filesystem[] | null;
  devices: BlockDevice[] | null;
}

export interface FirewallRule {
  to: string;
  action: string;
  from: string;
}

export interface FirewallStatus {
  available: boolean;
  enabled: boolean;
  rules: FirewallRule[] | null;
}

export interface LogEntry {
  timestamp: string;
  hostname: string;
  unit: string;
  priority: number;
  message: string;
}

export interface Timer {
  unit: string;
  next: string;
  left: string;
  last: string;
  passed: string;
  activates: string;
}

export interface CronJob {
  source: string;
  schedule: string;
  user?: string;
  command: string;
}

export interface ScheduledTasks {
  timers: Timer[] | null;
  cron: CronJob[] | null;
}

export interface TimeInfo {
  localTime: string;
  universalTime: string;
  timezone: string;
  ntpEnabled: boolean;
  ntpSynced: boolean;
  rtcInLocalTime: boolean;
}

export interface NetRate {
  name: string;
  rxRate: number;
  txRate: number;
}

export interface MetricsSample {
  timestamp: number;
  cpuPercent: number;
  memory: MemoryInfo;
  swap: MemoryInfo;
  loadAvg: [number, number, number];
  interfaces: NetRate[];
}

export interface Job {
  id: string;
  name: string;
  status: "running" | "succeeded" | "failed" | "canceled";
  startedAt: string;
  finishedAt?: string;
  exitCode: number;
  error?: string;
}

export interface JobLine {
  seq: number;
  time: string;
  stream: "stdout" | "stderr" | "system";
  text: string;
}

export interface AuditEntry {
  time: string;
  user: string;
  action: string;
  target?: string;
  success: boolean;
  detail?: string;
  ip?: string;
}

export interface PodmanStatus {
  installed: boolean;
  version?: string;
  composeAvailable: boolean;
  composeCommand?: string;
  socketActive: boolean;
  containers: number;
  running: number;
  images: number;
}

export interface Container {
  id: string;
  name: string;
  image: string;
  state: string;
  status: string;
  created?: string;
  ports: string[] | null;
  stack?: string;
  service?: string;
}

export interface ContainerImage {
  id: string;
  names: string[] | null;
  size: number;
  created?: string;
}

export interface ContainerVolume {
  name: string;
  driver: string;
  mountpoint: string;
  createdAt?: string;
}

export interface Stack {
  name: string;
  path: string;
  containers: number;
  running: number;
  status: "running" | "partial" | "stopped" | "unknown";
  modifiedAt?: string;
}

export interface Process {
  pid: number;
  ppid: number;
  user: string;
  state: string;
  command: string;
  rssMB: number;
  threads: number;
}
