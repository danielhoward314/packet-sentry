import { useEnv } from "@/contexts/EnvContext";
import { LOCALSTORAGE } from "@/lib/consts";
import { useEffect, useRef } from "react";

export function useWebSocket() {
  const { API_BASE_URL } = useEnv();
  const ws = useRef<WebSocket | null>(null);
  const { API_ACCESS_TOKEN } = LOCALSTORAGE;

  useEffect(() => {
    const token = localStorage.getItem(API_ACCESS_TOKEN);
    if (!token) {
      console.warn("No access token found in localStorage");
      return;
    }

    const baseWsUrl = toWebSocketUrl(API_BASE_URL);
    const wsUrl = `${baseWsUrl}/ws?token=${encodeURIComponent(token)}`;

    ws.current = new WebSocket(wsUrl);

    ws.current.onopen = () => {
      console.log("WebSocket connected");
    };

    ws.current.onmessage = (event) => {
      console.log("Message from server:", event.data);
    };

    ws.current.onclose = () => {
      console.log("WebSocket disconnected");
    };

    ws.current.onerror = (err) => {
      console.error("WebSocket error:", err);
    };

    return () => {
      ws.current?.close();
    };
  }, []);
}

function toWebSocketUrl(httpUrl: string): string {
  if (httpUrl.startsWith("https://")) {
    return httpUrl.replace("https://", "wss://");
  } else if (httpUrl.startsWith("http://")) {
    return httpUrl.replace("http://", "ws://");
  }
  throw new Error(`Invalid HTTP URL: ${httpUrl}`);
}
