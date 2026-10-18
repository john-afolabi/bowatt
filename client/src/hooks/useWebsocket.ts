import { useEffect, useRef, useState } from "react";

import type { ConnectionStatus } from "@/types";

const BASE_DELAY_MS = 500;
const MAX_DELAY_MS = 30_000;

export function reconnectDelay(attempt: number, random: () => number = Math.random): number {
  const exponential = Math.min(MAX_DELAY_MS, BASE_DELAY_MS * 2 ** Math.max(0, attempt - 1));
  return Math.round(exponential * (0.5 + random() * 0.5));
}

type Options = {
  url: string;
  onMessage: (payload: unknown) => void;
  onReconnect?: () => void;
};

export function useWebSocket({ url, onMessage, onReconnect }: Options): ConnectionStatus {
  const [status, setStatus] = useState<ConnectionStatus>("connecting");

  const messageRef = useRef(onMessage);
  const reconnectRef = useRef(onReconnect);

  useEffect(() => {
    messageRef.current = onMessage;
    reconnectRef.current = onReconnect;
  }, [onMessage, onReconnect]);

  useEffect(() => {
    let socket: WebSocket | null = null;
    let timer: ReturnType<typeof setTimeout> | undefined;
    let attempt = 0;
    let disposed = false;
    let wasLive = false;

    const connect = () => {
      if (disposed) return;
      setStatus((current) => (current === "live" ? current : "connecting"));

      const next = new WebSocket(url);
      socket = next;

      next.onopen = () => {
        attempt = 0;
        setStatus("live");
        if (wasLive) reconnectRef.current?.();
        wasLive = true;
      };

      next.onmessage = (event: MessageEvent<string>) => {
        messageRef.current(event.data);
      };

      next.onerror = () => {
        next.close();
      };

      next.onclose = () => {
        if (disposed) return;
        setStatus("offline");
        attempt += 1;
        timer = setTimeout(connect, reconnectDelay(attempt));
      };
    };

    connect();

    return () => {
      disposed = true;
      if (timer) clearTimeout(timer);
      socket?.close();
    };
  }, [url]);

  return status;
}