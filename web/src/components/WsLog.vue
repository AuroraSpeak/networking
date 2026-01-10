<script setup lang="ts">
import { computed, ref, toRef } from "vue";
import { useLogParser } from "@/composables/useLogParser";
import LogDetailModal from "./LogDetailModal.vue";
import type { LogEntry } from "@/api/types";

const props = defineProps<{
  lines: string[];
}>();

const { parsedLines } = useLogParser(toRef(props, "lines"));

const selectedLog = ref<LogEntry | null>(null);

function getLevelBadgeClass(level: string): string {
  const levelLower = level.toLowerCase();
  if (levelLower === "error" || levelLower === "fatal") {
    return "badge-error";
  }
  if (levelLower === "warn" || levelLower === "warning") {
    return "badge-warning";
  }
  if (levelLower === "info") {
    return "badge-info";
  }
  if (levelLower === "debug" || levelLower === "trace") {
    return "badge-ghost";
  }
  return "badge";
}

function openLogDetail(entry: LogEntry) {
  selectedLog.value = entry;
}

function closeLogDetail() {
  selectedLog.value = null;
}
</script>

<template>
  <div class="bg-base-200 rounded-lg h-72 overflow-auto p-3">
    <div class="space-y-1">
      <div
        v-for="line in parsedLines"
        :key="line.index"
        class="flex items-start gap-2"
      >
        <div
          v-if="line.type === 'log' && line.log"
          class="flex-1 cursor-pointer hover:bg-base-300 rounded px-2 py-1 transition-colors border border-base-300"
          @click="openLogDetail(line.log)"
        >
          <div class="flex items-center gap-2">
            <span
              class="badge badge-sm"
              :class="getLevelBadgeClass(line.log.level)"
            >
              {{ line.log.level }}
            </span>
            <span class="text-sm">{{ line.log.msg }}</span>
          </div>
        </div>
        <div
          v-else-if="line.type === 'text' && line.text"
          class="flex-1 text-sm whitespace-pre-wrap border border-base-300 rounded px-2 py-1"
        >
          {{ line.text }}
        </div>
      </div>
    </div>
    
    <LogDetailModal :entry="selectedLog" @close="closeLogDetail" />
  </div>
</template>
