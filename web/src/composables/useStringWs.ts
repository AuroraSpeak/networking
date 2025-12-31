import { onBeforeUnmount, ref } from "vue";

export function useStringWs(url: string) {
  const status = ref<"connecting" | "open" | "closed" | "error">("connecting");
  const lines = ref<string[]>([]);
  const error = ref<string | null>(null);

  let ws: WebSocket | null = null;

  function connect() {
    status.value = "connecting";
    error.value = null;

    ws = new WebSocket(url);

    ws.onopen = () => {
      status.value = "open";
      lines.value.push("[ws] connected");
    };

    ws.onmessage = (ev) => {
      lines.value.push(String(ev.data));
    };

    ws.onerror = () => {
      status.value = "error";
      error.value = "WebSocket error";
      lines.value.push("[ws] error");
    };

    ws.onclose = () => {
      status.value = "closed";
      lines.value.push("[ws] closed");
    };
  }

  function send(text: string) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return false;
    ws.send(text); // <- plain string
    return true;
  }

  function close() {
    ws?.close();
    ws = null;
  }

  connect();
  onBeforeUnmount(close);

  return { status, lines, error, send, connect, close };
}
