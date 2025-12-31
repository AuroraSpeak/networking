<script setup lang="ts">
import { computed } from "vue";
import { useStringWs } from "@/composables/useStringWs";
import WsLog from "@/components/WsLog.vue";
import WsSendInput from "@/components/WsSendInput.vue";

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
</template>
