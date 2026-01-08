import { defineStore } from "pinia"

type UpdateStateResponse = {
  isAlive: boolean
  shouldStop: boolean
}

export const useServerStore = defineStore("server", {
  state: () => ({
    isAlive: false,
    shouldStop: false,
    lastUpdated: null as number | null,

    // optional UX bits
    loading: false,
    error: null as string | null,
  }),

  getters: {
    status(state): "offline" | "running" | "stopping" | "unknown" {
      if (state.lastUpdated === null) return "unknown"
      if (!state.isAlive) return "offline"
      if (state.shouldStop) return "stopping"
      return "running"
    },
  },

  actions: {
    applyServerState(payload: UpdateStateResponse) {
      this.isAlive = payload.isAlive
      this.shouldStop = payload.shouldStop
      this.lastUpdated = Date.now()
    },

    async fetchState() {
      this.loading = true
      this.error = null
      try {
        const res = await fetch("/udp/server-state", {
          method: "GET",
          headers: { "Accept": "application/json" },
          cache: "no-store",
        })

        if (!res.ok) throw new Error(`HTTP ${res.status}`)

        const data = (await res.json()) as UpdateStateResponse
        this.applyServerState(data)
      } catch (e: any) {
        this.error = e?.message ?? "Failed to fetch server state"
      } finally {
        this.loading = false
      }
    },

    async startServer() {
      this.loading = true
      this.error = null
      try {
        const res = await fetch("/udp/server-start", {
          method: "POST",
          headers: { "Accept": "application/json" },
        })

        if (!res.ok) throw new Error(`HTTP ${res.status}`)

        // Aktualisiere den Status nach dem Start
        await this.fetchState()
      } catch (e: any) {
        this.error = e?.message ?? "Failed to start server"
      } finally {
        this.loading = false
      }
    },

    async stopServer() {
      this.loading = true
      this.error = null
      try {
        const res = await fetch("/udp/server-stop", {
          method: "POST",
          headers: { "Accept": "application/json" },
        })

        if (!res.ok) throw new Error(`HTTP ${res.status}`)

        // Aktualisiere den Status nach dem Stoppen
        await this.fetchState()
      } catch (e: any) {
        this.error = e?.message ?? "Failed to stop server"
      } finally {
        this.loading = false
      }
    },
  },
})
