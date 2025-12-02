<script setup>
import {ref, computed, onMounted, onUnmounted} from 'vue'
import {useI18n} from 'vue-i18n'
import {EventsOn} from '../wailsjs/runtime/runtime.js'

const { t } = useI18n()
const logs = ref([])
const activeTab = ref('running') // 'running' (运行日志) 或 'operation' (操作日志)
const logContainer = ref(null)

const filteredLogs = computed(() => {
  return logs.value.filter(log => log.type === activeTab.value)
})

const addLog = (msg) => {
  const time = new Date().toLocaleTimeString()
  logs.value.push({
    time,
    content: msg.message || msg,
    type: msg.type || 'running',
    level: msg.level || 'INFO',
    id: Date.now() + Math.random()
  })
  
  // 保留最后 1000 条日志
  if (logs.value.length > 1000) {
    logs.value.shift()
  }
  
  // 自动滚动
  setTimeout(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  }, 10)
}

const copyLogs = () => {
  const text = filteredLogs.value.map(l => `[${l.time}] [${l.level}] ${l.content}`).join('\n')
  navigator.clipboard.writeText(text)
}

onMounted(() => {
  EventsOn("log", (msg) => {
    addLog(msg)
  })
  // addLog(t('logs.init'))
})
</script>

<template>
  <div class="log-panel">
    <div class="card log-card">
      <div class="card-header">
        <div class="header-left">
          <h2>{{ $t('logs.title') }}</h2>
          <div class="tabs">
            <button 
              :class="['tab-btn', { active: activeTab === 'running' }]"
              @click="activeTab = 'running'"
            >
              {{ $t('logs.running_logs') }}
            </button>
            <button 
              :class="['tab-btn', { active: activeTab === 'operation' }]"
              @click="activeTab = 'operation'"
            >
              {{ $t('logs.operation_logs') }}
            </button>
          </div>
          <span class="log-count">{{ filteredLogs.length }} {{ $t('logs.lines') }}</span>
        </div>
        <div class="actions">
          <button @click="copyLogs" class="action-btn" :title="$t('logs.copy_tooltip')">
            📋 {{ $t('logs.copy') }}
          </button>
          <button @click="logs = []" class="action-btn" :title="$t('logs.clear_tooltip')">
            🗑️ {{ $t('logs.clear') }}
          </button>
        </div>
      </div>
      
      <div class="log-container" ref="logContainer">
        <div v-if="filteredLogs.length === 0" class="empty-logs">
          {{ $t('logs.empty') }}
        </div>
        <div v-else v-for="log in filteredLogs" :key="log.id" class="log-line">
          <span class="log-time">[{{ log.time }}]</span>
          <span class="log-content">{{ log.content }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.log-panel {
  height: 100%;
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
}

.card {
  background: white;
  border-radius: 12px;
  box-shadow: var(--card-shadow);
  border: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.card-header {
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.tabs {
  display: flex;
  gap: 10px;
  background: #f0f2f5;
  padding: 4px;
  border-radius: 8px;
}

.tab-btn {
  border: none;
  background: transparent;
  padding: 6px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-secondary);
  transition: all 0.2s;
}

.tab-btn.active {
  background: white;
  color: var(--primary-color);
  box-shadow: 0 2px 4px rgba(0,0,0,0.05);
  font-weight: 500;
}

.tab-btn:hover:not(.active) {
  color: var(--text-primary);
}

h2 {
  margin: 0;
  font-size: 18px;
  color: var(--text-primary);
  border: none;
  padding: 0;
}

.log-count {
  font-size: 12px;
  color: var(--text-secondary);
  background: #f0f2f5;
  padding: 2px 8px;
  border-radius: 12px;
}

.actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  padding: 6px 12px;
  border: 1px solid var(--border-color);
  background: white;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-secondary);
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.action-btn:hover {
  background: #f5f7fa;
  color: var(--text-primary);
  border-color: #dcdfe6;
}

.log-container {
  flex: 1;
  background: #1e1e1e;
  padding: 16px;
  overflow-y: auto;
  font-family: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.5;
}

.empty-logs {
  color: #666;
  text-align: center;
  padding-top: 40px;
  font-style: italic;
}

.log-line {
  display: flex;
  gap: 12px;
  padding: 2px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.log-line:hover {
  background: rgba(255, 255, 255, 0.05);
}

.log-time {
  color: #858585;
  white-space: nowrap;
  user-select: none;
}

.log-content {
  color: #d4d4d4;
  white-space: pre-wrap;
  word-break: break-all;
}

/* Custom scrollbar for dark theme */
.log-container::-webkit-scrollbar {
  width: 10px;
}

.log-container::-webkit-scrollbar-track {
  background: #1e1e1e;
}

.log-container::-webkit-scrollbar-thumb {
  background: #424242;
  border-radius: 5px;
  border: 2px solid #1e1e1e;
}

.log-container::-webkit-scrollbar-thumb:hover {
  background: #4f4f4f;
}
</style>
