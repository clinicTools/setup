// Gemultiplexter WebSocket-Client nach Cockpit-Vorbild: eine Verbindung trägt
// beliebig viele Channel-Subscriptions. Subscriptions werden bei
// Verbindungsabbruch automatisch wiederhergestellt; der Server-Collector läuft
// nur, solange die Subscription besteht.
import { ref } from "vue";

export type WsStatus = "connecting" | "open" | "closed";

interface Subscription {
  id: string;
  channel: string;
  params?: unknown;
  onMessage: (payload: unknown) => void;
  onError?: (error: string) => void;
  ready: boolean;
}

interface ServerMessage {
  type: "ready" | "message" | "error" | "closed" | "pong";
  id?: string;
  channel?: string;
  payload?: unknown;
  error?: string;
}

class WsClient {
  readonly status = ref<WsStatus>("closed");

  private socket: WebSocket | null = null;
  private subs = new Map<string, Subscription>();
  private counter = 0;
  private reconnectDelay = 1000;
  private reconnectTimer: number | null = null;
  private manualClose = false;

  private url(): string {
    const proto = location.protocol === "https:" ? "wss" : "ws";
    return `${proto}://${location.host}/api/ws`;
  }

  /** Eröffnet einen Channel und liefert eine Unsubscribe-Funktion zurück. */
  subscribe(
    channel: string,
    params: unknown,
    onMessage: (payload: unknown) => void,
    onError?: (error: string) => void,
  ): () => void {
    const id = `s${++this.counter}`;
    const sub: Subscription = { id, channel, params, onMessage, onError, ready: false };
    this.subs.set(id, sub);

    this.ensureConnection();
    if (this.socket?.readyState === WebSocket.OPEN) this.sendSubscribe(sub);

    return () => this.unsubscribe(id);
  }

  private unsubscribe(id: string): void {
    if (!this.subs.delete(id)) return;
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify({ type: "unsubscribe", id }));
    }
    // Ohne aktive Subscriptions die Verbindung schließen (spart Ressourcen,
    // entspricht dem on-demand-Gedanken).
    if (this.subs.size === 0) this.close();
  }

  private ensureConnection(): void {
    if (this.socket && this.socket.readyState <= WebSocket.OPEN) return;
    this.manualClose = false;
    this.status.value = "connecting";
    const socket = new WebSocket(this.url());
    this.socket = socket;

    socket.onopen = () => {
      this.status.value = "open";
      this.reconnectDelay = 1000;
      // Alle bestehenden Subscriptions (neu) anmelden.
      for (const sub of this.subs.values()) this.sendSubscribe(sub);
    };

    socket.onmessage = (ev) => this.handleMessage(ev.data as string);

    socket.onclose = () => {
      this.status.value = "closed";
      this.socket = null;
      if (!this.manualClose && this.subs.size > 0) this.scheduleReconnect();
    };

    socket.onerror = () => socket.close();
  }

  private sendSubscribe(sub: Subscription): void {
    sub.ready = false;
    this.socket?.send(
      JSON.stringify({ type: "subscribe", id: sub.id, channel: sub.channel, params: sub.params }),
    );
  }

  private handleMessage(data: string): void {
    let msg: ServerMessage;
    try {
      msg = JSON.parse(data);
    } catch {
      return;
    }
    if (!msg.id) return;
    const sub = this.subs.get(msg.id);
    if (!sub) return;

    switch (msg.type) {
      case "ready":
        sub.ready = true;
        break;
      case "message":
        sub.onMessage(msg.payload);
        break;
      case "error":
        sub.onError?.(msg.error || "Channel-Fehler");
        break;
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer !== null) return;
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null;
      if (this.subs.size > 0) this.ensureConnection();
    }, this.reconnectDelay);
    // Exponentielles Backoff bis max. 15 s.
    this.reconnectDelay = Math.min(this.reconnectDelay * 2, 15000);
  }

  private close(): void {
    this.manualClose = true;
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    this.socket?.close();
    this.socket = null;
    this.status.value = "closed";
  }
}

export const ws = new WsClient();
