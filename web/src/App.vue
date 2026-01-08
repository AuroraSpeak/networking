<script setup lang="ts">
import { computed } from "vue";
import { useStringWs } from "@/composables/useStringWs";
import WsLog from "@/components/WsLog.vue";
import WsSendInput from "@/components/WsSendInput.vue";
import ServerState from "./components/ServerState.vue";
import { useServerStore } from "@/stores/UDPServer";
import { useServerButton } from "./composables/useServerButton";

const udpServerButtonConfig = useServerButton();
const serverStore = useServerStore();
serverStore.fetchState();

const buttonConfig = udpServerButtonConfig.buttonConfig;

async function handleServerAction() {
  const action = buttonConfig.value.action;
  if (action === "connect") {
    await serverStore.startServer();
  } else if (action === "disconnect") {
    await serverStore.stopServer();
  }
}

const wsUrl = import.meta.env.DEV
  ? "ws://localhost:8080/ws"
  : (location.protocol === "https:" ? "wss://" : "ws://") + location.host + "/ws";

const { status, lines, error, send, connect, close } = useStringWs(wsUrl);

const canSend = computed(() => status.value === "open");

function handleSend(text: string) {
  const ok = send(text);
  if (!ok) lines.value.push("[ws] not connected");
}
</script>

<template>
  <div class="drawer">
    <input id="drawer-sidebar" type="checkbox" class="drawer-toggle">
    <div class="drawer-content">
      <label for="drawer-sidebar" class="btn drawer-button">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"
          stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </label>
      <div class="p-6 space-y-4">
        <div class="flex items-center gap-3">
          <span class="badge"
            :class="status === 'open' ? 'badge-success' : status === 'connecting' ? 'badge-warning' : 'badge-ghost'">
            {{ status }}
          </span>
          <span v-if="error" class="text-error text-sm">{{ error }}</span>

          <div class="ml-auto flex gap-2">
            <button class="btn btn-sm" @click="connect">Connect</button>
            <button class="btn btn-sm btn-outline" @click="close">Close</button>
          </div>
        </div>

        <WsLog :lines="lines" />

        <WsSendInput :disabled="!canSend" @send="handleSend" />
      </div>
    </div>
    <div class="drawer-side">
      <label for="drawer-sidebar" class="drawer-overlay"></label>
      <ul class="menu bg-base-200 min-h-full w-80 p-4">
        <li><button onclick="server_modal.showModal()">Server</button></li>
        <li><a href="#">Clients</a></li>
      </ul>
    </div>
    <!-- Server Modal -->
    <dialog id="server_modal" class="modal">
      <div class="modal-box">
        <ServerState />
        <div class="modal-action">
          <form method="dialog">
          <button :class="buttonConfig.class" :disabled="buttonConfig.disabled" @click="handleServerAction">
            {{ buttonConfig.label }}
          </button>
          <button class="btn">Close</button>
          </form>
        </div>
      </div>
    </dialog>
  </div>
</template>
