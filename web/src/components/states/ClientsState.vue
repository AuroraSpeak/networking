<template>
    <div class="h-full flex gap-4">
        <div class="w-1/5 border border-base-300 rounded-lg p-4 bg-base-200 flex flex-col h-full">
            <div class="flex flex-col gap-2 overflow-y-auto flex-1 min-h-0">
                <div v-for="client in clients" :key="client.id" class="flex items-center gap-2">
                    <button class="btn btn-primary w-full" @click="getClientState(client.id)">{{ client.name }}</button>
                </div>
            </div>
            <button class="btn btn-primary w-full mt-2" @click="api.udpClients.start()">Start</button>
        </div>
        <div class="flex-1 border border-base-300 rounded-lg p-4 bg-base-200 h-full overflow-y-auto">
            2
        </div>
    </div>
</template>

<script setup lang="ts">
import { useApi } from "@/api/useApi";
import type { UDPClientState, UDPClient } from "@/api/types";
import { onMounted, ref, watch } from "vue";
const props = defineProps<{
    needsUpdate: boolean;
    usuEvent: { id: number, seq: number } | null;
}>();
const emit = defineEmits<{
    (e: "done:clientStateUpdate"): void;
    (e: "done:clientsUpdate"): void;
}>();

const api = useApi();
const clients = ref<UDPClient[]>([]);
const clientState = ref<UDPClientState>();
const usuSeq = ref<number | null>(null);
onMounted(async () => {
    const clientData = await api.udpClients.getAll();
    clients.value = clientData.udpClients;
    console.log(clients.value);
});

watch(
    () => props.usuEvent,
    (e) => {
        if (!e) return;
        handleUsUEvent(e);
    },
    { flush: "post" }
)

async function handleUsUEvent(e: { id: number, seq: number }) {
    if (usuSeq.value === e.seq) return;
    if (e.id === clientState.value?.id) {
        const StateData = await api.udpClients.getStateById(e.id);
        clientState.value = StateData;
    }
}

watch(() =>props.needsUpdate, async () => {
    const clientData = await api.udpClients.getAll();
    clients.value = clientData.udpClients;
    console.log("clients updated", clients.value);
    emit("done:clientsUpdate");
});

async function getClientState(id: number) {
    const clientStateData = await api.udpClients.getStateById(id);
    console.log(clientStateData);
    clientState.value = clientStateData;
}

</script>