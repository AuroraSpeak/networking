<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  lines: string[];
}>();

const text = computed(() => props.lines.join("\n"));
</script>

<template>
  <pre class="bg-base-200 p-3 rounded-lg h-72 overflow-auto whitespace-pre-wrap">{{ text }}</pre>
</template>
