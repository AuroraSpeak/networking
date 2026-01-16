<template>
    <div class="flex flex-col h-full space-y-4">
        <div class="flex items-center justify-between">
            <h2 class="text-2xl font-bold">Packet Sniffer</h2>
            <div class="flex gap-2">
                <button class="btn btn-sm btn-primary" @click="refreshPackets" :disabled="loading">
                    <span v-if="loading" class="loading loading-spinner loading-xs"></span>
                    {{ loading ? 'Laden...' : 'Aktualisieren' }}
                </button>
                <button class="btn btn-sm btn-error" @click="clearPackets" :disabled="loading">
                    Löschen
                </button>
            </div>
        </div>

        <div class="flex-1 overflow-auto">
            <div v-if="packets.length === 0" class="alert alert-info">
                <span>Keine Pakete erfasst</span>
            </div>
            <div v-else class="overflow-x-auto">
                <table class="table table-zebra w-full">
                    <thead>
                        <tr>
                            <th>Timestamp</th>
                            <th>Richtung</th>
                            <th>Local</th>
                            <th>Remote</th>
                            <th>Client ID</th>
                            <th>Pakettyp</th>
                            <th>Hex</th>
                            <th>String</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="(packet, index) in packets" :key="index">
                            <td class="text-xs">{{ formatTimestamp(packet.ts) }}</td>
                            <td>
                                <div class="badge" :class="packet.dir === 'in' ? 'badge-info' : 'badge-warning'">
                                    {{ packet.dir === 'in' ? 'IN' : 'OUT' }}
                                </div>
                            </td>
                            <td class="text-xs">{{ packet.local }}</td>
                            <td class="text-xs">{{ packet.remote }}</td>
                            <td>
                                <span v-if="packet.client_id === 0" class="text-xs text-base-content/60">Unbekannt</span>
                                <span v-else class="text-xs">{{ packet.client_id }}</span>
                            </td>
                            <td class="text-xs">
                                <span v-if="packet.packet_type" class="badge badge-ghost badge-sm">{{ packet.packet_type }}</span>
                                <span v-else class="text-base-content/40">-</span>
                            </td>
                            <td>
                                <code class="text-xs break-all">
                                    {{ formatDatagram(packet.payload) }}
                                </code>
                            </td>
                            <td>
                                <code class="text-xs break-all">
                                    {{ formatDatagramText(packet.payload) }}
                                </code>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import type { SnifferPacket } from "@/api/types";
import { useApi } from "@/api/useApi";
import { useDatagramFormatting } from "@/composables/useDatagramFormatting";

const api = useApi();
const { formatDatagram, formatDatagramText } = useDatagramFormatting();

const packets = ref<SnifferPacket[]>([]);
const loading = ref(false);

function formatTimestamp(ts: string): string {
    try {
        const date = new Date(ts);
        const timeStr = date.toLocaleTimeString('de-DE', { 
            hour: '2-digit', 
            minute: '2-digit', 
            second: '2-digit'
        });
        // Add milliseconds manually
        const ms = date.getMilliseconds().toString().padStart(3, '0');
        return `${timeStr}.${ms}`;
    } catch {
        return ts;
    }
}

async function refreshPackets() {
    loading.value = true;
    try {
        packets.value = await api.sniffer.getPackets();
    } catch (error) {
        console.error("Failed to fetch packets:", error);
    } finally {
        loading.value = false;
    }
}

async function clearPackets() {
    if (!confirm("Möchten Sie wirklich alle Pakete löschen?")) {
        return;
    }
    loading.value = true;
    try {
        await api.sniffer.clearPackets();
        packets.value = [];
    } catch (error) {
        console.error("Failed to clear packets:", error);
    } finally {
        loading.value = false;
    }
}

onMounted(() => {
    refreshPackets();
});
</script>
