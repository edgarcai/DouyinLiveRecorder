<script setup>
import {reactive, onMounted, onUnmounted} from 'vue'
import {AddUrl, RemoveUrl, GetUrls, GetRecordingStatus} from '../wailsjs/go/main/App.js'

const state = reactive({
  url: '',
  urls: [],
  statusMap: {},
  message: ''
})

let timer = null

const updateStatus = () => {
  GetUrls().then(urls => {
    state.urls = urls || []
    GetRecordingStatus().then(result => {
      state.statusMap = result
    })
  })
}

const start = () => {
  if (!state.url) return
  AddUrl(state.url).then(result => {
    state.message = result
    state.url = ''
    updateStatus()
    // Auto clear message after 3 seconds
    setTimeout(() => {
      state.message = ''
    }, 3000)
  })
}

const stop = (url) => {
  RemoveUrl(url).then(result => {
    state.message = result
    updateStatus()
    setTimeout(() => {
      state.message = ''
    }, 3000)
  })
}

const getStatus = (url) => {
  const status = state.statusMap[url] || 'Stopped'
  if (status.includes('Recording')) return 'recording'
  if (status.includes('Monitoring')) return 'monitoring'
  if (status.includes('Stopped')) return 'stopped'
  if (status.includes('Error')) return 'error'
  return 'starting'
}

const getStatusText = (url) => {
  const key = getStatus(url)
  return key === 'error' ? state.statusMap[url] : key
}

const getStatusClass = (statusKey) => {
  if (statusKey === 'recording') return 'badge-recording'
  if (statusKey === 'monitoring') return 'badge-monitoring'
  if (statusKey === 'stopped' || statusKey === 'error') return 'badge-stopped'
  return 'badge-starting'
}

onMounted(() => {
  updateStatus()
  timer = setInterval(updateStatus, 2000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div class="status-panel">
    <h2>{{ $t('status.title') }}</h2>
    
    <div class="card control-card">
      <div class="input-group">
        <input 
          v-model="state.url" 
          :placeholder="$t('status.placeholder')" 
          type="text" 
          @keyup.enter="start"
        />
        <button @click="start" class="primary-btn">
          <span class="btn-icon">▶</span>
          {{ $t('status.start') }}
        </button>
        <button @click="$emit('toggle-mini')" class="icon-btn" :title="$t('status.mini_mode')">
          🪟
        </button>
      </div>
      
      <transition name="slide-fade">
        <div v-if="state.message" class="message-box" :class="{error: state.message.includes('Error')}">
          {{ state.message }}
        </div>
      </transition>
    </div>
    
    <div class="card list-card">
      <div class="card-header">
        <h3>{{ $t('status.active') }}</h3>
        <span class="count-badge">{{ state.urls.length }}</span>
      </div>
      
      <div v-if="state.urls.length === 0" class="empty-state">
        <div class="empty-icon">📭</div>
        <p>{{ $t('status.empty') }}</p>
      </div>
      
      <ul v-else class="url-list">
        <transition-group name="list">
          <li v-for="url in state.urls" :key="url" class="list-item">
            <div class="item-content">
              <div class="url-text" :title="url">{{ url }}</div>
              <div class="status-badge" :class="getStatusClass(getStatus(url))">
                <span class="status-dot"></span>
                {{ getStatus(url) === 'error' ? getStatusText(url) : $t('status.' + getStatus(url)) }}
              </div>
            </div>
            <button @click="stop(url)" class="stop-btn" :title="$t('status.stop')">
              ⏹
            </button>
          </li>
        </transition-group>
      </ul>
    </div>
  </div>
</template>

<style scoped>
.status-panel {
  max-width: 1000px;
  margin: 0 auto;
}

.card {
  background: white;
  border-radius: 12px;
  box-shadow: var(--card-shadow);
  margin-bottom: 24px;
  border: 1px solid var(--border-color);
  overflow: hidden;
}

.control-card {
  padding: 24px;
}

.input-group {
  display: flex;
  gap: 12px;
}

.input-group input {
  flex: 1;
  height: 44px;
}

.input-group button {
  height: 44px;
  display: flex;
  align-items: center;
  gap: 8px;
  gap: 8px;
  white-space: nowrap;
}

.icon-btn {
  height: 44px;
  width: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border-color);
  background: white;
  border-radius: 8px;
  cursor: pointer;
  font-size: 20px;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: #f5f7fa;
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.message-box {
  margin-top: 16px;
  padding: 12px 16px;
  background: #e3f2fd;
  color: #1565c0;
  border-radius: 8px;
  font-size: 14px;
  display: flex;
  align-items: center;
  border-left: 4px solid #1565c0;
}

.message-box.error {
  background: #ffebee;
  color: #c62828;
  border-left-color: #c62828;
}

.list-card {
  min-height: 300px;
}

.card-header {
  padding: 20px 24px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-header h3 {
  margin: 0;
  font-size: 16px;
  color: var(--text-primary);
  border: none;
  padding: 0;
}

.count-badge {
  background: #f0f2f5;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  color: var(--text-secondary);
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.url-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.list-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color);
  transition: background-color 0.2s;
}

.list-item:last-child {
  border-bottom: none;
}

.list-item:hover {
  background-color: #fafafa;
}

.item-content {
  flex: 1;
  min-width: 0;
  margin-right: 16px;
}

.url-text {
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 6px;
  font-size: 14px;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  gap: 6px;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: currentColor;
}

.badge-recording {
  background-color: rgba(46, 125, 50, 0.1);
  color: #2e7d32;
}

.badge-monitoring {
  background-color: rgba(21, 101, 192, 0.1);
  color: #1565c0;
}

.badge-stopped {
  background-color: rgba(198, 40, 40, 0.1);
  color: #c62828;
}

.badge-starting {
  background-color: rgba(239, 108, 0, 0.1);
  color: #ef6c00;
}

.stop-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1px solid var(--border-color);
  background: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  color: var(--text-secondary);
  font-size: 18px;
}

.stop-btn:hover {
  background: #ffebee;
  border-color: #ef9a9a;
  color: #c62828;
  transform: scale(1.05);
}

/* Animations */
.slide-fade-enter-active,
.slide-fade-leave-active {
  transition: all 0.3s ease;
}

.slide-fade-enter-from,
.slide-fade-leave-to {
  transform: translateY(-10px);
  opacity: 0;
}

.list-enter-active,
.list-leave-active {
  transition: all 0.4s ease;
}

.list-enter-from,
.list-leave-to {
  opacity: 0;
  transform: translateX(30px);
}
</style>
