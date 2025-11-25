<script setup>
import {ref, onMounted, onUnmounted} from 'vue'
import {EventsOn} from '../../wailsjs/runtime/runtime'

const logs = ref([])
const logContainer = ref(null)

const addLog = (msg) => {
  const time = new Date().toLocaleTimeString()
  logs.value.push(`[${time}] ${msg}`)
  // Keep last 1000 logs
  if (logs.value.length > 1000) {
    logs.value.shift()
  }
  // Auto scroll
  setTimeout(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  }, 10)
}

onMounted(() => {
  EventsOn("log", (msg) => {
    addLog(msg)
  })
  addLog("Log viewer initialized...")
})
</script>

<template>
  <div class="log-panel">
    <div class="header">
      <h2>{{ $t('logs.title') }}</h2>
      <button @click="logs = []" class="clear-btn">{{ $t('logs.clear') }}</button>
    </div>
    <div class="log-container" ref="logContainer">
      <div v-for="(log, index) in logs" :key="index" class="log-line">
        {{ log }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.log-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

h2 {
  margin: 0;
  color: #2c3e50;
  font-size: 16px;
  font-weight: 600;
  border-left: 4px solid var(--primary-color);
  padding-left: 10px;
}

.clear-btn {
  padding: 4px 12px;
  background: #f5f5f5;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  color: #666;
  transition: all 0.2s;
}

.clear-btn:hover {
  background: #e0e0e0;
  color: #333;
}

.log-container {
  flex: 1;
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 15px;
  border-radius: 8px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  overflow-y: auto;
  white-space: pre-wrap;
  box-shadow: inset 0 0 10px rgba(0,0,0,0.2);
}

.log-line {
  margin-bottom: 4px;
  line-height: 1.4;
  border-bottom: 1px solid #333;
  padding-bottom: 2px;
}

.log-line:last-child {
  border-bottom: none;
}
</style>
