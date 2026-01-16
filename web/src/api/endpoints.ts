import type { ApiClient } from "./client";
import type { ID, Paginated, ServerState, UDPClient, UDPClientState, SendDatagramRequest, MermaidTraces, SnifferPacket } from "./types";

export interface ServerApi {
    start: () => Promise<void>;
    stop: () => Promise<void>;
    getState: () => Promise<ServerState>;
}

export interface UDPClientApi {
    start: () => Promise<void>;
    stop: () => Promise<void>;
    getStateByName: (name: string) => Promise<UDPClientState>;
    getStateById: (id: ID) => Promise<UDPClientState>;
    getAll: () => Promise<{ udpClients: UDPClient[] }>;
    list: (params: { page?: number, pageSize?: number, q?: string }) => Promise<Paginated<UDPClient>>;
    sendDatagram: (request: SendDatagramRequest) => Promise<void>;
}

export interface TraceApi {
    getAll: (name: string) => Promise<MermaidTraces>;
}

export interface SnifferApi {
    getPackets: (name?: string) => Promise<SnifferPacket[]>;
    clearPackets: () => Promise<void>;
}

export function createServerApi(client: ApiClient): ServerApi {
    return {
        start: () => client.post("/api/server/start"),
        stop: () => client.post("/api/server/stop"),
        getState: () => client.get("/api/server/get"),
    }
}

export function createUDPClientApi(client: ApiClient): UDPClientApi {
    return {
        start: () => client.post("/api/client/start"),
        stop: () => client.post("/api/client/stop"),
        getStateByName: (name: string) => client.get("/api/client/get/name", { query: { name } }),
        getStateById: (id: ID) => client.get("/api/client/get/id", { query: { id } }),
        getAll: () => client.get("/api/client/get/all"),
        list: (params: { page?: number, pageSize?: number, q?: string }) => client.get("/api/client/get/all/paginated", { query: params }),
        sendDatagram: (request: SendDatagramRequest) => client.post("/api/client/send", { body: request }),
    };
}

export function createTraceApi(client: ApiClient): TraceApi {
    return {
        getAll: (name: string) => client.get("/api/traces/all", { query: {name}}),
    };
}

export function createSnifferApi(client: ApiClient): SnifferApi {
    return {
        getPackets: (name?: string) => {
            const query = name ? { name } : undefined;
            return client.get<{ packets: Array<{ ts: string; dir: string; local: string; remote: string; payload: string; client_id: number; packet_type?: string }> }>("/api/sniffer/packets", { query })
                .then(res => {
                    // Decode Base64 payload to Uint8Array
                    return res.packets.map(p => ({
                        ...p,
                        payload: Uint8Array.from(atob(p.payload), c => c.charCodeAt(0)),
                        dir: p.dir as "in" | "out"
                    }));
                });
        },
        clearPackets: () => client.post("/api/sniffer/clear"),
    };
}